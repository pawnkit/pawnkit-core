package diagnostic

import "errors"

// Use errors.Is to test for these; they may be wrapped with extra context.
var (
	ErrMissingCode        = errors.New("diagnostic: missing code")
	ErrMissingSource      = errors.New("diagnostic: missing source")
	ErrMissingMessage     = errors.New("diagnostic: missing message")
	ErrInvalidSeverity    = errors.New("diagnostic: invalid severity")
	ErrInvalidPrimary     = errors.New("diagnostic: invalid primary location")
	ErrInvalidFix         = errors.New("diagnostic: invalid fix")
	ErrInvalidSuppression = errors.New("diagnostic: invalid suppression")
)
