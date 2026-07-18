package hash_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pawnkit/pawnkit-core/hash"
)

func TestContentIsStableAndPrefixed(t *testing.T) {
	t.Parallel()

	a := hash.Content([]byte("hello"))
	b := hash.Content([]byte("hello"))

	if a != b {
		t.Fatalf("Content is not stable: %q != %q", a, b)
	}

	if !strings.HasPrefix(a, "sha256:") {
		t.Fatalf("Content() = %q, want sha256: prefix", a)
	}
}

func TestContentDiffersForDifferentInput(t *testing.T) {
	t.Parallel()

	a := hash.Content([]byte("hello"))
	b := hash.Content([]byte("world"))

	if a == b {
		t.Fatalf("expected different hashes for different input")
	}
}

func TestContentEmptyInput(t *testing.T) {
	t.Parallel()

	got := hash.Content(nil)
	if !strings.HasPrefix(got, "sha256:") {
		t.Fatalf("Content(nil) = %q", got)
	}

	// Known SHA-256 of the empty string.
	want := "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Fatalf("Content(nil) = %q, want %q", got, want)
	}
}

func TestStringMatchesContent(t *testing.T) {
	t.Parallel()

	if hash.String("abc") != hash.Content([]byte("abc")) {
		t.Fatalf("String and Content disagree")
	}
}

func TestReader(t *testing.T) {
	t.Parallel()

	got, err := hash.Reader(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("Reader error: %v", err)
	}

	if got != hash.String("hello") {
		t.Fatalf("Reader() = %q, want %q", got, hash.String("hello"))
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

func TestReaderPropagatesError(t *testing.T) {
	t.Parallel()

	if _, err := hash.Reader(errReader{}); err == nil {
		t.Fatalf("expected an error from a failing reader")
	}
}

func TestJSONIsStableForEqualValues(t *testing.T) {
	t.Parallel()

	type config struct {
		Name    string
		Version int
	}

	a, err := hash.JSON(config{Name: "pawnlint", Version: 1})
	if err != nil {
		t.Fatalf("JSON error: %v", err)
	}

	b, err := hash.JSON(config{Name: "pawnlint", Version: 1})
	if err != nil {
		t.Fatalf("JSON error: %v", err)
	}

	if a != b {
		t.Fatalf("JSON hash not stable: %q != %q", a, b)
	}

	c, err := hash.JSON(config{Name: "pawnlint", Version: 2})
	if err != nil {
		t.Fatalf("JSON error: %v", err)
	}

	if a == c {
		t.Fatalf("expected different hashes for different config values")
	}
}

func TestJSONPropagatesMarshalError(t *testing.T) {
	t.Parallel()

	// Channels cannot be marshaled to JSON.
	if _, err := hash.JSON(make(chan int)); err == nil {
		t.Fatalf("expected an error marshaling an unsupported type")
	}
}

func TestCombineIsOrderSensitive(t *testing.T) {
	t.Parallel()

	a := hash.Combine("one", "two")
	b := hash.Combine("two", "one")

	if a == b {
		t.Fatalf("Combine should be order-sensitive")
	}
}

func TestCombineAvoidsConcatenationAmbiguity(t *testing.T) {
	t.Parallel()

	a := hash.Combine("ab", "c")
	b := hash.Combine("a", "bc")

	if a == b {
		t.Fatalf("Combine(\"ab\",\"c\") should differ from Combine(\"a\",\"bc\")")
	}
}

func TestCombineIsStable(t *testing.T) {
	t.Parallel()

	a := hash.Combine("content-hash", "config-hash", "v1.2.3")
	b := hash.Combine("content-hash", "config-hash", "v1.2.3")

	if a != b {
		t.Fatalf("Combine is not stable: %q != %q", a, b)
	}
}
