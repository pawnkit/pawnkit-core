package protocol_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/pawnkit/pawnkit-core/protocol"
	"github.com/pawnkit/pawnkit-core/source"
)

// This fixture must never be updated to match code changes silently: a
// genuine wire-format change must bump DiagnosticSchemaVersion and add a
// new fixture, leaving this one documenting what version 1 looked like.
func TestDiagnosticV1FixtureIsStable(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/diagnostic_v1.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	var wire protocol.Diagnostic
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if wire.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", wire.SchemaVersion)
	}

	if wire.Code != "pawnlint:unused-variable" {
		t.Fatalf("Code = %q", wire.Code)
	}

	if wire.Severity != "warning" {
		t.Fatalf("Severity = %q", wire.Severity)
	}

	if wire.Primary.URI != "file:///gamemode/main.pwn" {
		t.Fatalf("Primary.URI = %q", wire.Primary.URI)
	}

	if wire.Primary.Span == nil || wire.Primary.Span.Start != 42 || wire.Primary.Span.End != 43 {
		t.Fatalf("Primary.Span = %+v, want {42 43}", wire.Primary.Span)
	}

	if wire.Primary.Range == nil || wire.Primary.Range.Start.Line != 3 || wire.Primary.Range.Start.Character != 8 {
		t.Fatalf("Primary.Range = %+v", wire.Primary.Range)
	}

	if len(wire.Related) != 1 || wire.Related[0].Message != "'x' was declared here" {
		t.Fatalf("Related = %+v", wire.Related)
	}

	if len(wire.Notes) != 1 {
		t.Fatalf("Notes = %+v", wire.Notes)
	}

	if wire.Help == "" || wire.DocsURL == "" {
		t.Fatalf("Help/DocsURL should be populated")
	}

	if len(wire.Tags) != 1 || wire.Tags[0] != "unnecessary" {
		t.Fatalf("Tags = %+v", wire.Tags)
	}

	if len(wire.SafeFixes) != 1 || wire.SafeFixes[0].Kind != "safe" {
		t.Fatalf("SafeFixes = %+v", wire.SafeFixes)
	}

	if len(wire.SafeFixes[0].Edit.Documents) != 1 {
		t.Fatalf("SafeFixes[0].Edit.Documents = %+v", wire.SafeFixes[0].Edit.Documents)
	}

	doc := wire.SafeFixes[0].Edit.Documents[0]
	if doc.Version == nil || *doc.Version != 7 {
		t.Fatalf("Version = %v, want 7", doc.Version)
	}

	if len(wire.ReviewFixes) != 0 {
		t.Fatalf("ReviewFixes = %+v, want empty", wire.ReviewFixes)
	}

	if wire.Suppressed != nil {
		t.Fatalf("Suppressed = %+v, want nil", wire.Suppressed)
	}

	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	var want, got map[string]any
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatalf("unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal encoded JSON: %v", err)
	}

	// Empty optional fields are omitted when encoded.
	delete(want, "reviewFixes")
	delete(want, "suppressed")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("encoded JSON no longer matches the version 1 fixture:\n got  %#v\n want %#v", got, want)
	}

	reg := source.NewRegistry()
	if _, err := protocol.DecodeDiagnostic(reg, wire); err != nil {
		t.Fatalf("DecodeDiagnostic(fixture) error: %v", err)
	}
}
