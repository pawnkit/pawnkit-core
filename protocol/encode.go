package protocol

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/source"
	"github.com/pawnkit/pawnkit-core/textedit"
)

// Resolver supplies the URI a source.FileID was interned from.
// *source.Registry satisfies this interface.
type Resolver interface {
	URI(file source.FileID) (source.URI, bool)
}

// LineIndexLookup optionally supplies the source.LineIndex for a file, so
// Encode can populate each Location's Range in addition to its byte Span.
// A nil LineIndexLookup still leaves Diagnostics fully round-trippable
// through byte spans alone.
type LineIndexLookup func(file source.FileID) (*source.LineIndex, bool)

// EncodeOptions configures Encode.
type EncodeOptions struct {
	// Resolver is required.
	Resolver Resolver

	// Lines is optional; when set, Encode also populates each Location's
	// Range using Encoding.
	Lines LineIndexLookup

	// Encoding has no effect if Lines is nil. The zero value is source.UTF8.
	Encoding source.Encoding
}

// EncodeDiagnostic converts d into its stable JSON wire form, resolving
// every source.FileID it references through opts.Resolver.
func EncodeDiagnostic(opts EncodeOptions, d diagnostic.Diagnostic) (Diagnostic, error) {
	primary, err := encodeLocation(opts, d.Primary)
	if err != nil {
		return Diagnostic{}, fmt.Errorf("primary: %w", err)
	}

	wire := Diagnostic{
		SchemaVersion: DiagnosticSchemaVersion,
		Code:          d.Code,
		Source:        d.Source,
		Severity:      d.Severity.String(),
		Message:       d.Message,
		Primary:       primary,
		Notes:         d.Notes,
		Help:          d.Help,
		DocsURL:       d.DocsURL,
	}

	for _, rl := range d.Related {
		loc, err := encodeLocation(opts, rl.Span)
		if err != nil {
			return Diagnostic{}, fmt.Errorf("related: %w", err)
		}

		wire.Related = append(wire.Related, RelatedLocation{Location: loc, Message: rl.Message})
	}

	for _, tag := range d.Tags {
		wire.Tags = append(wire.Tags, string(tag))
	}

	for _, fx := range d.SafeFixes {
		wf, err := encodeFix(opts, fx)
		if err != nil {
			return Diagnostic{}, fmt.Errorf("safe fix: %w", err)
		}

		wire.SafeFixes = append(wire.SafeFixes, wf)
	}

	for _, fx := range d.ReviewFixes {
		wf, err := encodeFix(opts, fx)
		if err != nil {
			return Diagnostic{}, fmt.Errorf("review fix: %w", err)
		}

		wire.ReviewFixes = append(wire.ReviewFixes, wf)
	}

	if d.Suppression != nil {
		wire.Suppressed = &Suppression{Kind: d.Suppression.Kind, Reason: d.Suppression.Reason}
	}

	return wire, nil
}

func encodeLocation(opts EncodeOptions, span source.Span) (Location, error) {
	uri, ok := opts.Resolver.URI(span.File)
	if !ok {
		return Location{}, fmt.Errorf("%w: %v", ErrUnknownFile, span.File)
	}

	loc := Location{
		URI:  string(uri),
		Span: &ByteSpan{Start: int(span.Start), End: int(span.End)},
	}

	if opts.Lines != nil {
		if idx, ok := opts.Lines(span.File); ok {
			rng, err := idx.Range(span, opts.Encoding)
			if err != nil {
				return Location{}, fmt.Errorf("computing range for %v: %w", span, err)
			}

			loc.Range = &Range{
				Start: Position{Line: rng.Start.Line, Character: rng.Start.Character},
				End:   Position{Line: rng.End.Line, Character: rng.End.Character},
			}
		}
	}

	return loc, nil
}

func encodeFix(opts EncodeOptions, fx diagnostic.Fix) (Fix, error) {
	edit, err := encodeWorkspaceEdit(opts, fx.Edit)
	if err != nil {
		return Fix{}, err
	}

	return Fix{Message: fx.Message, Kind: fx.Kind.String(), Edit: edit}, nil
}

func encodeWorkspaceEdit(opts EncodeOptions, we textedit.WorkspaceEdit) (WorkspaceEdit, error) {
	wire := WorkspaceEdit{}

	for _, doc := range we.Documents {
		uri, ok := opts.Resolver.URI(doc.File)
		if !ok {
			return WorkspaceEdit{}, fmt.Errorf("%w: %v", ErrUnknownFile, doc.File)
		}

		wireDoc := DocumentEdit{URI: string(uri)}

		if doc.Version != textedit.AnyVersion {
			v := doc.Version
			wireDoc.Version = &v
		}

		for _, e := range doc.Edits {
			wireDoc.Edits = append(wireDoc.Edits, TextEdit{
				Span:    ByteSpan{Start: int(e.Span.Start), End: int(e.Span.End)},
				NewText: e.NewText,
			})
		}

		wire.Documents = append(wire.Documents, wireDoc)
	}

	return wire, nil
}
