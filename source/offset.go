package source

import "strconv"

// Offset is a zero-based byte offset into a file's content; the internal
// source of truth for all positions in PawnKit.
//
// The zero value denotes the start of a file and is always valid on its
// own; validity against a particular content depends on that content's
// length (see [LineIndex.ValidOffset]).
type Offset int

func (o Offset) String() string {
	return strconv.Itoa(int(o))
}
