package auction

import (
	"github.com/stretchr/testify/require"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO can this be less verbose?
func TestForwardAuction_PlaceBid(t *testing.T) {
	seller := sdk.AccAddress([]byte("a_seller"))
	buyer1 := sdk.AccAddress([]byte("buyer1"))
	buyer2 := sdk.AccAddress([]byte("buyer2"))
	end := endTime(10000)
	now := endTime(10)

	type args struct {
		currentBlockHeight endTime
		bidder             sdk.AccAddress
		lot                sdk.Coin
		bid                sdk.Coin
	}
	tests := []struct {
		name            string
		auction         ForwardAuction
		args            args
		expectedOutputs []bankOutput
		expectedInputs  []bankInput
		expectedEndTime endTime
		expectedBidder  sdk.AccAddress
		expectedBid     sdk.Coin
		expectpass      bool
	}{
		{
			"normal",
			ForwardAuction{baseAuction{
				Initiator:  seller,
				Lot:        c("usdx", 100),
				Bidder:     buyer1,
				Bid:        c("xrs", 6),
				EndTime:    end,
				MaxEndTime: end,
			}},
			args{now, buyer2, c("usdx", 100), c("xrs", 10)},
			[]bankOutput{{buyer2, c("xrs", 10)}},
			[]bankInput{{buyer1, c("xrs", 6)}, {seller, c("xrs", 4)}},
			now + bidDuration,
			buyer2,
			c("xrs", 10),
			true,
		},
		{
			"lowBid",
			ForwardAuction{baseAuction{
				Initiator:  seller,
				Lot:        c("usdx", 100),
				Bidder:     buyer1,
				Bid:        c("xrs", 6),
				EndTime:    end,
				MaxEndTime: end,
			}},
			args{now, buyer2, c("usdx", 100), c("xrs", 5)},
			[]bankOutput{},
			[]bankInput{},
			end,
			buyer1,
			c("xrs", 6),
			false,
		},
		{
			"timeout",
			ForwardAuction{baseAuction{
				Initiator:  seller,
				Lot:        c("usdx", 100),
				Bidder:     buyer1,
				Bid:        c("xrs", 6),
				EndTime:    end,
				MaxEndTime: end,
			}},
			args{end + 1, buyer2, c("usdx", 100), c("xrs", 10)},
			[]bankOutput{},
			[]bankInput{},
			end,
			buyer1,
			c("xrs", 6),
			false,
		},
		{
			"hitMaxEndTime",
			ForwardAuction{baseAuction{
				Initiator:  seller,
				Lot:        c("usdx", 100),
				Bidder:     buyer1,
				Bid:        c("xrs", 6),
				EndTime:    end,
				MaxEndTime: end,
			}},
			args{end - 1, buyer2, c("usdx", 100), c("xrs", 10)},
			[]bankOutput{{buyer2, c("xrs", 10)}},
			[]bankInput{{buyer1, c("xrs", 6)}, {seller, c("xrs", 4)}},
			end, // end time should be capped at MaxEndTime
			buyer2,
			c("xrs", 10),
			true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// update auction and return in/outputs
			outputs, inputs, err := tc.auction.PlaceBid(tc.args.currentBlockHeight, tc.args.bidder, tc.args.lot, tc.args.bid)

			// check for err
			if tc.expectpass {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
			// check for correct in/outputs
			require.Equal(t, tc.expectedOutputs, outputs)
			require.Equal(t, tc.expectedInputs, inputs)
			// check for correct endTime, bidder, bid
			require.Equal(t, tc.expectedEndTime, tc.auction.EndTime)
			require.Equal(t, tc.expectedBidder, tc.auction.Bidder)
			require.Equal(t, tc.expectedBid, tc.auction.Bid)
		})
	}
}

// defined to avoid cluttering test cases with long function name
func c(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
}

// func TestReverseAuction_PlaceBid(t *testing.T) {
// 	type fields struct {
// 		baseAuction baseAuction
// 	}
// 	type args struct {
// 		currentBlockHeight endTime
// 		bidder             sdk.AccAddress
// 		lot                sdk.Coin
// 		bid                sdk.Coin
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   []bankOutput
// 		want1  []bankInput
// 		want2  sdk.Error
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := ReverseAuction{
// 				baseAuction: tt.fields.baseAuction,
// 			}
// 			got, got1, got2 := a.PlaceBid(tt.args.currentBlockHeight, tt.args.bidder, tt.args.lot, tt.args.bid)
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ReverseAuction.PlaceBid() got = %v, want %v", got, tt.want)
// 			}
// 			if !reflect.DeepEqual(got1, tt.want1) {
// 				t.Errorf("ReverseAuction.PlaceBid() got1 = %v, want %v", got1, tt.want1)
// 			}
// 			if !reflect.DeepEqual(got2, tt.want2) {
// 				t.Errorf("ReverseAuction.PlaceBid() got2 = %v, want %v", got2, tt.want2)
// 			}
// 		})
// 	}
// }

// func TestForwardReverseAuction_PlaceBid(t *testing.T) {
// 	type fields struct {
// 		baseAuction baseAuction
// 		MaxBid      sdk.Coin
// 		OtherPerson sdk.AccAddress
// 	}
// 	type args struct {
// 		currentBlockHeight endTime
// 		bidder             sdk.AccAddress
// 		lot                sdk.Coin
// 		bid                sdk.Coin
// 	}
// 	tests := []struct {
// 		name        string
// 		fields      fields
// 		args        args
// 		wantOutputs []bankOutput
// 		wantInputs  []bankInput
// 		wantErr     sdk.Error
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := ForwardReverseAuction{
// 				baseAuction: tt.fields.baseAuction,
// 				MaxBid:      tt.fields.MaxBid,
// 				OtherPerson: tt.fields.OtherPerson,
// 			}
// 			gotOutputs, gotInputs, gotErr := a.PlaceBid(tt.args.currentBlockHeight, tt.args.bidder, tt.args.lot, tt.args.bid)
// 			if !reflect.DeepEqual(gotOutputs, tt.wantOutputs) {
// 				t.Errorf("ForwardReverseAuction.PlaceBid() gotOutputs = %v, want %v", gotOutputs, tt.wantOutputs)
// 			}
// 			if !reflect.DeepEqual(gotInputs, tt.wantInputs) {
// 				t.Errorf("ForwardReverseAuction.PlaceBid() gotInputs = %v, want %v", gotInputs, tt.wantInputs)
// 			}
// 			if !reflect.DeepEqual(gotErr, tt.wantErr) {
// 				t.Errorf("ForwardReverseAuction.PlaceBid() gotErr = %v, want %v", gotErr, tt.wantErr)
// 			}
// 		})
// 	}
// }
