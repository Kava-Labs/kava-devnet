package app

import (
	"encoding/json"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/liquidator"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	appName = "usdx"
)

// UsdxApp - Extended ABCI application
type UsdxApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain             *sdk.KVStoreKey
	keyAccount          *sdk.KVStoreKey
	keyFeeCollection    *sdk.KVStoreKey
	keyParams           *sdk.KVStoreKey
	tkeyParams          *sdk.TransientStoreKey
	keyPricefeed        *sdk.KVStoreKey
	keyAuction          *sdk.KVStoreKey
	keyCdp              *sdk.KVStoreKey
	keyLiquidator       *sdk.KVStoreKey
	accountKeeper       auth.AccountKeeper
	auctionKeeper       auction.Keeper
	bankKeeper          bank.Keeper
	cdpKeeper           cdp.Keeper
	liquidatorKeeper    liquidator.Keeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	paramsKeeper        params.Keeper
	pricefeedKeeper     pricefeed.Keeper
}

// NewUsdxApp is a constructor function for usdxApp
func NewUsdxApp(logger log.Logger, db dbm.DB) *UsdxApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	// Here you initialize your application with the store keys it requires
	var app = &UsdxApp{
		BaseApp: bApp,
		cdc:     cdc,

		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey("acc"),
		keyFeeCollection: sdk.NewKVStoreKey("fee_collection"),
		keyParams:        sdk.NewKVStoreKey("params"),
		tkeyParams:       sdk.NewTransientStoreKey("transient_params"),
		keyPricefeed:     sdk.NewKVStoreKey("pricefeed"),
		keyAuction:       sdk.NewKVStoreKey("auction"),
		keyCdp:           sdk.NewKVStoreKey("cdp"),
		keyLiquidator:    sdk.NewKVStoreKey("liquidator"),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)
	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	// The FeeCollectionKeeper collects transaction fees and renders them to the fee distribution module
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)

	// pricefeedKeeper handles postPrice transactions posted by oracles
	app.pricefeedKeeper = pricefeed.NewKeeper(app.keyPricefeed, app.cdc, pricefeed.DefaultCodespace)

	app.cdpKeeper = cdp.NewKeeper(
		app.cdc,
		app.keyCdp,
		app.paramsKeeper.Subspace("cdp"),
		app.pricefeedKeeper,
		app.bankKeeper,
	)
	app.auctionKeeper = auction.NewKeeper(
		app.cdc,
		app.cdpKeeper, // CDP keeper standing in for bank
		app.keyAuction,
	)
	app.liquidatorKeeper = liquidator.NewKeeper(
		app.cdc,
		app.keyLiquidator,
		app.cdpKeeper,
		app.auctionKeeper,
		app.cdpKeeper, // CDP keeper standing in for bank
	)

	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))

	// The app.Router is the main transaction router where each module registers its routes
	// Register the bank and nameservice routes here
	app.Router().
		AddRoute("bank", bank.NewHandler(app.bankKeeper)).
		AddRoute("pricefeed", pricefeed.NewHandler(app.pricefeedKeeper)).
		AddRoute("auction", auction.NewHandler(app.auctionKeeper)).
		AddRoute("cdp", cdp.NewHandler(app.cdpKeeper)).
		AddRoute("liquidator", liquidator.NewHandler(app.liquidatorKeeper))

	// The app.QueryRouter is the main query router where each module registers its routes
	app.QueryRouter().
		AddRoute(auth.QuerierRoute, auth.NewQuerier(app.accountKeeper)).
		AddRoute("pricefeed", pricefeed.NewQuerier(app.pricefeedKeeper))

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.initChainer)

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keyFeeCollection,
		app.keyParams,
		app.tkeyParams,
		app.keyPricefeed,
		app.keyAuction,
		app.keyCdp,
		app.keyLiquidator,
	)

	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState struct {
	AuthData      auth.GenesisState      `json:"auth"`
	BankData      bank.GenesisState      `json:"bank"`
	PricefeedData pricefeed.GenesisState `json:"pricfeed"`
	CdpData       cdp.GenesisState       `json:"cdp"`
	Accounts      []*auth.BaseAccount    `json:"accounts"`
}

func (app *UsdxApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
	}

	for _, acc := range genesisState.Accounts {
		acc.AccountNumber = app.accountKeeper.GetNextAccountNumber(ctx)
		app.accountKeeper.SetAccount(ctx, acc)
	}

	auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthData)
	bank.InitGenesis(ctx, app.bankKeeper, genesisState.BankData)
	pricefeed.InitGenesis(ctx, app.pricefeedKeeper, genesisState.PricefeedData)
	cdp.InitGenesis(ctx, app.cdpKeeper, cdp.DefaultGenesisState())
	return abci.ResponseInitChain{}
}

// ExportAppStateAndValidators does the things
func (app *UsdxApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
	ctx := app.NewContext(true, abci.Header{})
	accounts := []*auth.BaseAccount{}

	appendAccountsFn := func(acc auth.Account) bool {
		account := &auth.BaseAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}

		accounts = append(accounts, account)
		return false
	}

	app.accountKeeper.IterateAccounts(ctx, appendAccountsFn)

	genState := GenesisState{
		Accounts: accounts,
		AuthData: auth.DefaultGenesisState(),
		BankData: bank.DefaultGenesisState(),
	}

	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	return appState, validators, err
}

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	pricefeed.RegisterCodec(cdc)
	staking.RegisterCodec(cdc) // TODO is this meant to be here? There's no staking module in this app.
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// SetAddressPrefixes sets the bech32 address prefixes globally for the sdk module.
func SetAddressPrefixes() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("usdx", "usdx"+"pub")
	config.SetBech32PrefixForValidator("usdx"+"val"+"oper", "usdx"+"val"+"oper"+"pub")
	config.SetBech32PrefixForConsensusNode("usdx"+"val"+"cons", "usdx"+"val"+"cons"+"pub")
	config.Seal()
}
