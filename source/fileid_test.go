package source_test

import (
	"testing"

	"github.com/pawnkit/pawnkit-core/source"
)

func TestFileIDZeroValueInvalid(t *testing.T) {
	t.Parallel()

	var id source.FileID
	if id.IsValid() {
		t.Fatalf("zero-value FileID should be invalid")
	}
}

func TestRegistryInternIsStable(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	a := reg.Intern("file:///a.pwn")
	b := reg.Intern("file:///b.pwn")
	aAgain := reg.Intern("file:///a.pwn")

	if a == b {
		t.Fatalf("distinct URIs got the same FileID")
	}

	if a != aAgain {
		t.Fatalf("interning the same URI twice produced different FileIDs: %v != %v", a, aAgain)
	}

	if !a.IsValid() || !b.IsValid() {
		t.Fatalf("interned FileIDs should be valid")
	}
}

func TestRegistryURIRoundTrip(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	uri := source.URI("file:///project/gamemode.pwn")

	id := reg.Intern(uri)

	got, ok := reg.URI(id)
	if !ok {
		t.Fatalf("URI(%v) not found", id)
	}

	if got != uri {
		t.Fatalf("URI(%v) = %q, want %q", id, got, uri)
	}
}

func TestRegistryURIUnknownID(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	if _, ok := reg.URI(source.FileID(999)); ok {
		t.Fatalf("expected unknown FileID to be not-found")
	}

	if _, ok := reg.URI(source.FileID(0)); ok {
		t.Fatalf("expected the zero FileID to be not-found")
	}
}

func TestRegistryLookup(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	if _, ok := reg.Lookup("file:///missing.pwn"); ok {
		t.Fatalf("expected Lookup to miss for un-interned URI")
	}

	id := reg.Intern("file:///present.pwn")

	got, ok := reg.Lookup("file:///present.pwn")
	if !ok || got != id {
		t.Fatalf("Lookup = (%v, %v), want (%v, true)", got, ok, id)
	}
}

func TestRegistryLen(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	if reg.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", reg.Len())
	}

	reg.Intern("file:///a.pwn")
	reg.Intern("file:///b.pwn")
	reg.Intern("file:///a.pwn") // duplicate, should not grow Len.

	if reg.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", reg.Len())
	}
}

func TestRegistryInternPanicsOnInvalidURI(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected Intern(\"\") to panic")
		}
	}()

	source.NewRegistry().Intern("")
}

func TestRegistryConcurrentIntern(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	const goroutines = 32

	done := make(chan source.FileID, goroutines)
	for i := range goroutines {
		go func(i int) {
			// All goroutines intern overlapping sets of URIs to exercise the
			// race detector and confirm identical URIs converge on one ID.
			uri := source.URI("file:///shared.pwn")
			if i%2 == 0 {
				uri = "file:///shared.pwn"
			}

			done <- reg.Intern(uri)
		}(i)
	}

	first := <-done

	for range goroutines - 1 {
		if id := <-done; id != first {
			t.Errorf("concurrent Intern produced divergent IDs: %v != %v", id, first)
		}
	}
}
