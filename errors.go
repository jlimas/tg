package main

import (
	"errors"
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/output"
)

// flagError renders a cliflags parse error in AXI's structured error format
// and returns the usage exit code (2).
func flagError(err error, usage string) int {
	var unknown *cliflags.UnknownFlagError
	var unexpected *cliflags.UnexpectedArgError
	var missing *cliflags.MissingValueError

	switch {
	case errors.As(err, &unknown):
		output.Error(unknown.Error(), usage)
	case errors.As(err, &unexpected):
		output.Error(unexpected.Error(), usage)
	case errors.As(err, &missing):
		output.Error(missing.Error(), usage)
	default:
		output.Error(fmt.Sprintf("%v", err), usage)
	}
	return 2
}
