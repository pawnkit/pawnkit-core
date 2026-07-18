package source

import "fmt"

// Span is a half-open byte range [Start, End) within a single file.
//
// The zero value has an invalid File and is therefore always invalid, even
// though Start and End are both zero; use [NewSpan] or [Span.IsValid].
type Span struct {
	File  FileID
	Start Offset
	End   Offset
}

// NewSpan builds a Span, validating that file is valid, both offsets are
// non-negative, and Start <= End.
func NewSpan(file FileID, start, end Offset) (Span, error) {
	s := Span{File: file, Start: start, End: end}
	if err := s.Validate(); err != nil {
		return Span{}, err
	}

	return s, nil
}

// Validate does not check the offsets against any content length; use
// [LineIndex] helpers for that.
func (s Span) Validate() error {
	if !s.File.IsValid() {
		return fmt.Errorf("%w: invalid file", ErrInvalidSpan)
	}

	if s.Start < 0 || s.End < 0 {
		return fmt.Errorf("%w: negative offset", ErrInvalidSpan)
	}

	if s.Start > s.End {
		return fmt.Errorf("%w: start %d > end %d", ErrInvalidSpan, s.Start, s.End)
	}

	return nil
}

// IsValid reports whether s passes validation.
func (s Span) IsValid() bool {
	return s.Validate() == nil
}

// IsEmpty reports whether s spans zero bytes (an insertion point).
func (s Span) IsEmpty() bool {
	return s.Start == s.End
}

// Len returns the span length in bytes.
func (s Span) Len() Offset {
	return s.End - s.Start
}

// Contains reports whether off falls within s. An empty span never
// contains any offset, including its own Start.
func (s Span) Contains(off Offset) bool {
	return off >= s.Start && off < s.End
}

// ContainsOrTouches also accepts either boundary, useful for insertion
// points that should count as "at" a span's edge.
func (s Span) ContainsOrTouches(off Offset) bool {
	return off >= s.Start && off <= s.End
}

// Overlaps reports whether s and other describe overlapping byte ranges in
// the same file. Spans that only touch at a boundary do not overlap.
func (s Span) Overlaps(other Span) bool {
	if s.File != other.File {
		return false
	}

	return s.Start < other.End && other.Start < s.End
}

// Union panics if s and other belong to different files.
func (s Span) Union(other Span) Span {
	if s.File != other.File {
		panic("source: Union of spans from different files")
	}

	start := min(s.Start, other.Start)
	end := max(s.End, other.End)

	return Span{File: s.File, Start: start, End: end}
}

// String returns a compact "file(id)[start:end)" form for debugging; it
// omits any URI, since Span only carries a process-local FileID.
func (s Span) String() string {
	return fmt.Sprintf("file(%d)[%d:%d)", s.File, s.Start, s.End)
}
