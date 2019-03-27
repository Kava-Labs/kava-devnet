package peg

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
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

	xrpTx, err := keeper.fetchXrpTransactionData("4C3AF3C9200289A0EA970CFE21F698DC6F3BBAEB3CB78E63CA3598A2F7FED5E9")
	if err != nil {
		t.Error(err)
	}
	t.Log(xrpTx)

	_, err = keeper.fetchXrpTransactionData("BAD")
	if err == nil {
		t.Error("Invalid Xrp Transaction failed to error")
	}
	// Tests pass if there is no errors
}

// example cosmos address:
// cosmos18uymlc5uelgjx5kr2eztacdzyy5jvjwf6nnw85
// above address converted to raw bytes then hex encoded:
// 636f736d6f73313875796d6c633575656c676a78356b7232657a746163647a7979356a766a7766366e6e773835
func TestMintPxrp(t *testing.T) {
	helper := setupTestHelper()
	keeper := NewKeeper(helper.bk, sdk.NewKVStoreKey("pegStoreKey"), helper.cdc)
	destAddr, err := sdk.AccAddressFromBech32("cosmos18uymlc5uelgjx5kr2eztacdzyy5jvjwf6nnw85")
	if err != nil {
		t.Error(err)
	}
	amount := 100
	xrpTx := XrpTx{}
	xrpTx.Transaction.Tx.Amount = strconv.Itoa(amount) // "Integer to Ascii"
	memo := Memo{}
	memo.Memo.MemoData = "636f736d6f73313875796d6c633575656c676a78356b7232657a746163647a7979356a766a7766366e6e773835"
	xrpTx.Transaction.Tx.Memos = []Memo{memo}

	tags, err := keeper.mintPxrp(helper.ctx, xrpTx)
	if err != nil {
		t.Error(err)
	}

	coins := keeper.coinKeeper.GetCoins(helper.ctx, destAddr)
	expectedCoins := sdk.Coins{sdk.NewCoin("pxrp", sdk.NewInt(int64(amount)))}
	if !coins.IsEqual(expectedCoins) {
		t.Errorf("Incorrect amount of PXRP minted. Expected %s, got %s", expectedCoins, coins)
	}

	expectedAddress := "cosmos18uymlc5uelgjx5kr2eztacdzyy5jvjwf6nnw85"
	receivedAddress := sdk.TagsToStringTags(tags)[0].Value
	if receivedAddress != expectedAddress {
		t.Errorf("Incorrect receiving address: Expected: %s Got: %s", expectedAddress, receivedAddress)
	}
}
