package textedit_test

import (
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func span(t *testing.T, f source.FileID, start, end source.Offset) source.Span {
	t.Helper()

	sp, err := source.NewSpan(f, start, end)
	if err != nil {
		t.Fatalf("NewSpan(%d,%d) error: %v", start, end, err)
	}

	return sp
}

func TestEditPredicates(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	insertion := textedit.Edit{Span: span(t, f, 3, 3), NewText: "hi"}
	if !insertion.IsInsertion() || insertion.IsDeletion() {
		t.Fatalf("expected insertion, got %+v", insertion)
	}

	deletion := textedit.Edit{Span: span(t, f, 3, 6), NewText: ""}
	if !deletion.IsDeletion() || deletion.IsInsertion() {
		t.Fatalf("expected deletion, got %+v", deletion)
	}

	replacement := textedit.Edit{Span: span(t, f, 3, 6), NewText: "hi"}
	if replacement.IsDeletion() || replacement.IsInsertion() {
		t.Fatalf("expected plain replacement, got %+v", replacement)
	}

	var zero textedit.Edit
	if zero.IsValid() {
		t.Fatalf("zero-value Edit should be invalid")
	}
}

func TestApplyNoEdits(t *testing.T) {
	t.Parallel()

	got, err := textedit.Apply("hello", nil)
	if err != nil || got != "hello" {
		t.Fatalf("Apply(nil) = (%q, %v), want (%q, nil)", got, err, "hello")
	}
}

func TestApplySingleInsertion(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	got, err := textedit.Apply("hello world", []textedit.Edit{
		{Span: span(t, f, 5, 5), NewText: ","},
	})
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if got != "hello, world" {
		t.Fatalf("Apply() = %q, want %q", got, "hello, world")
	}
}

func TestApplyMultipleNonOverlappingEditsAnyOrder(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	content := "one two three"

	// Deliberately given out of order to confirm Apply sorts internally.
	edits := []textedit.Edit{
		{Span: span(t, f, 8, 13), NewText: "3"},
		{Span: span(t, f, 0, 3), NewText: "1"},
		{Span: span(t, f, 4, 7), NewText: "2"},
	}

	got, err := textedit.Apply(content, edits)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if got != "1 2 3" {
		t.Fatalf("Apply() = %q, want %q", got, "1 2 3")
	}
}

func TestApplyRejectsOverlap(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	edits := []textedit.Edit{
		{Span: span(t, f, 0, 5), NewText: "a"},
		{Span: span(t, f, 3, 8), NewText: "b"},
	}

	if _, err := textedit.Apply("0123456789", edits); !errors.Is(err, textedit.ErrOverlappingEdits) {
		t.Fatalf("err = %v, want ErrOverlappingEdits", err)
	}
}

func TestApplyAllowsAdjacentEdits(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	edits := []textedit.Edit{
		{Span: span(t, f, 0, 5), NewText: "AAAAA"},
		{Span: span(t, f, 5, 10), NewText: "BBBBB"},
	}

	got, err := textedit.Apply("0123456789", edits)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if got != "AAAAABBBBB" {
		t.Fatalf("Apply() = %q", got)
	}
}

func TestApplySameOffsetInsertionsPreserveOrder(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	edits := []textedit.Edit{
		{Span: span(t, f, 2, 2), NewText: "A"},
		{Span: span(t, f, 2, 2), NewText: "B"},
	}

	got, err := textedit.Apply("xxyy", edits)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	// Same-offset insertions are not treated as overlapping; Apply applies
	// them in the order given, so "A" precedes "B" here.
	want := "xxAByy"
	if got != want {
		t.Fatalf("Apply() = %q, want %q (insertion order preserved)", got, want)
	}
}

func TestApplyRejectsOutOfBoundsEdit(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	edits := []textedit.Edit{
		{Span: span(t, f, 0, 100), NewText: "x"},
	}

	if _, err := textedit.Apply("short", edits); !errors.Is(err, textedit.ErrRangeOutOfBounds) {
		t.Fatalf("err = %v, want ErrRangeOutOfBounds", err)
	}
}

func TestApplyRejectsInvalidEdit(t *testing.T) {
	t.Parallel()

	var zero textedit.Edit

	if _, err := textedit.Apply("content", []textedit.Edit{zero}); !errors.Is(err, textedit.ErrInvalidEdit) {
		t.Fatalf("err = %v, want ErrInvalidEdit", err)
	}
}

func TestValidateEditsRejectsCrossFileEdits(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f1 := reg.Intern("file:///a.pwn")
	f2 := reg.Intern("file:///b.pwn")

	edits := []textedit.Edit{
		{Span: span(t, f1, 0, 1), NewText: "a"},
		{Span: span(t, f2, 0, 1), NewText: "b"},
	}

	if err := textedit.ValidateEdits(edits); !errors.Is(err, source.ErrSpanCrossesFiles) {
		t.Fatalf("err = %v, want ErrSpanCrossesFiles", err)
	}
}

func TestValidateEditsEmpty(t *testing.T) {
	t.Parallel()

	if err := textedit.ValidateEdits(nil); err != nil {
		t.Fatalf("ValidateEdits(nil) = %v, want nil", err)
	}
}

func TestApplyDeletion(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	got, err := textedit.Apply("hello world", []textedit.Edit{
		{Span: span(t, f, 5, 11), NewText: ""},
	})
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if got != "hello" {
		t.Fatalf("Apply() = %q, want %q", got, "hello")
	}
}
