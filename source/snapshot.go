package source

import "fmt"

// Snapshot is an immutable view of one file's content at a specific
// version, used to check computed edits/diagnostics for staleness against
// a later version (see the textedit package).
//
// The zero value is invalid; use [NewSnapshot].
type Snapshot struct {
	file    FileID
	version int32
	content string
	lines   *LineIndex
}

// NewSnapshot builds a Snapshot for file at version, over content. version
// is caller-defined; tools that don't track versions may use 0 consistently
// and skip staleness checks.
func NewSnapshot(file FileID, version int32, content string) (Snapshot, error) {
	if !file.IsValid() {
		return Snapshot{}, ErrInvalidFile
	}

	return Snapshot{
		file:    file,
		version: version,
		content: content,
		lines:   NewLineIndex(content),
	}, nil
}

func (s Snapshot) File() FileID {
	return s.file
}

func (s Snapshot) Version() int32 {
	return s.version
}

func (s Snapshot) Content() string {
	return s.content
}

func (s Snapshot) Lines() *LineIndex {
	return s.lines
}

func (s Snapshot) IsValid() bool {
	return s.file.IsValid()
}

// FullSpan returns a Span covering the entire content of the snapshot.
func (s Snapshot) FullSpan() Span {
	return Span{File: s.file, Start: 0, End: Offset(len(s.content))}
}

// Span builds a validated Span within this snapshot's content.
func (s Snapshot) Span(start, end Offset) (Span, error) {
	if start > end {
		return Span{}, fmt.Errorf("%w: start %d > end %d", ErrInvalidSpan, start, end)
	}

	if !s.lines.ValidOffset(start) {
		return Span{}, fmt.Errorf("%w: start %d", ErrOffsetOutOfRange, start)
	}

	if !s.lines.ValidOffset(end) {
		return Span{}, fmt.Errorf("%w: end %d", ErrOffsetOutOfRange, end)
	}

	return Span{File: s.file, Start: start, End: end}, nil
}

// Text returns the substring of content covered by span. span.File must
// match this snapshot's file.
func (s Snapshot) Text(span Span) (string, error) {
	if span.File != s.file {
		return "", ErrSpanCrossesFiles
	}

	if int(span.End) > len(s.content) || span.Start < 0 || span.Start > span.End {
		return "", fmt.Errorf("%w: %v", ErrOffsetOutOfRange, span)
	}

	return s.content[span.Start:span.End], nil
}
