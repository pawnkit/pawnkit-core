package source_test

import (
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestPositionZeroValue(t *testing.T) {
	t.Parallel()

	var p source.Position
	if p.Line != 0 || p.Character != 0 {
		t.Fatalf("zero-value Position = %v, want (0,0)", p)
	}
}

func TestEncodingString(t *testing.T) {
	t.Parallel()

	tests := map[source.Encoding]string{
		source.UTF8:  "utf-8",
		source.UTF16: "utf-16",
		source.UTF32: "utf-32",
	}

	for enc, want := range tests {
		if got := enc.String(); got != want {
			t.Errorf("Encoding(%d).String() = %q, want %q", enc, got, want)
		}
	}

	var zero source.Encoding
	if zero != source.UTF8 {
		t.Fatalf("zero-value Encoding should be UTF8")
	}
}

func TestRangeString(t *testing.T) {
	t.Parallel()

	r := source.Range{Start: source.Position{Line: 1, Character: 2}, End: source.Position{Line: 1, Character: 5}}
	if r.String() != "1:2-1:5" {
		t.Fatalf("String() = %q", r.String())
	}
}
