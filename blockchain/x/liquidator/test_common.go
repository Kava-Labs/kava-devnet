package liquidator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

func setUpMockAppWithoutGenesis() (*mock.App, Keeper) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	bank.RegisterCodec(mapp.Cdc)
	pricefeed.RegisterCodec(mapp.Cdc)
	auction.RegisterCodec(mapp.Cdc)
	cdp.RegisterCodec(mapp.Cdc)
	RegisterCodec(mapp.Cdc)

	// Create keepers
	keyPriceFeed := sdk.NewKVStoreKey(pricefeed.StoreKey)
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyAuction := sdk.NewKVStoreKey("auction")
	keyLiquidator := sdk.NewKVStoreKey("liquidator")

	priceFeedKeeper := pricefeed.NewKeeper(keyPriceFeed, mapp.Cdc, pricefeed.DefaultCodespace)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	cdpKeeper := cdp.NewKeeper(mapp.Cdc, keyCDP, mapp.ParamsKeeper.Subspace("cdpSubspace"), priceFeedKeeper, bankKeeper)
	auctionKeeper := auction.NewKeeper(mapp.Cdc, cdpKeeper, keyAuction)                         // Note: cdp keeper stands in for bank keeper
	liquidatorKeeper := NewKeeper(mapp.Cdc, keyLiquidator, cdpKeeper, auctionKeeper, cdpKeeper) // Note: cdp keeper stands in for bank keeper

	// Register routes
	mapp.Router().
		AddRoute("bank", bank.NewHandler(bankKeeper)).
		AddRoute("pricefeed", pricefeed.NewHandler(priceFeedKeeper)).
		AddRoute("cdp", cdp.NewHandler(cdpKeeper)).
		AddRoute("liquidator", NewHandler(liquidatorKeeper)).
		AddRoute("auction", auction.NewHandler(auctionKeeper))

	mapp.SetInitChainer(
		func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
			res := mapp.InitChainer(ctx, req)
			bank.InitGenesis(ctx, bankKeeper, bank.DefaultGenesisState())
			pfGenState := pricefeed.GenesisState{
				Assets: []pricefeed.Asset{
					{AssetCode: "btc", Description: ""},
					{AssetCode: "xrp", Description: ""},
				},
			}
			pricefeed.InitGenesis(ctx, priceFeedKeeper, pfGenState)
			cdp.InitGenesis(ctx, cdpKeeper, cdp.DefaultGenesisState())
			return res
		},
	)
	mapp.SetEndBlocker(
		func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			auctionTags := auction.EndBlocker(ctx, auctionKeeper)
			pricefeedTags := pricefeed.EndBlocker(ctx, priceFeedKeeper)
			return abci.ResponseEndBlock{
				Tags: append(auctionTags, pricefeedTags...),
			}
		},
	)

	// Mount and load the stores
	err := mapp.CompleteSetup(keyPriceFeed, keyCDP, keyAuction, keyLiquidator)
	if err != nil {
		panic(err)
	}

	return mapp, liquidatorKeeper
}

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

type keepers struct {
	paramsKeeper     params.Keeper
	accountKeeper    auth.AccountKeeper
	bankKeeper       bank.Keeper
	pricefeedKeeper  pricefeed.Keeper
	auctionKeeper    auction.Keeper
	cdpKeeper        cdp.Keeper
	liquidatorKeeper Keeper
}

func setupTestKeepers() (sdk.Context, keepers) {

	// Setup in memory database
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyPriceFeed := sdk.NewKVStoreKey(pricefeed.StoreKey)
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyAuction := sdk.NewKVStoreKey("auction")
	keyLiquidator := sdk.NewKVStoreKey("liquidator")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyPriceFeed, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyCDP, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAuction, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyLiquidator, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// Create Codec
	cdc := MakeTestCodec()

	// Create Keepers
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams)
	accountKeeper := auth.NewAccountKeeper(
		cdc,
		keyAcc,
		paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)
	bankKeeper := bank.NewBaseKeeper(
		accountKeeper,
		paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)
	pricefeedKeeper := pricefeed.NewKeeper(keyPriceFeed, cdc, pricefeed.DefaultCodespace)
	cdpKeeper := cdp.NewKeeper(
		cdc,
		keyCDP,
		paramsKeeper.Subspace("cdpSubspace"),
		pricefeedKeeper,
		bankKeeper,
	)
	auctionKeeper := auction.NewKeeper(cdc, cdpKeeper, keyAuction)                         // Note: cdp keeper stands in for bank keeper
	liquidatorKeeper := NewKeeper(cdc, keyLiquidator, cdpKeeper, auctionKeeper, cdpKeeper) // Note: cdp keeper stands in for bank keeper

	// Create context
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "testchain"}, false, log.NewNopLogger())

	// Setup all the state within the keepers (including genesis)
	// TODO move out of this function
	// auth genesis - requires fee keeper
	cdp.InitGenesis(ctx, cdpKeeper, cdp.DefaultGenesisState())

	return ctx, keepers{
		paramsKeeper,
		accountKeeper,
		bankKeeper,
		pricefeedKeeper,
		auctionKeeper,
		cdpKeeper,
		liquidatorKeeper,
	}
}

func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	pricefeed.RegisterCodec(cdc)
	auction.RegisterCodec(cdc)
	cdp.RegisterCodec(cdc)
	RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
