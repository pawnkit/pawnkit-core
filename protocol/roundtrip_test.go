package protocol_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/protocol"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func buildFullDiagnostic(t *testing.T, reg *source.Registry) (diagnostic.Diagnostic, source.FileID, source.FileID) {
	t.Helper()

	main := reg.Intern("file:///gamemode/main.pwn")
	other := reg.Intern("file:///gamemode/other.pwn")

	primary, err := source.NewSpan(main, 10, 20)
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	relatedSpan, err := source.NewSpan(other, 0, 5)
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	fixSpan, err := source.NewSpan(main, 10, 20)
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	d := diagnostic.New("pawnlint:deprecated-func", "pawnlint", diagnostic.SeverityWarning, "SetPlayerColor is deprecated", primary)
	d.Related = []diagnostic.RelatedLocation{{Span: relatedSpan, Message: "similar usage here"}}
	d.Notes = []string{"consider using SetPlayerColour instead"}
	d.Help = "replace the call with the recommended alternative"
	d.DocsURL = "https://pawnkit.dev/rules/deprecated-func"
	d.Tags = []diagnostic.Tag{diagnostic.TagDeprecated}
	d.SafeFixes = []diagnostic.Fix{{
		Message: "Replace with SetPlayerColour",
		Kind:    diagnostic.FixSafe,
		Edit: textedit.WorkspaceEdit{Documents: []textedit.DocumentEdit{
			{File: main, Version: 3, Edits: []textedit.Edit{{Span: fixSpan, NewText: "SetPlayerColour"}}},
		}},
	}}
	d.ReviewFixes = []diagnostic.Fix{{
		Message: "Remove the call entirely",
		Kind:    diagnostic.FixReviewRequired,
		Edit: textedit.WorkspaceEdit{Documents: []textedit.DocumentEdit{
			{File: main, Version: textedit.AnyVersion, Edits: []textedit.Edit{{Span: fixSpan, NewText: ""}}},
		}},
	}}
	d.Suppression = &diagnostic.Suppression{Kind: "inline-comment", Reason: "legacy code, migrating later"}

	if err := d.Validate(); err != nil {
		t.Fatalf("test diagnostic is invalid: %v", err)
	}

	return d, main, other
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	original, _, _ := buildFullDiagnostic(t, reg)

	wire, err := protocol.EncodeDiagnostic(protocol.EncodeOptions{Resolver: reg}, original)
	if err != nil {
		t.Fatalf("EncodeDiagnostic error: %v", err)
	}

	if wire.SchemaVersion != protocol.DiagnosticSchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", wire.SchemaVersion, protocol.DiagnosticSchemaVersion)
	}

	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var decodedWire protocol.Diagnostic
	if err := json.Unmarshal(data, &decodedWire); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	back, err := protocol.DecodeDiagnostic(reg, decodedWire)
	if err != nil {
		t.Fatalf("DecodeDiagnostic error: %v", err)
	}

	if !reflect.DeepEqual(original, back) {
		t.Fatalf("round trip mismatch:\n original = %#v\n back     = %#v", original, back)
	}
}

func TestEncodeDecodeRoundTripAcrossProcesses(t *testing.T) {
	t.Parallel()

	// Encode in registry A, decode in registry B: only URIs cross the boundary.
	regA := source.NewRegistry()
	original, mainA, otherA := buildFullDiagnostic(t, regA)

	wire, err := protocol.EncodeDiagnostic(protocol.EncodeOptions{Resolver: regA}, original)
	if err != nil {
		t.Fatalf("EncodeDiagnostic error: %v", err)
	}

	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decodedWire protocol.Diagnostic
	if err := json.Unmarshal(data, &decodedWire); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	regB := source.NewRegistry()
	regB.Intern("file:///decoy.pwn") // makes FileID numbering diverge from regA

	back, err := protocol.DecodeDiagnostic(regB, decodedWire)
	if err != nil {
		t.Fatalf("DecodeDiagnostic error: %v", err)
	}

	mainB, ok := regB.Lookup("file:///gamemode/main.pwn")
	if !ok {
		t.Fatalf("expected main.pwn to be interned in regB")
	}

	if back.Primary.File != mainB {
		t.Fatalf("Primary.File = %v, want %v", back.Primary.File, mainB)
	}

	if mainA == mainB {
		t.Fatalf("test is meaningless if FileIDs happened to coincide; adjust the decoy interning")
	}

	otherB, ok := regB.Lookup("file:///gamemode/other.pwn")
	if !ok || back.Related[0].Span.File != otherB {
		t.Fatalf("related location did not resolve through regB correctly")
	}

	_ = otherA
}

func TestEncodeDecodeWithLineIndex(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	content := "stock Add(a, b) {\n\treturn a + b;\n}\n"
	idx := source.NewLineIndex(content)

	sp, err := source.NewSpan(f, 6, 9) // "Add"
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	d := diagnostic.New("pawnlint:naming", "pawnlint", diagnostic.SeverityHint, "consider PascalCase", sp)

	opts := protocol.EncodeOptions{
		Resolver: reg,
		Lines: func(file source.FileID) (*source.LineIndex, bool) {
			if file != f {
				return nil, false
			}

			return idx, true
		},
		Encoding: source.UTF16,
	}

	wire, err := protocol.EncodeDiagnostic(opts, d)
	if err != nil {
		t.Fatalf("EncodeDiagnostic error: %v", err)
	}

	if wire.Primary.Range == nil {
		t.Fatalf("expected Range to be populated")
	}

	want := protocol.Range{Start: protocol.Position{Line: 0, Character: 6}, End: protocol.Position{Line: 0, Character: 9}}
	if *wire.Primary.Range != want {
		t.Fatalf("Range = %+v, want %+v", *wire.Primary.Range, want)
	}

	if wire.Primary.Span == nil || wire.Primary.Span.Start != 6 || wire.Primary.Span.End != 9 {
		t.Fatalf("Span = %+v, want {6 9}", wire.Primary.Span)
	}
}

func TestDecodeRejectsMissingSpan(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	wire := protocol.Diagnostic{
		SchemaVersion: protocol.DiagnosticSchemaVersion,
		Code:          "x",
		Source:        "tool",
		Severity:      "error",
		Message:       "msg",
		Primary:       protocol.Location{URI: "file:///a.pwn"}, // no Span.
	}

	if _, err := protocol.DecodeDiagnostic(reg, wire); err == nil {
		t.Fatalf("expected an error decoding a Location without a Span")
	}
}

func TestDecodeRejectsUnsupportedSchemaVersion(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	wire := protocol.Diagnostic{SchemaVersion: 999}

	if _, err := protocol.DecodeDiagnostic(reg, wire); err == nil {
		t.Fatalf("expected an error for an unsupported schema version")
	}
}

func TestDecodeRejectsInvalidSeverity(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()

	wire := protocol.Diagnostic{
		SchemaVersion: protocol.DiagnosticSchemaVersion,
		Severity:      "critical", // not a known severity.
		Primary:       protocol.Location{URI: "file:///a.pwn", Span: &protocol.ByteSpan{Start: 0, End: 1}},
	}

	if _, err := protocol.DecodeDiagnostic(reg, wire); err == nil {
		t.Fatalf("expected an error for an invalid severity")
	}
}

func TestEncodeRejectsUnknownFile(t *testing.T) {
	t.Parallel()

	regA := source.NewRegistry()
	f := regA.Intern("file:///a.pwn")

	sp, err := source.NewSpan(f, 0, 1)
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "msg", sp)

	// A different, empty registry cannot resolve f.
	regB := source.NewRegistry()

	if _, err := protocol.EncodeDiagnostic(protocol.EncodeOptions{Resolver: regB}, d); err == nil {
		t.Fatalf("expected an error for an unresolvable file id")
	}
}
