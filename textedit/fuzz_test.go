package textedit_test

import (
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

// FuzzApply asserts Apply never panics and, on success, the result length
// matches what the edits imply.
func FuzzApply(f *testing.F) {
	seeds := []struct {
		content string
		s1, e1  int
		text1   string
		s2, e2  int
		text2   string
	}{
		{"hello world", 0, 5, "hi", 6, 11, "pawn"},
		{"", 0, 0, "x", 0, 0, "y"},
		{"abc", 1, 1, "", 2, 2, ""},
		{"héllo", 0, 1, "", 0, 1, ""}, // deliberately overlapping.
		{"line1\nline2\r\n", 0, 5, "L1", 6, 11, "L2"},
	}

	for _, s := range seeds {
		f.Add(s.content, s.s1, s.e1, s.text1, s.s2, s.e2, s.text2)
	}

	f.Fuzz(func(t *testing.T, content string, s1, e1 int, text1 string, s2, e2 int, text2 string) {
		reg := source.NewRegistry()
		file := reg.Intern("file:///fuzz.pwn")

		mk := func(start, end int) (source.Span, bool) {
			if start < 0 || end < 0 || start > end {
				return source.Span{}, false
			}

			sp, err := source.NewSpan(file, source.Offset(start), source.Offset(end))
			if err != nil {
				return source.Span{}, false
			}

			return sp, true
		}

		var edits []textedit.Edit

		if sp, ok := mk(s1, e1); ok {
			edits = append(edits, textedit.Edit{Span: sp, NewText: text1})
		}

		if sp, ok := mk(s2, e2); ok {
			edits = append(edits, textedit.Edit{Span: sp, NewText: text2})
		}

		result, err := textedit.Apply(content, edits)
		if err != nil {
			return // invalid/overlapping/out-of-range input is expected to error, not panic.
		}

		expectedLen := len(content)
		for _, e := range edits {
			expectedLen += len(e.NewText) - int(e.Span.Len())
		}

		if len(result) != expectedLen {
			t.Fatalf(
				"Apply(%q, %+v) = %q (len %d), want len %d",
				content, edits, result, len(result), expectedLen,
			)
		}
	})
}
