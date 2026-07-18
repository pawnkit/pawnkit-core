package source

import "errors"

// Use errors.Is to test for these; they may be wrapped with extra context.
var (
	ErrInvalidFile            = errors.New("source: invalid file id")
	ErrInvalidURI             = errors.New("source: invalid uri")
	ErrInvalidSpan            = errors.New("source: invalid span")
	ErrOffsetOutOfRange       = errors.New("source: offset out of range")
	ErrInvalidUTF8Boundary    = errors.New("source: offset does not fall on a UTF-8 rune boundary")
	ErrLineOutOfRange         = errors.New("source: line out of range")
	ErrColumnOutOfRange       = errors.New("source: character out of range")
	ErrSurrogateSplit         = errors.New("source: position splits a UTF-16 surrogate pair")
	ErrSpanCrossesFiles       = errors.New("source: span crosses files")
	ErrOffsetInLineTerminator = errors.New("source: offset falls inside a line terminator")
)
