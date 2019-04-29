package pricefeed

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultCodespace codespace for the module
	DefaultCodespace sdk.CodespaceType = ModuleName

	// CodeEmptyInput error code for empty input errors
	CodeEmptyInput sdk.CodeType = 1
)

// ErrEmptyInput Error constructor
func ErrEmptyInput(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyInput, fmt.Sprintf("Input must not be empty."))
}
