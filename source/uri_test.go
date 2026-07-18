package source_test

import (
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestURIIsValid(t *testing.T) {
	t.Parallel()

	var zero source.URI
	if zero.IsValid() {
		t.Fatalf("zero-value URI should be invalid")
	}

	if !source.URI("file:///a").IsValid() {
		t.Fatalf("non-empty URI should be valid")
	}
}

func TestFileURIUnixRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		uri  source.URI
	}{
		{"/home/user/project/gamemode.pwn", "file:///home/user/project/gamemode.pwn"},
		{"/tmp/a b/space.pwn", "file:///tmp/a%20b/space.pwn"},
		{"/tmp/héllo.inc", "file:///tmp/h%C3%A9llo.inc"},
	}

	for _, tc := range tests {
		got := source.FileURI(tc.path)
		if got != tc.uri {
			t.Errorf("FileURI(%q) = %q, want %q", tc.path, got, tc.uri)
		}

		back, err := got.Filename()
		if err != nil {
			t.Fatalf("Filename(%q) error: %v", got, err)
		}

		if back != tc.path {
			t.Errorf("round trip: Filename(FileURI(%q)) = %q", tc.path, back)
		}
	}
}

func TestFileURIWindowsDriveLetter(t *testing.T) {
	t.Parallel()

	got := source.FileURI(`C:\Users\dev\project\gamemode.pwn`)
	want := source.URI("file:///C:/Users/dev/project/gamemode.pwn")

	if got != want {
		t.Fatalf("FileURI(windows path) = %q, want %q", got, want)
	}

	back, err := got.Filename()
	if err != nil {
		t.Fatalf("Filename error: %v", err)
	}

	if back != "C:/Users/dev/project/gamemode.pwn" {
		t.Fatalf("Filename() = %q", back)
	}
}

func TestFilenameRejectsNonFileScheme(t *testing.T) {
	t.Parallel()

	_, err := source.URI("https://example.com/a.pwn").Filename()
	if err == nil {
		t.Fatalf("expected error for non-file URI")
	}
}

func TestFilenameRejectsInvalidURI(t *testing.T) {
	t.Parallel()

	_, err := source.URI("").Filename()
	if err == nil {
		t.Fatalf("expected error for empty URI")
	}
}

func TestFilenameRejectsUnparsableURI(t *testing.T) {
	t.Parallel()

	// A raw control character makes the URI unparsable by net/url.
	_, err := source.URI("file://\x7f").Filename()
	if err == nil {
		t.Fatalf("expected error for unparsable URI")
	}
}
