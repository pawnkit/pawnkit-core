package diagnostic

import (
	"fmt"
	"slices"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

type RelatedLocation struct {
	Span    source.Span
	Message string
}

func (r RelatedLocation) Validate() error {
	if !r.Span.IsValid() {
		return fmt.Errorf("%w: related location", ErrInvalidPrimary)
	}

	return nil
}

type Suppression struct {
	Kind   string
	Reason string
}

func (s Suppression) Validate() error {
	if s.Kind == "" {
		return ErrInvalidSuppression
	}

	return nil
}

type Fix struct {
	Message string
	Kind    FixKind
	Edit    textedit.WorkspaceEdit
}

func (f Fix) Validate() error {
	if !f.Kind.IsValid() {
		return fmt.Errorf("%w: invalid kind", ErrInvalidFix)
	}

	if len(f.Edit.Documents) == 0 {
		return fmt.Errorf("%w: no document edits", ErrInvalidFix)
	}

	for _, doc := range f.Edit.Documents {
		if err := textedit.ValidateEdits(doc.Edits); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidFix, err)
		}
	}

	return nil
}

// Zero value is invalid; use New then Validate.
type Diagnostic struct {
	Code   string // stable, never repurposed, e.g. "pawnlint:unused-variable"
	Source string

	Severity Severity
	Message  string

	Primary source.Span

	Related []RelatedLocation
	Notes   []string

	Help    string
	DocsURL string

	Tags []Tag

	SafeFixes   []Fix
	ReviewFixes []Fix

	Suppression *Suppression
}

// New does not validate; call Validate once the Diagnostic is populated.
func New(code, src string, severity Severity, message string, primary source.Span) Diagnostic {
	return Diagnostic{
		Code:     code,
		Source:   src,
		Severity: severity,
		Message:  message,
		Primary:  primary,
	}
}

func (d Diagnostic) Validate() error {
	if d.Code == "" {
		return ErrMissingCode
	}

	if d.Source == "" {
		return ErrMissingSource
	}

	if d.Message == "" {
		return ErrMissingMessage
	}

	if !d.Severity.IsValid() {
		return fmt.Errorf("%w: %v", ErrInvalidSeverity, d.Severity)
	}

	if !d.Primary.IsValid() {
		return ErrInvalidPrimary
	}

	for i, r := range d.Related {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("related[%d]: %w", i, err)
		}
	}

	for i, fx := range d.SafeFixes {
		if err := fx.Validate(); err != nil {
			return fmt.Errorf("safeFixes[%d]: %w", i, err)
		}
	}

	for i, fx := range d.ReviewFixes {
		if err := fx.Validate(); err != nil {
			return fmt.Errorf("reviewFixes[%d]: %w", i, err)
		}
	}

	if d.Suppression != nil {
		if err := d.Suppression.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (d Diagnostic) HasTag(tag Tag) bool {
	return slices.Contains(d.Tags, tag)
}

func (d Diagnostic) IsSuppressed() bool {
	return d.Suppression != nil
}
