package cdp

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/usdx/blockchain/x/cdp/mockpricefeed"
)

// How could one reduce the number of params in the test cases. Create a table driven test for each of the 4 add/withdraw collateral/debt?
// Keeper test just needs a mock keeper, not a whole mock app. Stake does this. It creates an in memory db, and creates all the sub keepers needed. Somewhat verbose though.
// Problem in that getMockApp returns randomly generated addresses. But I want to use an address in the test cases.

// TODO sort the coins in the test to avoid error in comparing coins. Use sdk.NewCoins in v0.34

func TestKeeper_ModifyCDP(t *testing.T) {
	ownerAddr, _, _ := generateAccAddress()

	type state struct { // TODO this allows invalid state to be set up, should it?
		CDP             CDP
		OwnerCoins      sdk.Coins
		GlobalDebt      sdk.Int
		CollateralState CollateralState
	}
	type args struct {
		owner              sdk.AccAddress
		collateralDenom    string
		changeInCollateral sdk.Int
		changeInDebt       sdk.Int
	}

	tests := []struct {
		name       string
		priorState state
		price      string
		// also missing moduleParams
		args          args
		expectPass    bool
		expectedState state
	}{
		{
			"addCollateralAndDecreaseDebt",
			state{CDP{ownerAddr, "xrp", i(100), i(2)}, sdk.Coins{c("usdx", 2), c("xrp", 10)}, i(2), CollateralState{"xrp", i(2)}},
			"10.345",
			args{ownerAddr, "xrp", i(10), i(-1)},
			true,
			state{CDP{ownerAddr, "xrp", i(110), i(1)}, sdk.Coins{c("usdx", 1) /* 0 xrp */}, i(1), CollateralState{"xrp", i(1)}},
		},
		{
			"removeTooMuchCollateral",
			state{CDP{ownerAddr, "xrp", i(1000), i(200)}, sdk.Coins{c(StableDenom, 10), c("xrp", 10)}, i(200), CollateralState{"xrp", i(200)}},
			"1.00",
			args{ownerAddr, "xrp", i(-601), i(0)},
			false,
			state{CDP{ownerAddr, "xrp", i(1000), i(200)}, sdk.Coins{c(StableDenom, 10), c("xrp", 10)}, i(200), CollateralState{"xrp", i(200)}},
		},
		{
			"withdrawTooMuchStableCoin",
			state{CDP{ownerAddr, "xrp", i(1000), i(200)}, sdk.Coins{c(StableDenom, 10), c("xrp", 10)}, i(200), CollateralState{"xrp", i(200)}},
			"1.00",
			args{ownerAddr, "xrp", i(0), i(301)},
			false,
			state{CDP{ownerAddr, "xrp", i(1000), i(200)}, sdk.Coins{c(StableDenom, 10), c("xrp", 10)}, i(200), CollateralState{"xrp", i(200)}},
		},
		{
			"createCDPAndWithdrawStable",
			state{CDP{}, sdk.Coins{c(StableDenom, 10), c("xrp", 10)}, i(0), CollateralState{"xrp", i(0)}},
			"1.00",
			args{ownerAddr, "xrp", i(5), i(2)},
			true,
			state{CDP{ownerAddr, "xrp", i(5), i(2)}, sdk.Coins{c(StableDenom, 12), c("xrp", 5)}, i(2), CollateralState{"xrp", i(2)}},
		},
		{
			"emptyCDP",
			state{CDP{ownerAddr, "xrp", i(1000), i(200)}, sdk.Coins{c(StableDenom, 201), c("xrp", 10)}, i(200), CollateralState{"xrp", i(200)}},
			"1.00",
			args{ownerAddr, "xrp", i(-1000), i(-200)},
			true,
			state{CDP{}, sdk.Coins{c(StableDenom, 1), c("xrp", 1010)}, i(0), CollateralState{"xrp", i(0)}},
		},
		{
			"invalidCollateralType",
			state{CDP{}, sdk.Coins{c("shitcoin", 5000000)}, i(0), CollateralState{}},
			"0.000001",
			args{ownerAddr, "shitcoin", i(5000000), i(1)}, // ratio of 5:1
			false,
			state{CDP{}, sdk.Coins{c("shitcoin", 5000000)}, i(0), CollateralState{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup keeper
			mapp, keeper := setUpMockAppWithoutGenAccounts()
			// initialize cdp owner account with coins
			genAcc := auth.BaseAccount{
				Address: ownerAddr,
				Coins:   tc.priorState.OwnerCoins,
			}
			mock.SetGenesis(mapp, []auth.Account{&genAcc})
			// set the pricefeed keeper to return the specified price
			keeper.pricefeed = mockpricefeed.NewKeeper(tc.price)
			// create a new context
			mapp.BeginBlock(abci.RequestBeginBlock{})
			ctx := mapp.BaseApp.NewContext(false, abci.Header{})
			// setup store state
			if tc.priorState.CDP.CollateralDenom != "" { // check if the prior CDP should be created or not (see if an empty one was specified)
				keeper.setCDP(ctx, tc.priorState.CDP)
			}
			keeper.setGlobalDebt(ctx, tc.priorState.GlobalDebt)
			if tc.priorState.CollateralState.Denom != "" {
				keeper.setCollateralState(ctx, tc.priorState.CollateralState)
			}
			// TODO close/commit block?

			// call func under test
			err := keeper.ModifyCDP(ctx, tc.args.owner, tc.args.collateralDenom, tc.args.changeInCollateral, tc.args.changeInDebt)
			mapp.EndBlock(abci.RequestEndBlock{})
			mapp.Commit()

			// check for err
			if tc.expectPass {
				require.Nil(t, err, fmt.Sprint(err))
			} else {
				require.NotNil(t, err)
			}
			// get new state for verification
			actualCDP, found := keeper.GetCDP(ctx, tc.args.owner, tc.args.collateralDenom)
			actualGDebt := keeper.GetGlobalDebt(ctx)
			actualCstate, _ := keeper.GetCollateralState(ctx, tc.args.collateralDenom)
			// check state
			require.Equal(t, tc.expectedState.CDP, actualCDP)
			if tc.expectedState.CDP.CollateralDenom == "" { // if the expected CDP is blank, then expect the CDP to have been deleted (hence not found)
				require.False(t, found)
			} else {
				require.True(t, found)
			}
			require.Equal(t, tc.expectedState.GlobalDebt, actualGDebt)
			require.Equal(t, tc.expectedState.CollateralState, actualCstate)
			// check owner balance
			mock.CheckBalance(t, mapp, ownerAddr, tc.expectedState.OwnerCoins)
		})
	}
}

func TestKeeper_GetSetDeleteCDP(t *testing.T) {
	// setup keeper, create CDP
	mapp, keeper := setUpMockAppWithoutGenAccounts()
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	testAddr, _, _ := generateAccAddress() // TODO is this bad because it is not deterministic?
	// TODO should this set the genesis state ? Same for tests below
	cdp := CDP{testAddr, "xrp", sdk.NewInt(412), sdk.NewInt(56)}

	// write and read from store
	keeper.setCDP(ctx, cdp)
	readCDP, found := keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)

	// check before and after match
	require.True(t, found)
	require.Equal(t, cdp, readCDP)

	// delete auction
	keeper.deleteCDP(ctx, cdp)

	// check auction does not exist
	_, found = keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
}
func TestKeeper_GetSetGDebt(t *testing.T) {
	// setup keeper, create GDebt
	mapp, keeper := setUpMockAppWithoutGenAccounts()
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	gDebt := sdk.NewInt(4120000)

	// write and read from store
	keeper.setGlobalDebt(ctx, gDebt)
	readGDebt := keeper.GetGlobalDebt(ctx)

	// check before and after match
	require.Equal(t, gDebt, readGDebt)
}

func TestKeeper_GetSetCollateralState(t *testing.T) {
	// setup keeper, create CState
	mapp, keeper := setUpMockAppWithoutGenAccounts()
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	cState := CollateralState{"xrp", sdk.NewInt(15400)}

	// write and read from store
	keeper.setCollateralState(ctx, cState)
	readCState, found := keeper.GetCollateralState(ctx, cState.Denom)

	// check before and after match
	require.Equal(t, cState, readCState)
	require.True(t, found)
}

func setUpMockAppWithoutGenAccounts() (*mock.App, Keeper) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	//RegisterCodec(mapp.Cdc) // Add back once messages are written

	// Create keepers
	keyCDP := sdk.NewKVStoreKey("cdp")
	priceFeedKeeper := mockpricefeed.Keeper{}
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	cdpKeeper := NewKeeper(mapp.Cdc, keyCDP, mapp.ParamsKeeper.Subspace("cdpSubspace"), priceFeedKeeper, bankKeeper)

	// Register routes
	//mapp.Router().AddRoute("cdp", NewHandler(cdpKeeper))

	mapp.SetInitChainer(
		func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
			res := mapp.InitChainer(ctx, req)
			InitGenesis(ctx, cdpKeeper, DefaultGenesisState()) // Create a default genesis state, then set the keeper store to it
			return res
		},
	)

	// Mount and load the stores
	err := mapp.CompleteSetup(keyCDP)
	if err != nil {
		panic("mock app setup failed")
	}

	return mapp, cdpKeeper
}

func generateAccAddress() (sdk.AccAddress, crypto.PubKey, crypto.PrivKey) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())
	return addr, pubKey, privKey
}

// defined to avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
