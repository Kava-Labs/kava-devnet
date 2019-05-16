package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/kava-labs/usdx/blockchain/x/cdp"

	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type modifyCdpReq struct {
	BaseReq         rest.BaseReq `json:"base_req"`
	OwnerAddr       string       `json:"owner_addr"`
	CollateralType  string       `json:"collateral_type"`
	CollateralDelta string       `json:"collateral_delta"`
	StableDelta     string       `json:"stable_delta"`
}

const (
	restOwner           = "owner"
	restCollateralType  = "collateral"
	restCollateralDelta = "collateral_delta"
	restStableDelta     = "stable_delta"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(fmt.Sprintf("/cdp/getcdpinfo/{%s}/{%s}", restOwner, restCollateralType), queryCdpHandlerFn(cdc, cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/cdp/modify/{%s}/{%s}/{%s}/{%s}", restOwner, restCollateralType, restCollateralDelta, restStableDelta), modifyCdpHandlerFn(cdc, cliCtx)).Methods("PUT")
}

func queryCdpHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bechOwnerAddr := vars[restOwner]
		strCollateralType := vars[restCollateralType]

		ownerAddr, err := sdk.AccAddressFromBech32(bechOwnerAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(strCollateralType) == 0 {
			err := errors.New("collateralType required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/getcdpinfo/%s/%s", ownerAddr, strCollateralType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)

	}
}

func modifyCdpHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req modifyCdpReq
		vars := mux.Vars(r)
		bechOwnerAddr := vars[restOwner]
		strCollateralType := vars[restCollateralType]
		strCollateralDelta := vars[restCollateralDelta]
		strStableDelta := vars[restStableDelta]

		ownerAddr, err := sdk.AccAddressFromBech32(bechOwnerAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(strCollateralType) == 0 {
			err := errors.New("collateralType required but not specified")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		collateralDelta, ok := sdk.NewIntFromString(strCollateralDelta)

		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("invalid collateral amount").Error())
			return
		}

		stableDelta, ok := sdk.NewIntFromString(strStableDelta)

		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("invalid stable amount").Error())
			return
		}

		msg := cdp.NewMsgCreateOrModifyCDP(ownerAddr, strCollateralType, collateralDelta, stableDelta)
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
