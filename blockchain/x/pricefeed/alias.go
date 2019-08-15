// nolint
// aliases for exported functions of the pricefeed module used by internal module
package pricefeed

import "github.com/kava-labs/kava-devnet/blockchain/x/pricefeed/types"

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	RouterKey         = types.RouterKey
	CodeEmptyInput    = types.CodeEmptyInput
	CodeExpired       = types.CodeExpired
	CodeInvalidPrice  = types.CodeInvalidPrice
	CodeInvalidAsset  = types.CodeInvalidAsset
	CodeInvalidOracle = types.CodeInvalidOracle
	QueryCurrentPrice = types.QueryCurrentPrice
	QueryRawPrices    = types.QueryRawPrices
	QueryAssets       = types.QueryAssets
)

var (
	DefaultAssetParams   = types.DefaultAssetParams
	DefaultCodespace     = types.DefaultCodespace
	DefaultGenesisState  = types.DefaultGenesisState
	DefaultOracleParams  = types.DefaultOracleParams
	ErrEmptyInput        = types.ErrEmptyInput
	ErrExpired           = types.ErrExpired
	ErrNoValidPrice      = types.ErrNoValidPrice
	ErrInvalidAsset      = types.ErrInvalidAsset
	ErrInvalidOracle     = types.ErrInvalidOracle
	ModuleCdc            = types.ModuleCdc
	NewAssetParams       = types.NewAssetParams
	NewGenesisState      = types.NewGenesisState
	NewMsgPostPrice      = types.NewMsgPostPrice
	NewOracleParams      = types.NewOracleParams
	ParamKeyTable        = types.ParamKeyTable
	ParamStoreKeyOracles = types.ParamStoreKeyOracles
	ParamStoreKeyAssets  = types.ParamStoreKeyAssets
	RegisterCodec        = types.RegisterCodec
	ValidateGenesis      = types.ValidateGenesis
)

type (
	Asset              = types.Asset
	AssetParams        = types.AssetParams
	CurrentPrice       = types.CurrentPrice
	GenesisState       = types.GenesisState
	MsgPostPrice       = types.MsgPostPrice
	Oracle             = types.Oracle
	OracleParams       = types.OracleParams
	ParamSubspace      = types.ParamSubspace
	PostedPrice        = types.PostedPrice
	QueryRawPricesResp = types.QueryRawPricesResp
	QueryAssetsResp    = types.QueryAssetsResp
	SortDecs           = types.SortDecs
)
