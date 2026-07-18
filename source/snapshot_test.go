package source_test

import (
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestSnapshotZeroValueInvalid(t *testing.T) {
	t.Parallel()

	var zero source.Snapshot
	if zero.IsValid() {
		t.Fatalf("zero-value Snapshot should be invalid")
	}
}

func TestNewSnapshotRejectsInvalidFile(t *testing.T) {
	t.Parallel()

	_, err := source.NewSnapshot(source.FileID(0), 1, "content")
	if !errors.Is(err, source.ErrInvalidFile) {
		t.Fatalf("err = %v, want ErrInvalidFile", err)
	}
}

func TestSnapshotBasics(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap, err := source.NewSnapshot(f, 3, "hello\nworld")
	if err != nil {
		t.Fatalf("NewSnapshot error: %v", err)
	}

	if snap.File() != f {
		t.Fatalf("File() = %v, want %v", snap.File(), f)
	}

	if snap.Version() != 3 {
		t.Fatalf("Version() = %d, want 3", snap.Version())
	}

	if snap.Content() != "hello\nworld" {
		t.Fatalf("Content() = %q", snap.Content())
	}

	if snap.Lines().LineCount() != 2 {
		t.Fatalf("Lines().LineCount() = %d, want 2", snap.Lines().LineCount())
	}

	full := snap.FullSpan()
	if full.Start != 0 || int(full.End) != len("hello\nworld") {
		t.Fatalf("FullSpan() = %v", full)
	}
}

func TestSnapshotSpanAndText(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap, err := source.NewSnapshot(f, 1, "hello world")
	if err != nil {
		t.Fatalf("NewSnapshot error: %v", err)
	}

	sp, err := snap.Span(0, 5)
	if err != nil {
		t.Fatalf("Span error: %v", err)
	}

	text, err := snap.Text(sp)
	if err != nil {
		t.Fatalf("Text error: %v", err)
	}

	if text != "hello" {
		t.Fatalf("Text() = %q, want %q", text, "hello")
	}
}

func TestSnapshotSpanOutOfRange(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	snap, err := source.NewSnapshot(f, 1, "hi")
	if err != nil {
		t.Fatalf("NewSnapshot error: %v", err)
	}

	if _, err := snap.Span(0, 10); !errors.Is(err, source.ErrOffsetOutOfRange) {
		t.Fatalf("Span(0,10) err = %v, want ErrOffsetOutOfRange", err)
	}

	if _, err := snap.Span(5, 1); err == nil {
		t.Fatalf("expected error for start > end")
	}
}

func TestSnapshotTextRejectsCrossFileSpan(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f1 := reg.Intern("file:///a.pwn")
	f2 := reg.Intern("file:///b.pwn")

	snap, err := source.NewSnapshot(f1, 1, "hi")
	if err != nil {
		t.Fatalf("NewSnapshot error: %v", err)
	}

	_, err = snap.Text(source.Span{File: f2, Start: 0, End: 2})
	if !errors.Is(err, source.ErrSpanCrossesFiles) {
		t.Fatalf("err = %v, want ErrSpanCrossesFiles", err)
	}
}
