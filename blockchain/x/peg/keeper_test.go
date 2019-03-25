package peg

import (
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Test helpers copied from the bank keeper tests.

type testHelper struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  auth.AccountKeeper
	pk  params.Keeper
	bk  bank.Keeper
}

func setupTestHelper() testHelper {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	fckCapKey := sdk.NewKVStoreKey("fckCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(fckCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	ak := auth.NewAccountKeeper(
		cdc,
		authCapKey,
		pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	bk := bank.NewBaseKeeper(
		ak,
		pk.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, auth.DefaultParams())
	bk.SetSendEnabled(ctx, true) // TODO is this needed?

	return testHelper{cdc: cdc, ctx: ctx, ak: ak, pk: pk, bk: bk}
}
func TestFetchXrpTx(t *testing.T) {
	helper := setupTestHelper()
	keeper := NewKeeper(helper.bk, sdk.NewKVStoreKey("pegStoreKey"), helper.cdc)

	xrpTx, err := keeper.fetchXrpTx("4C3AF3C9200289A0EA970CFE21F698DC6F3BBAEB3CB78E63CA3598A2F7FED5E9")
	if err != nil {
		t.Error(err)
	}
	t.Log(xrpTx)

	// Test passes if there is no error
}
