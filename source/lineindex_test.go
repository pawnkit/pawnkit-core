package source_test

import (
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestLineIndexEmptyContent(t *testing.T) {
	t.Parallel()

	idx := source.NewLineIndex("")

	if idx.LineCount() != 1 {
		t.Fatalf("LineCount() = %d, want 1", idx.LineCount())
	}

	start, err := idx.LineStart(0)
	if err != nil || start != 0 {
		t.Fatalf("LineStart(0) = (%d, %v), want (0, nil)", start, err)
	}

	end, err := idx.LineEnd(0)
	if err != nil || end != 0 {
		t.Fatalf("LineEnd(0) = (%d, %v), want (0, nil)", end, err)
	}

	pos, err := idx.Position(0, source.UTF16)
	if err != nil {
		t.Fatalf("Position(0) error: %v", err)
	}

	if pos != (source.Position{Line: 0, Character: 0}) {
		t.Fatalf("Position(0) = %v, want (0,0)", pos)
	}
}

func TestLineIndexLFOnly(t *testing.T) {
	t.Parallel()

	content := "abc\ndef\nghi"
	idx := source.NewLineIndex(content)

	if idx.LineCount() != 3 {
		t.Fatalf("LineCount() = %d, want 3", idx.LineCount())
	}

	wantStarts := []source.Offset{0, 4, 8}
	for i, want := range wantStarts {
		got, err := idx.LineStart(i)
		if err != nil || got != want {
			t.Errorf("LineStart(%d) = (%d, %v), want (%d, nil)", i, got, err, want)
		}
	}

	wantEnds := []source.Offset{3, 7, 11}
	for i, want := range wantEnds {
		got, err := idx.LineEnd(i)
		if err != nil || got != want {
			t.Errorf("LineEnd(%d) = (%d, %v), want (%d, nil)", i, got, err, want)
		}
	}
}

func TestLineIndexCRLF(t *testing.T) {
	t.Parallel()

	content := "abc\r\ndef\r\nghi"
	idx := source.NewLineIndex(content)

	if idx.LineCount() != 3 {
		t.Fatalf("LineCount() = %d, want 3", idx.LineCount())
	}

	// Line starts point just after "\r\n".
	wantStarts := []source.Offset{0, 5, 10}
	for i, want := range wantStarts {
		got, err := idx.LineStart(i)
		if err != nil || got != want {
			t.Errorf("LineStart(%d) = (%d, %v), want (%d, nil)", i, got, err, want)
		}
	}

	// Line ends exclude the "\r\n" terminator.
	wantEnds := []source.Offset{3, 8, 13}
	for i, want := range wantEnds {
		got, err := idx.LineEnd(i)
		if err != nil || got != want {
			t.Errorf("LineEnd(%d) = (%d, %v), want (%d, nil)", i, got, err, want)
		}
	}
}

func TestLineIndexOffsetInsideCRLFTerminatorRejected(t *testing.T) {
	t.Parallel()

	// Regression test (found by FuzzLineIndexPosition): offsets strictly
	// inside a "\r\n" terminator must be rejected, not silently rounded.
	content := "000\r\n"
	idx := source.NewLineIndex(content)

	pos, err := idx.Position(3, source.UTF8)
	if err != nil {
		t.Fatalf("Position(3) unexpected error: %v", err)
	}

	if pos != (source.Position{Line: 0, Character: 3}) {
		t.Fatalf("Position(3) = %v, want (0,3)", pos)
	}

	if _, err := idx.Position(4, source.UTF8); !errors.Is(err, source.ErrOffsetInLineTerminator) {
		t.Fatalf("Position(4) err = %v, want ErrOffsetInLineTerminator", err)
	}

	pos, err = idx.Position(5, source.UTF8)
	if err != nil {
		t.Fatalf("Position(5) unexpected error: %v", err)
	}

	if pos != (source.Position{Line: 1, Character: 0}) {
		t.Fatalf("Position(5) = %v, want (1,0)", pos)
	}
}

func TestLineIndexMixedLineEndings(t *testing.T) {
	t.Parallel()

	content := "a\r\nb\nc"
	idx := source.NewLineIndex(content)

	if idx.LineCount() != 3 {
		t.Fatalf("LineCount() = %d, want 3", idx.LineCount())
	}

	for i, want := range []source.Offset{0, 3, 5} {
		got, err := idx.LineStart(i)
		if err != nil || got != want {
			t.Errorf("LineStart(%d) = (%d, %v), want (%d, nil)", i, got, err, want)
		}
	}
}

func TestLineIndexTrailingNewlineCreatesEmptyFinalLine(t *testing.T) {
	t.Parallel()

	content := "abc\n"
	idx := source.NewLineIndex(content)

	if idx.LineCount() != 2 {
		t.Fatalf("LineCount() = %d, want 2", idx.LineCount())
	}

	start, err := idx.LineStart(1)
	if err != nil || start != 4 {
		t.Fatalf("LineStart(1) = (%d, %v), want (4, nil)", start, err)
	}

	end, err := idx.LineEnd(1)
	if err != nil || end != 4 {
		t.Fatalf("LineEnd(1) = (%d, %v), want (4, nil)", end, err)
	}
}

func TestLineIndexApplyMatchesRebuild(t *testing.T) {
	t.Parallel()

	const twoLines = "one\ntwo\n"
	tests := []struct {
		name        string
		content     string
		start       source.Offset
		end         source.Offset
		replacement string
	}{
		{"insert into empty", "", 0, 0, "text"},
		{"insert text", twoLines, 5, 5, "X"},
		{"insert lines", twoLines, 4, 4, "new\nlines\n"},
		{"replace lines", "one\ntwo\nthree\n", 4, 8, "second\n"},
		{"delete newline", twoLines, 3, 4, ""},
		{"replace all", twoLines, 0, 8, "single"},
		{"unicode", "café\nnext\n", 2, 5, "at"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			index, err := source.NewLineIndex(test.content).Apply(test.start, test.end, test.replacement)
			if err != nil {
				t.Fatalf("Apply() error: %v", err)
			}
			wantContent := test.content[:test.start] + test.replacement + test.content[test.end:]
			want := source.NewLineIndex(wantContent)
			if index.Content() != wantContent {
				t.Fatalf("Content() = %q, want %q", index.Content(), wantContent)
			}
			if index.LineCount() != want.LineCount() {
				t.Fatalf("LineCount() = %d, want %d", index.LineCount(), want.LineCount())
			}
			for line := range want.LineCount() {
				gotStart, gotErr := index.LineStart(line)
				wantStart, wantErr := want.LineStart(line)
				if gotStart != wantStart || !errors.Is(gotErr, wantErr) {
					t.Fatalf("LineStart(%d) = (%d, %v), want (%d, %v)", line, gotStart, gotErr, wantStart, wantErr)
				}
			}
		})
	}
}

func TestLineIndexApplyRejectsInvalidRange(t *testing.T) {
	t.Parallel()

	index := source.NewLineIndex("café")
	for _, span := range [][2]source.Offset{{3, 4}, {5, 4}, {0, 10}} {
		if _, err := index.Apply(span[0], span[1], "x"); !errors.Is(err, source.ErrInvalidSpan) {
			t.Fatalf("Apply(%d, %d) error = %v, want ErrInvalidSpan", span[0], span[1], err)
		}
	}
}

func TestLineIndexLoneCRIsNotABreak(t *testing.T) {
	t.Parallel()

	content := "abc\rdef" // lone CR is not a line break
	idx := source.NewLineIndex(content)

	if idx.LineCount() != 1 {
		t.Fatalf("LineCount() = %d, want 1 (lone CR is not a line break)", idx.LineCount())
	}
}

func TestLineIndexOutOfRangeLine(t *testing.T) {
	t.Parallel()

	idx := source.NewLineIndex("abc")

	if _, err := idx.LineStart(1); !errors.Is(err, source.ErrLineOutOfRange) {
		t.Fatalf("LineStart(1) err = %v, want ErrLineOutOfRange", err)
	}

	if _, err := idx.LineStart(-1); !errors.Is(err, source.ErrLineOutOfRange) {
		t.Fatalf("LineStart(-1) err = %v, want ErrLineOutOfRange", err)
	}
}

func TestLineIndexPositionOffsetRoundTripASCII(t *testing.T) {
	t.Parallel()

	content := "line one\nline two\nline three"
	idx := source.NewLineIndex(content)

	for off := 0; off <= len(content); off++ {
		for _, enc := range []source.Encoding{source.UTF8, source.UTF16, source.UTF32} {
			pos, err := idx.Position(source.Offset(off), enc)
			if err != nil {
				t.Fatalf("Position(%d, %v) error: %v", off, enc, err)
			}

			back, err := idx.Offset(pos, enc)
			if err != nil {
				t.Fatalf("Offset(%v, %v) error: %v", pos, enc, err)
			}

			if int(back) != off {
				t.Fatalf("round trip mismatch at offset %d (enc %v): got %d via %v", off, enc, back, pos)
			}
		}
	}
}

func TestLineIndexUnicodeBMP(t *testing.T) {
	t.Parallel()

	content := "héllo\nworld" // é is 2 UTF-8 bytes, still 1 UTF-16 unit
	idx := source.NewLineIndex(content)

	pos, err := idx.Position(3, source.UTF16)
	if err != nil {
		t.Fatalf("Position error: %v", err)
	}

	if pos != (source.Position{Line: 0, Character: 2}) {
		t.Fatalf("Position(3, UTF16) = %v, want (0,2)", pos)
	}

	posUTF8, err := idx.Position(3, source.UTF8)
	if err != nil {
		t.Fatalf("Position error: %v", err)
	}

	if posUTF8 != (source.Position{Line: 0, Character: 3}) {
		t.Fatalf("Position(3, UTF8) = %v, want (0,3)", posUTF8)
	}
}

func TestLineIndexSurrogatePair(t *testing.T) {
	t.Parallel()

	content := "a\U0001F600b" // emoji: 4 UTF-8 bytes, 2 UTF-16 units, 1 UTF-32 unit
	idx := source.NewLineIndex(content)

	posBeforeB, err := idx.Position(5, source.UTF16)
	if err != nil {
		t.Fatalf("Position error: %v", err)
	}

	if posBeforeB != (source.Position{Line: 0, Character: 3}) {
		t.Fatalf("Position(5, UTF16) = %v, want (0,3) [1 + 2 surrogate units]", posBeforeB)
	}

	posUTF32, err := idx.Position(5, source.UTF32)
	if err != nil {
		t.Fatalf("Position error: %v", err)
	}

	if posUTF32 != (source.Position{Line: 0, Character: 2}) {
		t.Fatalf("Position(5, UTF32) = %v, want (0,2)", posUTF32)
	}

	_, err = idx.Offset(source.Position{Line: 0, Character: 2}, source.UTF16)
	if !errors.Is(err, source.ErrSurrogateSplit) {
		t.Fatalf("Offset at surrogate split: err = %v, want ErrSurrogateSplit", err)
	}

	before, err := idx.Offset(source.Position{Line: 0, Character: 1}, source.UTF16)
	if err != nil || before != 1 {
		t.Fatalf("Offset(char 1, UTF16) = (%d, %v), want (1, nil)", before, err)
	}

	after, err := idx.Offset(source.Position{Line: 0, Character: 3}, source.UTF16)
	if err != nil || after != 5 {
		t.Fatalf("Offset(char 3, UTF16) = (%d, %v), want (5, nil)", after, err)
	}
}

func TestLineIndexCombiningCharacters(t *testing.T) {
	t.Parallel()

	content := "\u0065\u0301x" // decomposed e + combining acute, not a grapheme cluster
	idx := source.NewLineIndex(content)

	pos, err := idx.Position(3, source.UTF32)
	if err != nil {
		t.Fatalf("Position error: %v", err)
	}

	if pos != (source.Position{Line: 0, Character: 2}) {
		t.Fatalf("Position(3, UTF32) = %v, want (0,2) [e, combining mark]", pos)
	}
}

func TestLineIndexEmptySpanAndEOF(t *testing.T) {
	t.Parallel()

	content := "abc"
	idx := source.NewLineIndex(content)

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	// Empty span at EOF.
	sp := source.Span{File: f, Start: 3, End: 3}
	if !sp.IsEmpty() {
		t.Fatalf("expected empty span")
	}

	rng, err := idx.Range(sp, source.UTF16)
	if err != nil {
		t.Fatalf("Range error: %v", err)
	}

	want := source.Range{Start: source.Position{Line: 0, Character: 3}, End: source.Position{Line: 0, Character: 3}}
	if rng != want {
		t.Fatalf("Range(EOF empty span) = %v, want %v", rng, want)
	}

	// Offset at len(content) is always valid, even though it's one past the
	// last byte.
	if !idx.ValidOffset(source.Offset(len(content))) {
		t.Fatalf("EOF offset should be valid")
	}
}

func TestLineIndexInvalidOffsets(t *testing.T) {
	t.Parallel()

	content := "héllo" // 'é' occupies bytes [1,3).
	idx := source.NewLineIndex(content)

	if _, err := idx.Position(-1, source.UTF16); !errors.Is(err, source.ErrOffsetOutOfRange) {
		t.Fatalf("Position(-1) err = %v, want ErrOffsetOutOfRange", err)
	}

	if _, err := idx.Position(source.Offset(len(content)+1), source.UTF16); !errors.Is(err, source.ErrOffsetOutOfRange) {
		t.Fatalf("Position(len+1) err = %v, want ErrOffsetOutOfRange", err)
	}

	if _, err := idx.Position(2, source.UTF16); !errors.Is(err, source.ErrInvalidUTF8Boundary) {
		t.Fatalf("Position(2) [mid-rune] err = %v, want ErrInvalidUTF8Boundary", err)
	}
}

func TestLineIndexColumnOutOfRange(t *testing.T) {
	t.Parallel()

	idx := source.NewLineIndex("abc\ndef")

	if _, err := idx.Offset(source.Position{Line: 0, Character: 100}, source.UTF8); !errors.Is(err, source.ErrColumnOutOfRange) {
		t.Fatalf("Offset(char 100) err = %v, want ErrColumnOutOfRange", err)
	}

	if _, err := idx.Offset(source.Position{Line: 0, Character: -1}, source.UTF8); !errors.Is(err, source.ErrColumnOutOfRange) {
		t.Fatalf("Offset(char -1) err = %v, want ErrColumnOutOfRange", err)
	}

	if _, err := idx.Offset(source.Position{Line: 5, Character: 0}, source.UTF8); !errors.Is(err, source.ErrLineOutOfRange) {
		t.Fatalf("Offset(line 5) err = %v, want ErrLineOutOfRange", err)
	}
}

func TestLineIndexLineAtBinarySearch(t *testing.T) {
	t.Parallel()

	var b []byte
	for range 1000 {
		b = append(b, []byte("0123456789\n")...)
	}

	idx := source.NewLineIndex(string(b))
	if idx.LineCount() != 1001 {
		t.Fatalf("LineCount() = %d, want 1001", idx.LineCount())
	}

	// Offset of the start of line 500 is 500*11.
	line, err := idx.LineAt(source.Offset(500 * 11))
	if err != nil || line != 500 {
		t.Fatalf("LineAt(500*11) = (%d, %v), want (500, nil)", line, err)
	}

	// One byte before that should be the previous line.
	line, err = idx.LineAt(source.Offset(500*11 - 1))
	if err != nil || line != 499 {
		t.Fatalf("LineAt(500*11-1) = (%d, %v), want (499, nil)", line, err)
	}
}

func TestLineIndexRangeSpanRoundTrip(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	content := "stock Add(a, b) {\n\treturn a + b;\n}\n"
	idx := source.NewLineIndex(content)

	sp, err := source.NewSpan(f, 6, 9) // "Add"
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	rng, err := idx.Range(sp, source.UTF16)
	if err != nil {
		t.Fatalf("Range error: %v", err)
	}

	back, err := idx.Span(f, rng, source.UTF16)
	if err != nil {
		t.Fatalf("Span error: %v", err)
	}

	if back != sp {
		t.Fatalf("round trip: Span(Range(sp)) = %v, want %v", back, sp)
	}
}

func TestLineIndexLineSpan(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	idx := source.NewLineIndex("abc\r\ndef")

	sp, err := idx.LineSpan(f, 0)
	if err != nil {
		t.Fatalf("LineSpan error: %v", err)
	}

	want := source.Span{File: f, Start: 0, End: 3}
	if sp != want {
		t.Fatalf("LineSpan(0) = %v, want %v", sp, want)
	}
}
