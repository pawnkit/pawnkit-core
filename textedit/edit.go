package textedit

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/source"
)

// AnyVersion applies an edit regardless of the document's current version.
const AnyVersion int32 = -1

// Edit replaces Span with NewText. Its zero value is invalid.
type Edit struct {
	Span    source.Span
	NewText string
}

// IsValid reports whether the edit span is valid.
func (e Edit) IsValid() bool {
	return e.Span.IsValid()
}

// IsInsertion reports whether e inserts text without replacing bytes.
func (e Edit) IsInsertion() bool {
	return e.Span.IsEmpty() && e.NewText != ""
}

// IsDeletion reports whether e removes bytes without adding text.
func (e Edit) IsDeletion() bool {
	return !e.Span.IsEmpty() && e.NewText == ""
}

func (e Edit) String() string {
	return fmt.Sprintf("%v -> %q", e.Span, e.NewText)
}

// DocumentEdit groups the edits meant for a single file together with the
// document version they were computed against. Use [AnyVersion] to skip
// the staleness check.
type DocumentEdit struct {
	File    source.FileID
	Version int32
	Edits   []Edit
}

// WorkspaceEdit groups DocumentEdits for one or more files into a single
// logical, transactional change.
type WorkspaceEdit struct {
	Documents []DocumentEdit
}

// ValidateEdits checks validity, file identity, and overlap.
// Same-offset insertions retain their input order.
func ValidateEdits(edits []Edit) error {
	if len(edits) == 0 {
		return nil
	}

	var file source.FileID

	for i, e := range edits {
		if !e.IsValid() {
			return fmt.Errorf("%w: edit %d: %v", ErrInvalidEdit, i, e)
		}

		if i == 0 {
			file = e.Span.File
		} else if e.Span.File != file {
			return fmt.Errorf("%w: edit %d belongs to a different file than edit 0", source.ErrSpanCrossesFiles, i)
		}
	}

	sorted := sortedIndices(edits)

	maxEnd := source.Offset(-1)

	for _, i := range sorted {
		e := edits[i]
		if e.Span.Start < maxEnd {
			return fmt.Errorf("%w: edit %d %v overlaps a preceding edit", ErrOverlappingEdits, i, e)
		}

		if e.Span.End > maxEnd {
			maxEnd = e.Span.End
		}
	}

	return nil
}

// sortedIndices returns indices sorted by (Start, End), stably.
func sortedIndices(edits []Edit) []int {
	idx := make([]int, len(edits))
	for i := range idx {
		idx[i] = i
	}

	// Insertion sort: edit counts per document are small, so this is
	// simpler and faster than sort.SliceStable here.
	for i := 1; i < len(idx); i++ {
		j := i
		for j > 0 && less(edits[idx[j]], edits[idx[j-1]]) {
			idx[j], idx[j-1] = idx[j-1], idx[j]
			j--
		}
	}

	return idx
}

func less(a, b Edit) bool {
	if a.Span.Start != b.Span.Start {
		return a.Span.Start < b.Span.Start
	}

	return a.Span.End < b.Span.End
}

// Apply computes the result of applying edits to content, without mutating
// content. edits do not need to be pre-sorted.
func Apply(content string, edits []Edit) (string, error) {
	if len(edits) == 0 {
		return content, nil
	}

	if err := ValidateEdits(edits); err != nil {
		return "", err
	}

	order := sortedIndices(edits)

	for _, i := range order {
		e := edits[i]
		if int(e.Span.End) > len(content) {
			return "", fmt.Errorf("%w: edit %v end exceeds content length %d", ErrRangeOutOfBounds, e, len(content))
		}
	}

	var b []byte

	estimate := len(content)
	for _, e := range edits {
		estimate += len(e.NewText) - int(e.Span.Len())
	}

	if estimate > 0 {
		b = make([]byte, 0, estimate)
	}

	cursor := source.Offset(0)

	for _, i := range order {
		e := edits[i]
		b = append(b, content[cursor:e.Span.Start]...)
		b = append(b, e.NewText...)
		cursor = e.Span.End
	}

	b = append(b, content[cursor:]...)

	return string(b), nil
}
