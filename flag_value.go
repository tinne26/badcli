package badcli

import "errors"

type FlagValue interface {
	// The given string will not include quotes nor "=" if those
	// are used. An empty string will be passed if no argument
	// has been given for the flag. If that's not allowed, just
	// return ErrMissingValue.
	ParseFromArg(string) error
}

var ErrMissingValue = errors.New("missing value")
