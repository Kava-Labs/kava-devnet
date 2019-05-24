package rest

import (
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

/*
API Design:

Currently CDPs do not have IDs so standard REST uri conventions (ie GET /cdps/{cdp-id}) don't work too well.

Get one or more cdps
	GET /cdps?collateralDenom={denom}&owner={address}&underCollateralizedAt={price}
Modify a CDP (idempotent). Create is not separated out because conceptually all CDPs already exist (just with zero collateral and debt). // TODO is making this idempotent actually useful?
	PUT /cdps
Get the module params, including authorized collateral denoms.
	GET /params
*/

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc("/cdps", getCdpsHandlerFn(cdc, cliCtx)).Methods("GET")
	r.HandleFunc("/cdps", modifyCdpHandlerFn(cdc, cliCtx)).Methods("PUT")
	r.HandleFunc("/cdps/params", getParamsHandlerFn(cdc, cliCtx)).Methods("GET")
}

const (
	RestOwner                 = "owner"
	RestCollateralDenom       = "collateralDenom"
	RestUnderCollateralizedAt = "underCollateralizedAt"
)

func getCdpsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get parameters from the URL
		ownerBech32 := r.URL.Query().Get(RestOwner)
		collateralDenom := r.URL.Query().Get(RestCollateralDenom)
		priceString := r.URL.Query().Get(RestUnderCollateralizedAt)

		// Construct querier params
		querierParams := cdp.QueryCdpsParams{}

		if len(ownerBech32) != 0 {
			owner, err := sdk.AccAddressFromBech32(ownerBech32)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			querierParams.Owner = owner
		}

		if len(collateralDenom) != 0 {
			// TODO validate denom
			querierParams.CollateralDenom = collateralDenom
		}

		if len(priceString) != 0 {
			price, err := sdk.NewDecFromStr(priceString)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			querierParams.UnderCollateralizedAt = price
		}

		querierParamsBz, err := cdc.MarshalJSON(querierParams)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Get the CDPs
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", cdp.QueryGetCdps), querierParamsBz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		// Return the CDPs
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)

	}
}

type ModifyCdpRequestBody struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Cdp     cdp.CDP      `json:"cdp"`
}

func modifyCdpHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode PUT request body
		var requestBody ModifyCdpRequestBody
		if !rest.ReadRESTReq(w, r, cdc, &requestBody) {
			return
		}
		requestBody.BaseReq = requestBody.BaseReq.Sanitize()
		if !requestBody.BaseReq.ValidateBasic(w) {
			return
		}

		// Get the stored CDP
		querierParams := cdp.QueryCdpsParams{
			Owner:           requestBody.Cdp.Owner,
			CollateralDenom: requestBody.Cdp.CollateralDenom,
		}
		querierParamsBz, err := cdc.MarshalJSON(querierParams)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", cdp.QueryGetCdps), querierParamsBz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		var cdps cdp.CDPs
		err = cdc.UnmarshalJSON(res, &cdps)
		if len(cdps) != 1 || err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Calculate CDP updates
		collateralDelta := requestBody.Cdp.CollateralAmount.Sub(cdps[0].CollateralAmount)
		debtDelta := requestBody.Cdp.Debt.Sub(cdps[0].Debt)

		// Create and return msg
		msg := cdp.NewMsgCreateOrModifyCDP(
			requestBody.Cdp.Owner,
			requestBody.Cdp.CollateralDenom,
			collateralDelta,
			debtDelta,
		)
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, requestBody.BaseReq, []sdk.Msg{msg})
	}
}

func getParamsHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the params
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/cdp/%s", cdp.QueryGetParams), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Return the params
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}
