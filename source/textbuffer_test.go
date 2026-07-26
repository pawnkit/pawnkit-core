package source_test

import (
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestTextBufferEditsShareRevisions(t *testing.T) {
	t.Parallel()
	original := source.NewTextBuffer([]byte("first\nsecond\n"))
	edited, err := original.Apply(6, 12, "changed")
	if err != nil {
		t.Fatal(err)
	}
	inserted, err := edited.Apply(0, 0, "start\n")
	if err != nil {
		t.Fatal(err)
	}

	if got := string(original.Bytes()); got != "first\nsecond\n" {
		t.Fatalf("original = %q", got)
	}
	if got := string(edited.Bytes()); got != "first\nchanged\n" {
		t.Fatalf("edited = %q", got)
	}
	if got := string(inserted.Bytes()); got != "start\nfirst\nchanged\n" {
		t.Fatalf("inserted = %q", got)
	}
}

func TestTextBufferApplyValidation(t *testing.T) {
	t.Parallel()
	buffer := source.NewTextBuffer([]byte("café"))
	for _, edit := range [][2]source.Offset{{3, 4}, {5, 4}, {0, 10}} {
		if _, err := buffer.Apply(edit[0], edit[1], "x"); err == nil {
			t.Fatalf("Apply(%d, %d) succeeded", edit[0], edit[1])
		}
	}
}

func TestTextBufferNoopReturnsSameRevision(t *testing.T) {
	t.Parallel()
	buffer := source.NewTextBuffer([]byte("value"))
	next, err := buffer.Apply(2, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if next != buffer {
		t.Fatal("no-op edit created a revision")
	}
}
