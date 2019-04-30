package auction

import (
	"testing"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgPlaceBid_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("someName"))
	tests := []struct {
		name   string
		msg MsgPlaceBid
		expectPass   bool
	}{
		{"normal", MsgPlaceBid{0, addr, sdk.NewInt64Coin("usdx", 10), sdk.NewInt64Coin("xrs", 20)}, true},
		{"emptyAddr", MsgPlaceBid{0, sdk.AccAddress{}, sdk.NewInt64Coin("usdx", 10), sdk.NewInt64Coin("xrs", 20)}, false},
		{"negativeBid", MsgPlaceBid{0, addr, sdk.Coin{"usdx", sdk.NewInt(-10)}, sdk.NewInt64Coin("xrs", 20)}, false},
		{"negativeLot", MsgPlaceBid{0, addr, sdk.NewInt64Coin("usdx", 10), sdk.Coin{"xrs", sdk.NewInt(-20)}}, false},
		{"zerocoins", MsgPlaceBid{0, addr, sdk.NewInt64Coin("usdx", 0), sdk.NewInt64Coin("xrs", 0)}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPass {
				require.Nil(t, tc.msg.ValidateBasic())
			} else {
				require.NotNil(t, tc.msg.ValidateBasic())
			}
		})
	}
}