package protocol

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

// Interner assigns a source.FileID to a URI, interning it if necessary.
// *source.Registry satisfies this interface.
type Interner interface {
	Intern(uri source.URI) source.FileID
}

// DecodeDiagnostic decodes a diagnostic and interns its URIs.
// Locations require ByteSpan because source text is unavailable here.
func DecodeDiagnostic(interner Interner, wire Diagnostic) (diagnostic.Diagnostic, error) {
	if wire.SchemaVersion != 1 && wire.SchemaVersion != DiagnosticSchemaVersion {
		return diagnostic.Diagnostic{}, fmt.Errorf(
			"%w: got %d, support 1 and %d", ErrUnsupportedSchemaVersion, wire.SchemaVersion, DiagnosticSchemaVersion,
		)
	}

	severity, err := decodeSeverity(wire.Severity)
	if err != nil {
		return diagnostic.Diagnostic{}, err
	}

	primary, err := decodeLocation(interner, wire.Primary)
	if err != nil {
		return diagnostic.Diagnostic{}, fmt.Errorf("primary: %w", err)
	}

	d := diagnostic.New(wire.Code, wire.Source, severity, wire.Message, primary)
	d.Notes = wire.Notes
	d.Help = wire.Help
	d.DocsURL = wire.DocsURL

	for _, rl := range wire.Related {
		span, err := decodeLocation(interner, rl.Location)
		if err != nil {
			return diagnostic.Diagnostic{}, fmt.Errorf("related: %w", err)
		}

		d.Related = append(d.Related, diagnostic.RelatedLocation{Span: span, Message: rl.Message})
	}

	for _, tag := range wire.Tags {
		d.Tags = append(d.Tags, diagnostic.Tag(tag))
	}

	for _, wf := range wire.SafeFixes {
		fx, err := decodeFix(interner, wf)
		if err != nil {
			return diagnostic.Diagnostic{}, fmt.Errorf("safe fix: %w", err)
		}

		d.SafeFixes = append(d.SafeFixes, fx)
	}

	for _, wf := range wire.ReviewFixes {
		fx, err := decodeFix(interner, wf)
		if err != nil {
			return diagnostic.Diagnostic{}, fmt.Errorf("review fix: %w", err)
		}

		d.ReviewFixes = append(d.ReviewFixes, fx)
	}

	if wire.Suppressed != nil {
		d.Suppression = &diagnostic.Suppression{Kind: wire.Suppressed.Kind, Reason: wire.Suppressed.Reason}
	}

	return d, nil
}

func decodeSeverity(s string) (diagnostic.Severity, error) {
	switch s {
	case "error":
		return diagnostic.SeverityError, nil
	case "warning":
		return diagnostic.SeverityWarning, nil
	case "info":
		return diagnostic.SeverityInfo, nil
	case "hint":
		return diagnostic.SeverityHint, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrInvalidSeverity, s)
	}
}

func decodeFixKind(s string) (diagnostic.FixKind, error) {
	switch s {
	case "safe":
		return diagnostic.FixSafe, nil
	case "review-required":
		return diagnostic.FixReviewRequired, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrInvalidFixKind, s)
	}
}

func decodeLocation(interner Interner, loc Location) (source.Span, error) {
	if loc.Span == nil {
		return source.Span{}, ErrMissingSpan
	}

	file := interner.Intern(source.URI(loc.URI))

	return source.NewSpan(file, source.Offset(loc.Span.Start), source.Offset(loc.Span.End))
}

func decodeFix(interner Interner, wf Fix) (diagnostic.Fix, error) {
	kind, err := decodeFixKind(wf.Kind)
	if err != nil {
		return diagnostic.Fix{}, err
	}

	edit, err := decodeWorkspaceEdit(interner, wf.Edit)
	if err != nil {
		return diagnostic.Fix{}, err
	}

	return diagnostic.Fix{Message: wf.Message, Kind: kind, Edit: edit}, nil
}

func decodeWorkspaceEdit(interner Interner, wire WorkspaceEdit) (textedit.WorkspaceEdit, error) {
	result := textedit.WorkspaceEdit{}

	for _, wireDoc := range wire.Documents {
		file := interner.Intern(source.URI(wireDoc.URI))

		version := textedit.AnyVersion
		if wireDoc.Version != nil {
			version = *wireDoc.Version
		}

		doc := textedit.DocumentEdit{File: file, Version: version}

		for _, wireEdit := range wireDoc.Edits {
			sp, err := source.NewSpan(file, source.Offset(wireEdit.Span.Start), source.Offset(wireEdit.Span.End))
			if err != nil {
				return textedit.WorkspaceEdit{}, err
			}

			doc.Edits = append(doc.Edits, textedit.Edit{Span: sp, NewText: wireEdit.NewText})
		}

		result.Documents = append(result.Documents, doc)
	}

	return result, nil
}
