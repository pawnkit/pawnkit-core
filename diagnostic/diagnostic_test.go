package diagnostic_test

import (
	"errors"
	"testing"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

func testSpan(t *testing.T, f source.FileID, start, end source.Offset) source.Span {
	t.Helper()

	sp, err := source.NewSpan(f, start, end)
	if err != nil {
		t.Fatalf("NewSpan error: %v", err)
	}

	return sp
}

func TestSeverityValidity(t *testing.T) {
	t.Parallel()

	var zero diagnostic.Severity
	if zero.IsValid() {
		t.Fatalf("zero-value Severity should be invalid")
	}

	for _, s := range []diagnostic.Severity{
		diagnostic.SeverityError,
		diagnostic.SeverityWarning,
		diagnostic.SeverityInfo,
		diagnostic.SeverityHint,
	} {
		if !s.IsValid() {
			t.Errorf("%v should be valid", s)
		}
	}
}

func TestFixKindValidity(t *testing.T) {
	t.Parallel()

	var zero diagnostic.FixKind
	if zero.IsValid() {
		t.Fatalf("zero-value FixKind should be invalid")
	}

	if !diagnostic.FixSafe.IsValid() || !diagnostic.FixReviewRequired.IsValid() {
		t.Fatalf("named FixKind constants should be valid")
	}
}

func TestNewAndValidate(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("pawnlint:unused-var", "pawnlint", diagnostic.SeverityWarning, "unused variable 'x'", testSpan(t, f, 0, 1))
	if err := d.Validate(); err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
}

func TestValidateRequiresCode(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("", "pawnlint", diagnostic.SeverityError, "msg", testSpan(t, f, 0, 1))
	if err := d.Validate(); !errors.Is(err, diagnostic.ErrMissingCode) {
		t.Fatalf("err = %v, want ErrMissingCode", err)
	}
}

func TestValidateRequiresSource(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "", diagnostic.SeverityError, "msg", testSpan(t, f, 0, 1))
	if err := d.Validate(); !errors.Is(err, diagnostic.ErrMissingSource) {
		t.Fatalf("err = %v, want ErrMissingSource", err)
	}
}

func TestValidateRequiresMessage(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "", testSpan(t, f, 0, 1))
	if err := d.Validate(); !errors.Is(err, diagnostic.ErrMissingMessage) {
		t.Fatalf("err = %v, want ErrMissingMessage", err)
	}
}

func TestValidateRequiresValidSeverity(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.Severity(0), "msg", testSpan(t, f, 0, 1))
	if err := d.Validate(); !errors.Is(err, diagnostic.ErrInvalidSeverity) {
		t.Fatalf("err = %v, want ErrInvalidSeverity", err)
	}
}

func TestValidateRequiresValidPrimary(t *testing.T) {
	t.Parallel()

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "msg", source.Span{})
	if err := d.Validate(); !errors.Is(err, diagnostic.ErrInvalidPrimary) {
		t.Fatalf("err = %v, want ErrInvalidPrimary", err)
	}
}

func TestValidateChecksRelatedLocations(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "msg", testSpan(t, f, 0, 1))
	d.Related = []diagnostic.RelatedLocation{{Span: source.Span{}, Message: "bad"}}

	if err := d.Validate(); !errors.Is(err, diagnostic.ErrInvalidPrimary) {
		t.Fatalf("err = %v, want ErrInvalidPrimary (from related location)", err)
	}
}

func TestValidateChecksFixes(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "msg", testSpan(t, f, 0, 1))
	d.SafeFixes = []diagnostic.Fix{{Message: "fix", Kind: diagnostic.FixSafe}} // no edits.

	if err := d.Validate(); !errors.Is(err, diagnostic.ErrInvalidFix) {
		t.Fatalf("err = %v, want ErrInvalidFix", err)
	}

	d.SafeFixes[0].Edit = textedit.WorkspaceEdit{
		Documents: []textedit.DocumentEdit{
			{File: f, Version: 1, Edits: []textedit.Edit{{Span: testSpan(t, f, 0, 1), NewText: "x"}}},
		},
	}

	if err := d.Validate(); err != nil {
		t.Fatalf("Validate() with a valid fix: unexpected error %v", err)
	}
}

func TestValidateChecksFixKind(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "msg", testSpan(t, f, 0, 1))
	d.ReviewFixes = []diagnostic.Fix{{
		Message: "fix",
		Kind:    diagnostic.FixKind(0),
		Edit: textedit.WorkspaceEdit{Documents: []textedit.DocumentEdit{
			{File: f, Edits: []textedit.Edit{{Span: testSpan(t, f, 0, 1), NewText: "x"}}},
		}},
	}}

	if err := d.Validate(); !errors.Is(err, diagnostic.ErrInvalidFix) {
		t.Fatalf("err = %v, want ErrInvalidFix", err)
	}
}

func TestValidateChecksSuppression(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.SeverityError, "msg", testSpan(t, f, 0, 1))
	d.Suppression = &diagnostic.Suppression{Kind: ""}

	if err := d.Validate(); !errors.Is(err, diagnostic.ErrInvalidSuppression) {
		t.Fatalf("err = %v, want ErrInvalidSuppression", err)
	}

	d.Suppression = &diagnostic.Suppression{Kind: "inline-comment", Reason: "false positive"}
	if err := d.Validate(); err != nil {
		t.Fatalf("Validate() with valid suppression: unexpected error %v", err)
	}

	if !d.IsSuppressed() {
		t.Fatalf("IsSuppressed() should be true")
	}
}

func TestHasTag(t *testing.T) {
	t.Parallel()

	reg := source.NewRegistry()
	f := reg.Intern("file:///a.pwn")

	d := diagnostic.New("code", "tool", diagnostic.SeverityWarning, "msg", testSpan(t, f, 0, 1))
	d.Tags = []diagnostic.Tag{diagnostic.TagDeprecated}

	if !d.HasTag(diagnostic.TagDeprecated) {
		t.Fatalf("expected TagDeprecated")
	}

	if d.HasTag(diagnostic.TagUnnecessary) {
		t.Fatalf("did not expect TagUnnecessary")
	}
}

func TestZeroValueDiagnosticIsInvalid(t *testing.T) {
	t.Parallel()

	var d diagnostic.Diagnostic
	if err := d.Validate(); err == nil {
		t.Fatalf("zero-value Diagnostic should fail Validate")
	}
}
