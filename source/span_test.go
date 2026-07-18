package source_test

import (
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestSpanZeroValueInvalid(t *testing.T) {
	t.Parallel()

	var zero source.Span
	if zero.IsValid() {
		t.Fatalf("zero-value Span should be invalid (invalid File)")
	}
}

func TestNewSpanValidation(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	if _, err := source.NewSpan(source.FileID(0), 0, 1); !errors.Is(err, source.ErrInvalidSpan) {
		t.Fatalf("invalid file: err = %v, want ErrInvalidSpan", err)
	}

	if _, err := source.NewSpan(f, 5, 2); !errors.Is(err, source.ErrInvalidSpan) {
		t.Fatalf("start > end: err = %v, want ErrInvalidSpan", err)
	}

	if _, err := source.NewSpan(f, -1, 2); !errors.Is(err, source.ErrInvalidSpan) {
		t.Fatalf("negative start: err = %v, want ErrInvalidSpan", err)
	}

	sp, err := source.NewSpan(f, 0, 0)
	if err != nil {
		t.Fatalf("empty span at 0: unexpected error %v", err)
	}

	if !sp.IsEmpty() {
		t.Fatalf("expected empty span")
	}
}

func TestSpanContains(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")
	sp := source.Span{File: f, Start: 5, End: 10}

	tests := []struct {
		off  source.Offset
		want bool
	}{
		{4, false},
		{5, true},
		{9, true},
		{10, false}, // half-open: End is exclusive.
	}

	for _, tc := range tests {
		if got := sp.Contains(tc.off); got != tc.want {
			t.Errorf("Contains(%d) = %v, want %v", tc.off, got, tc.want)
		}
	}

	empty := source.Span{File: f, Start: 5, End: 5}
	if empty.Contains(5) {
		t.Fatalf("empty span should not contain its own boundary")
	}

	if !empty.ContainsOrTouches(5) {
		t.Fatalf("ContainsOrTouches should include the boundary")
	}
}

func TestSpanOverlaps(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f1 := reg.Intern("file:///a.pwn")
	f2 := reg.Intern("file:///b.pwn")

	a := source.Span{File: f1, Start: 0, End: 10}
	b := source.Span{File: f1, Start: 5, End: 15}
	c := source.Span{File: f1, Start: 10, End: 20} // touches a, does not overlap.
	d := source.Span{File: f2, Start: 0, End: 10}  // same range, different file.

	if !a.Overlaps(b) {
		t.Errorf("a should overlap b")
	}

	if a.Overlaps(c) {
		t.Errorf("touching spans should not overlap")
	}

	if a.Overlaps(d) {
		t.Errorf("spans in different files should never overlap")
	}
}

func TestSpanUnion(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	a := source.Span{File: f, Start: 0, End: 5}
	b := source.Span{File: f, Start: 3, End: 10}

	u := a.Union(b)
	want := source.Span{File: f, Start: 0, End: 10}

	if u != want {
		t.Fatalf("Union = %v, want %v", u, want)
	}
}

func TestSpanUnionPanicsOnCrossFile(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f1 := reg.Intern("file:///a.pwn")
	f2 := reg.Intern("file:///b.pwn")

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for cross-file union")
		}
	}()

	source.Span{File: f1, Start: 0, End: 1}.Union(source.Span{File: f2, Start: 0, End: 1})
}

func TestSpanLen(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	sp := source.Span{File: f, Start: 3, End: 8}
	if sp.Len() != 5 {
		t.Fatalf("Len() = %d, want 5", sp.Len())
	}
}
