package source

import "fmt"

// Encoding selects the unit used to count "characters" within a line,
// mirroring the LSP positionEncodingKind negotiation ("utf-8", "utf-16",
// "utf-32").
//
// The zero value is [UTF8].
type Encoding uint8

const (
	// UTF8 counts characters as bytes.
	UTF8 Encoding = iota

	// UTF16 counts characters as UTF-16 code units (supplementary-plane
	// runes, e.g. most emoji, count as 2). This is the LSP default.
	UTF16

	// UTF32 counts characters as Unicode code points (Go runes).
	UTF32
)

func (e Encoding) String() string {
	switch e {
	case UTF8:
		return "utf-8"
	case UTF16:
		return "utf-16"
	case UTF32:
		return "utf-32"
	default:
		return fmt.Sprintf("Encoding(%d)", uint8(e))
	}
}

// Position is a zero-based line/character position, in whatever [Encoding]
// the caller requested when it was produced; track the Encoding alongside
// it, since a Position is meaningless without knowing which was used.
//
// The zero value is valid.
type Position struct {
	Line      int
	Character int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Character)
}

// Range is a half-open line/character range [Start, End).
//
// The zero value is valid.
type Range struct {
	Start Position
	End   Position
}

func (r Range) String() string {
	return fmt.Sprintf("%s-%s", r.Start, r.End)
}
