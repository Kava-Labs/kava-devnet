package app

import (
	"io"
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/liquidator"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	appName = "usdx"
)

var (
	// DefaultCLIHome default home directories for usdxcli
	DefaultCLIHome = os.ExpandEnv("$HOME/.usdxcli")

	// DefaultNodeHome default home directories for usdxd
	DefaultNodeHome = os.ExpandEnv("$HOME/.usdxd")

	// ModuleBasics The ModuleBasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics sdk.ModuleBasicManager
)

func init() {
	ModuleBasics = sdk.NewModuleBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.AppModuleBasic{},
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		auction.AppModuleBasic{},
		cdp.AppModuleBasic{},
		liquidator.AppModuleBasic{},
		pricefeed.AppModule{},
	)
}

// MakeCodec custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
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

// UsdxApp - Extended ABCI application
type UsdxApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyStaking       *sdk.KVStoreKey
	tkeyStaking      *sdk.TransientStoreKey
	keySlashing      *sdk.KVStoreKey
	keyMint          *sdk.KVStoreKey
	keyDistr         *sdk.KVStoreKey
	tkeyDistr        *sdk.TransientStoreKey
	keyGov           *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey
	keyPricefeed     *sdk.KVStoreKey
	keyAuction       *sdk.KVStoreKey
	keyCdp           *sdk.KVStoreKey
	keyLiquidator    *sdk.KVStoreKey

	// keepers from cosmos-sdk
	accountKeeper       auth.AccountKeeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	bankKeeper          bank.Keeper
	stakingKeeper       staking.Keeper
	slashingKeeper      slashing.Keeper
	mintKeeper          mint.Keeper
	distrKeeper         distr.Keeper
	govKeeper           gov.Keeper
	crisisKeeper        crisis.Keeper
	paramsKeeper        params.Keeper

	// app specific keepers
	auctionKeeper    auction.Keeper
	cdpKeeper        cdp.Keeper
	liquidatorKeeper liquidator.Keeper
	pricefeedKeeper  pricefeed.Keeper

	// the module manager
	mm *sdk.ModuleManager
}

// NewUsdxApp is a constructor function for usdxApp
func NewUsdxApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp)) *UsdxApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	// Here you initialize your application with the store keys it requires
	var app = &UsdxApp{
		BaseApp:          bApp,
		cdc:              cdc,
		invCheckPeriod:   invCheckPeriod,
		keyMain:          sdk.NewKVStoreKey(bam.MainStoreKey),
		keyAccount:       sdk.NewKVStoreKey(auth.StoreKey),
		keyStaking:       sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking:      sdk.NewTransientStoreKey(staking.TStoreKey),
		keyMint:          sdk.NewKVStoreKey(mint.StoreKey),
		keyDistr:         sdk.NewKVStoreKey(distr.StoreKey),
		tkeyDistr:        sdk.NewTransientStoreKey(distr.TStoreKey),
		keySlashing:      sdk.NewKVStoreKey(slashing.StoreKey),
		keyGov:           sdk.NewKVStoreKey(gov.StoreKey),
		keyFeeCollection: sdk.NewKVStoreKey(auth.FeeStoreKey),
		keyParams:        sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:       sdk.NewTransientStoreKey(params.TStoreKey),
		keyPricefeed:     sdk.NewKVStoreKey("pricefeed"),
		keyAuction:       sdk.NewKVStoreKey("auction"),
		keyCdp:           sdk.NewKVStoreKey("cdp"),
		keyLiquidator:    sdk.NewKVStoreKey("liquidator"),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams, params.DefaultCodespace)
	// params subspaces for each module
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace)
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	cdpSubspace := app.paramsKeeper.Subspace("cdp")
	liquidatorSubspace := app.paramsKeeper.Subspace("liquidator")

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(app.cdc, app.keyAccount, authSubspace, auth.ProtoBaseAccount)
	app.bankKeeper = bank.NewBaseKeeper(app.accountKeeper, bankSubspace, bank.DefaultCodespace)
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)
	stakingKeeper := staking.NewKeeper(app.cdc, app.keyStaking, app.tkeyStaking, app.bankKeeper,
		stakingSubspace, staking.DefaultCodespace)
	app.mintKeeper = mint.NewKeeper(app.cdc, app.keyMint, mintSubspace, &stakingKeeper, app.feeCollectionKeeper)
	app.distrKeeper = distr.NewKeeper(app.cdc, app.keyDistr, distrSubspace, app.bankKeeper, &stakingKeeper,
		app.feeCollectionKeeper, distr.DefaultCodespace)
	app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, &stakingKeeper,
		slashingSubspace, slashing.DefaultCodespace)
	app.crisisKeeper = crisis.NewKeeper(crisisSubspace, invCheckPeriod, app.distrKeeper,
		app.bankKeeper, app.feeCollectionKeeper)

	app.pricefeedKeeper = pricefeed.NewKeeper(app.keyPricefeed, app.cdc, pricefeed.DefaultCodespace)
	app.cdpKeeper = cdp.NewKeeper(
		app.cdc,
		app.keyCdp,
		cdpSubspace,
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
		liquidatorSubspace,
		app.cdpKeeper,
		app.auctionKeeper,
		app.cdpKeeper, // CDP keeper standing in for bank
	)

	// register the proposal types
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper))
	app.govKeeper = gov.NewKeeper(app.cdc, app.keyGov, app.paramsKeeper, govSubspace,
		app.bankKeeper, &stakingKeeper, gov.DefaultCodespace, govRouter)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	app.mm = sdk.NewModuleManager(
		genaccounts.NewAppModule(app.accountKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper, app.feeCollectionKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(app.crisisKeeper, app.Logger()),
		distr.NewAppModule(app.distrKeeper),
		gov.NewAppModule(app.govKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),

		auction.NewAppModule(app.auctionKeeper),
		cdp.NewAppModule(app.cdpKeeper),
		liquidator.NewAppModule(app.liquidatorKeeper),
		pricefeed.NewAppModule(app.pricefeedKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName, slashing.ModuleName)

	// During the endblock, governance proposals expire, staking rewards are distributed, and the pricefeed updates
	app.mm.SetOrderEndBlockers(gov.ModuleName, staking.ModuleName, pricefeed.ModuleName)

	// genutils must occur after staking so that pools are properly
	// initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(genaccounts.ModuleName, distr.ModuleName,
		staking.ModuleName, auth.ModuleName, bank.ModuleName, slashing.ModuleName,
		gov.ModuleName, mint.ModuleName, crisis.ModuleName, genutil.ModuleName,
		auction.ModuleName, cdp.ModuleName, liquidator.ModuleName, pricefeed.ModuleName)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keyStaking,
		app.keyMint,
		app.keyDistr,
		app.keySlashing,
		app.keyGov,
		app.keyFeeCollection,
		app.keyParams,
		app.tkeyParams,
		app.tkeyStaking,
		app.tkeyDistr,
		app.keyPricefeed,
		app.keyAuction,
		app.keyCdp,
		app.keyLiquidator,
	)

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper, auth.DefaultSigVerificationGasConsumer))
	// Set the function to be run at the end of every block
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keyMain)
		if err != nil {
			cmn.Exit(err.Error())
		}
	}
	return app
}

// BeginBlocker application updates every begin block
func (app *UsdxApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *UsdxApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *UsdxApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, genesisState)
}

// LoadHeight load a particular height
func (app *UsdxApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keyMain)
}
