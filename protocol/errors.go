package protocol

import "errors"

// Use errors.Is to test for these; they may be wrapped with extra context.
var (
	ErrUnknownFile              = errors.New("protocol: unknown file id")
	ErrUnsupportedSchemaVersion = errors.New("protocol: unsupported schema version")
	ErrMissingSpan              = errors.New("protocol: location is missing its byte span")
	ErrInvalidSeverity          = errors.New("protocol: invalid severity")
	ErrInvalidFixKind           = errors.New("protocol: invalid fix kind")
)
