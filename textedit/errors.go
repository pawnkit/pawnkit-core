package textedit

import "errors"

// Use errors.Is to test for these; they may be wrapped with extra context.
var (
	ErrInvalidEdit       = errors.New("textedit: invalid edit")
	ErrOverlappingEdits  = errors.New("textedit: overlapping edits")
	ErrRangeOutOfBounds  = errors.New("textedit: edit range out of bounds")
	ErrStaleDocument     = errors.New("textedit: stale document version")
	ErrUnknownDocument   = errors.New("textedit: unknown document")
	ErrDuplicateDocument = errors.New("textedit: duplicate document in workspace edit")
)
