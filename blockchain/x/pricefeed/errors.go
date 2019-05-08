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
	// CodeExpired error code for expired prices
	CodeExpired sdk.CodeType = 2
	// CodeInvalid error code for all input prices expired
	CodeInvalid sdk.CodeType = 3
)

// ErrEmptyInput Error constructor
func ErrEmptyInput(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyInput, fmt.Sprintf("Input must not be empty."))
}

// ErrExpired Error constructor for posted price messages with expired price
func ErrExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpired, fmt.Sprintf("Price is expired."))
}

// ErrNoValidPrice Error constructor for posted price messages with expired price
func ErrNoValidPrice(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalid, fmt.Sprintf("All input prices are expired."))
}
