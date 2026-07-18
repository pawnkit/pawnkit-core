package diagnostic

import (
	"fmt"
	"slices"

	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

// RelatedLocation adds context at another source span.
type RelatedLocation struct {
	Span    source.Span
	Message string
}

// Validate checks the related span.
func (r RelatedLocation) Validate() error {
	if !r.Span.IsValid() {
		return fmt.Errorf("%w: related location", ErrInvalidPrimary)
	}

	return nil
}

// Suppression records why a diagnostic was suppressed.
type Suppression struct {
	Kind   string
	Reason string
}

// Validate checks the suppression kind.
func (s Suppression) Validate() error {
	if s.Kind == "" {
		return ErrInvalidSuppression
	}

	return nil
}

// Fix describes a diagnostic correction.
type Fix struct {
	Message string
	Kind    FixKind
	Edit    textedit.WorkspaceEdit
}

// Validate checks the fix kind and edits.
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

// Diagnostic describes a source finding. Its zero value is invalid.
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

// Validate checks all required fields and nested values.
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

// HasTag reports whether d contains tag.
func (d Diagnostic) HasTag(tag Tag) bool {
	return slices.Contains(d.Tags, tag)
}

// IsSuppressed reports whether d has suppression metadata.
func (d Diagnostic) IsSuppressed() bool {
	return d.Suppression != nil
}
