package source_test

import (
	"strings"
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

// FuzzLineIndexPosition asserts Position never panics and, on success,
// round-trips back to the same offset via Offset.
func FuzzLineIndexPosition(f *testing.F) {
	seeds := []struct {
		content string
		offset  int
	}{
		{"", 0},
		{"hello world", 5},
		{"line one\nline two\r\nline three", 12},
		{"héllo\nworld", 3},
		{"a\U0001F600b", 5},
		{"\r\n\r\n\r\n", 4},
		{"éx", 3}, // combining character.
		{"trailing newline\n", 17},
	}
	for _, s := range seeds {
		f.Add(s.content, s.offset)
	}

	f.Fuzz(func(t *testing.T, content string, offset int) {
		idx := source.NewLineIndex(content)

		for _, enc := range []source.Encoding{source.UTF8, source.UTF16, source.UTF32} {
			pos, err := idx.Position(source.Offset(offset), enc)
			if err != nil {
				continue // invalid offsets are expected for arbitrary fuzz input.
			}

			if pos.Line < 0 || pos.Line >= idx.LineCount() {
				t.Fatalf("Position(%d, %v) = %v: line out of range", offset, enc, pos)
			}

			if pos.Character < 0 {
				t.Fatalf("Position(%d, %v) = %v: negative character", offset, enc, pos)
			}

			back, err := idx.Offset(pos, enc)
			if err != nil {
				t.Fatalf("Offset(%v, %v) failed after successful Position(): %v", pos, enc, err)
			}

			if int(back) != offset {
				t.Fatalf("round trip mismatch: offset %d -> %v -> %d (enc %v)", offset, pos, back, enc)
			}

			if !idx.ValidOffset(back) {
				t.Fatalf("round-tripped offset %d is not a valid offset", back)
			}
		}
	})
}

// FuzzFileURIFilename asserts the FileURI/Filename round trip never panics.
func FuzzFileURIFilename(f *testing.F) {
	seeds := []string{
		"/home/user/project/gamemode.pwn",
		`C:\Users\dev\project.pwn`,
		"/tmp/a b/space file.inc",
		"/tmp/héllo.inc",
		"",
		"relative/path.pwn",
		"/",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, path string) {
		uri := source.FileURI(path)
		if !uri.IsValid() {
			t.Fatalf("FileURI(%q) produced an invalid URI", path)
		}

		back, err := uri.Filename()
		if err != nil {
			t.Fatalf("Filename(%q) failed for URI built by FileURI: %v", uri, err)
		}

		// The normalized form must itself be stable under a second
		// round trip.
		again := source.FileURI(back)

		backAgain, err := again.Filename()
		if err != nil {
			t.Fatalf("second round trip failed: %v", err)
		}

		if backAgain != strings.ReplaceAll(back, `\`, "/") {
			t.Fatalf("round trip not stable: %q -> %q -> %q", path, back, backAgain)
		}
	})
}
