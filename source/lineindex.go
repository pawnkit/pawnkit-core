package source

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// LineIndex maps byte offsets to positions in immutable text.
// Use [NewLineIndex]; the zero value is unusable.
type LineIndex struct {
	content    string
	lineStarts []Offset // lineStarts[i] is the byte offset where line i begins; lineStarts[0] == 0.
}

// NewLineIndex builds a LineIndex over content. content is retained, not
// copied; callers must treat it as immutable.
func NewLineIndex(content string) *LineIndex {
	starts := []Offset{0}

	for i := range len(content) {
		if content[i] == '\n' {
			starts = append(starts, Offset(i+1))
		}
	}

	return &LineIndex{content: content, lineStarts: starts}
}

// Apply returns an index for one byte-range replacement.
func (idx *LineIndex) Apply(start, end Offset, replacement string) (*LineIndex, error) {
	if idx == nil {
		return nil, fmt.Errorf("%w: nil line index", ErrInvalidSpan)
	}
	if !idx.ValidOffset(start) || !idx.ValidOffset(end) || end < start {
		return nil, fmt.Errorf("%w: replacement range [%d,%d)", ErrInvalidSpan, start, end)
	}

	var content strings.Builder
	content.Grow(len(idx.content) - int(end-start) + len(replacement))
	content.WriteString(idx.content[:start])
	content.WriteString(replacement)
	content.WriteString(idx.content[end:])

	starts := make([]Offset, 0, len(idx.lineStarts))
	for _, lineStart := range idx.lineStarts {
		if lineStart > start {
			break
		}
		starts = append(starts, lineStart)
	}
	for i := range len(replacement) {
		if replacement[i] == '\n' {
			starts = appendUniqueOffset(starts, start+Offset(i+1))
		}
	}
	delta := Offset(len(replacement)) - (end - start)
	suffix := sort.Search(len(idx.lineStarts), func(i int) bool {
		return idx.lineStarts[i] > end
	})
	for _, lineStart := range idx.lineStarts[suffix:] {
		starts = appendUniqueOffset(starts, lineStart+delta)
	}

	return &LineIndex{content: content.String(), lineStarts: starts}, nil
}

func appendUniqueOffset(offsets []Offset, offset Offset) []Offset {
	if len(offsets) == 0 || offsets[len(offsets)-1] != offset {
		return append(offsets, offset)
	}
	return offsets
}

// Content returns the indexed text.
func (idx *LineIndex) Content() string {
	return idx.content
}

// LineCount returns the number of lines; always at least 1, even for empty content.
func (idx *LineIndex) LineCount() int {
	return len(idx.lineStarts)
}

// ValidOffset reports whether off is within [0, len(content)] and does not
// split a multi-byte UTF-8 rune.
func (idx *LineIndex) ValidOffset(off Offset) bool {
	if off < 0 || int(off) > len(idx.content) {
		return false
	}

	if int(off) == len(idx.content) {
		return true
	}

	return idx.content[off]&0xC0 != 0x80
}

// LineStart returns the byte offset where line begins. Lines are numbered from 0.
func (idx *LineIndex) LineStart(line int) (Offset, error) {
	if line < 0 || line >= len(idx.lineStarts) {
		return 0, fmt.Errorf("%w: line %d (have %d lines)", ErrLineOutOfRange, line, len(idx.lineStarts))
	}

	return idx.lineStarts[line], nil
}

// LineEnd returns the byte offset where line's content ends, excluding any
// terminator. For the final line, this is len(content).
func (idx *LineIndex) LineEnd(line int) (Offset, error) {
	start, err := idx.LineStart(line)
	if err != nil {
		return 0, err
	}

	if line+1 < len(idx.lineStarts) {
		end := idx.lineStarts[line+1]
		if end > start && idx.content[end-1] == '\n' {
			end--
			if end > start && idx.content[end-1] == '\r' {
				end--
			}
		}

		return end, nil
	}

	return Offset(len(idx.content)), nil
}

// LineSpan returns the Span covering line's content, excluding its terminator.
func (idx *LineIndex) LineSpan(file FileID, line int) (Span, error) {
	start, err := idx.LineStart(line)
	if err != nil {
		return Span{}, err
	}

	end, err := idx.LineEnd(line)
	if err != nil {
		return Span{}, err
	}

	return Span{File: file, Start: start, End: end}, nil
}

// LineAt returns the (0-based) line number containing byte offset off.
func (idx *LineIndex) LineAt(off Offset) (int, error) {
	if !idx.ValidOffset(off) {
		return 0, fmt.Errorf("%w: %d (content length %d)", ErrOffsetOutOfRange, off, len(idx.content))
	}

	// Largest i such that lineStarts[i] <= off.
	line := sort.Search(len(idx.lineStarts), func(i int) bool {
		return idx.lineStarts[i] > off
	}) - 1

	return line, nil
}

// Position converts a byte offset into a line/character Position using enc
// to count characters within the line.
func (idx *LineIndex) Position(off Offset, enc Encoding) (Position, error) {
	if !idx.ValidOffset(off) {
		if off < 0 || int(off) > len(idx.content) {
			return Position{}, fmt.Errorf("%w: %d (content length %d)", ErrOffsetOutOfRange, off, len(idx.content))
		}

		return Position{}, fmt.Errorf("%w: offset %d", ErrInvalidUTF8Boundary, off)
	}

	line, err := idx.LineAt(off)
	if err != nil {
		return Position{}, err
	}

	start := idx.lineStarts[line]

	end, err := idx.LineEnd(line)
	if err != nil {
		return Position{}, err
	}

	if off > end {
		// off falls strictly inside a "\r\n" terminator, which has no
		// corresponding line/character position.
		return Position{}, fmt.Errorf("%w: offset %d falls inside the line terminator of line %d", ErrOffsetInLineTerminator, off, line)
	}

	character := countCharacters(idx.content[start:off], enc)

	return Position{Line: line, Character: character}, nil
}

// Offset converts a line/character Position, interpreted with enc, back
// into a byte offset.
func (idx *LineIndex) Offset(pos Position, enc Encoding) (Offset, error) {
	start, err := idx.LineStart(pos.Line)
	if err != nil {
		return 0, err
	}

	end, err := idx.LineEnd(pos.Line)
	if err != nil {
		return 0, err
	}

	if pos.Character < 0 {
		return 0, fmt.Errorf("%w: negative character %d", ErrColumnOutOfRange, pos.Character)
	}

	off, ok := advanceCharacters(idx.content[start:end], pos.Character, enc)
	if !ok {
		return 0, fmt.Errorf(
			"%w: character %d splits a UTF-16 surrogate pair on line %d",
			ErrSurrogateSplit, pos.Character, pos.Line,
		)
	}

	if off < 0 {
		return 0, fmt.Errorf(
			"%w: character %d beyond end of line %d",
			ErrColumnOutOfRange, pos.Character, pos.Line,
		)
	}

	return start + Offset(off), nil
}

// Range converts span (whose File is ignored) into a line/character Range using enc.
func (idx *LineIndex) Range(span Span, enc Encoding) (Range, error) {
	startPos, err := idx.Position(span.Start, enc)
	if err != nil {
		return Range{}, fmt.Errorf("start: %w", err)
	}

	endPos, err := idx.Position(span.End, enc)
	if err != nil {
		return Range{}, fmt.Errorf("end: %w", err)
	}

	return Range{Start: startPos, End: endPos}, nil
}

// Span converts rng, interpreted with enc, into a byte Span for file.
func (idx *LineIndex) Span(file FileID, rng Range, enc Encoding) (Span, error) {
	start, err := idx.Offset(rng.Start, enc)
	if err != nil {
		return Span{}, fmt.Errorf("start: %w", err)
	}

	end, err := idx.Offset(rng.End, enc)
	if err != nil {
		return Span{}, fmt.Errorf("end: %w", err)
	}

	return Span{File: file, Start: start, End: end}, nil
}

func countCharacters(s string, enc Encoding) int {
	switch enc {
	case UTF8:
		return len(s)
	case UTF32:
		return utf8.RuneCountInString(s)
	case UTF16:
		count := 0

		for _, r := range s {
			units := utf16.RuneLen(r)
			if units < 0 {
				units = 1 // invalid rune: treat as one replacement unit.
			}

			count += units
		}

		return count
	default:
		return len(s)
	}
}

// advanceCharacters consumes n characters (in the units of enc) from the
// start of s and returns the resulting byte offset. ok is false if n lands
// in the middle of a UTF-16 surrogate pair. A negative returned offset
// (with ok == true) indicates n exceeds the number of characters in s.
func advanceCharacters(s string, n int, enc Encoding) (offset int, ok bool) {
	if n == 0 {
		return 0, true
	}

	switch enc {
	case UTF8:
		if n > len(s) {
			return -1, true
		}

		return n, true
	case UTF32:
		remaining := n
		i := 0

		for i < len(s) {
			if remaining == 0 {
				return i, true
			}

			_, width := utf8.DecodeRuneInString(s[i:])
			i += width
			remaining--
		}

		if remaining == 0 {
			return i, true
		}

		return -1, true
	case UTF16:
		remaining := n
		i := 0

		for i < len(s) {
			if remaining == 0 {
				return i, true
			}

			r, width := utf8.DecodeRuneInString(s[i:])

			units := utf16.RuneLen(r)
			if units < 0 {
				units = 1
			}

			if units == 2 && remaining == 1 {
				// n lands between the two halves of a surrogate pair.
				return 0, false
			}

			i += width
			remaining -= units
		}

		if remaining == 0 {
			return i, true
		}

		return -1, true
	default:
		if n > len(s) {
			return -1, true
		}

		return n, true
	}
}
