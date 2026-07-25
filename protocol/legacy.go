package protocol

import (
	"encoding/json"
	"strconv"
)

type diagnosticAlias Diagnostic

type legacyPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

type legacyLocation struct {
	File  string         `json:"file"`
	Start legacyPosition `json:"start"`
	End   legacyPosition `json:"end"`
}

type legacyRelatedLocation struct {
	File    string         `json:"file"`
	Start   legacyPosition `json:"start"`
	End     legacyPosition `json:"end"`
	Message string         `json:"message"`
}

type legacyEdit struct {
	File  string `json:"file"`
	Range struct {
		Start legacyPosition `json:"start"`
		End   legacyPosition `json:"end"`
	} `json:"range"`
	NewText string          `json:"newText"`
	Version json.RawMessage `json:"version"`
}

type legacySuppression struct {
	Suppressed bool   `json:"suppressed"`
	Reason     string `json:"reason"`
	Mechanism  string `json:"mechanism"`
}

type legacyDiagnostic struct {
	Code             string                  `json:"code"`
	Source           string                  `json:"source"`
	Severity         string                  `json:"severity"`
	Message          string                  `json:"message"`
	Range            legacyLocation          `json:"range"`
	RelatedLocations []legacyRelatedLocation `json:"relatedLocations"`
	Notes            []string                `json:"notes"`
	Help             string                  `json:"help"`
	DocumentationURL string                  `json:"documentationUrl"`
	Tags             []string                `json:"tags"`
	Fixes            []legacyEdit            `json:"fixes"`
	UnsafeFixes      []legacyEdit            `json:"unsafeFixes"`
	Suppression      *legacySuppression      `json:"suppression"`
}

// UnmarshalJSON accepts both published version 1 shapes and version 2.
func (d *Diagnostic) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, oldSchema := fields["range"]; !oldSchema {
		var current diagnosticAlias
		if err := json.Unmarshal(data, &current); err != nil {
			return err
		}
		*d = Diagnostic(current)
		return nil
	}

	var old legacyDiagnostic
	if err := json.Unmarshal(data, &old); err != nil {
		return err
	}
	*d = Diagnostic{
		SchemaVersion: 1,
		Code:          old.Code,
		Source:        old.Source,
		Severity:      old.Severity,
		Message:       old.Message,
		Primary:       legacyWireLocation(old.Range),
		Notes:         old.Notes,
		Help:          old.Help,
		DocsURL:       old.DocumentationURL,
		Tags:          old.Tags,
	}
	for _, related := range old.RelatedLocations {
		d.Related = append(d.Related, RelatedLocation{
			Location: legacyWireLocation(legacyLocation{
				File: related.File, Start: related.Start, End: related.End,
			}),
			Message: related.Message,
		})
	}
	for _, edit := range old.Fixes {
		d.SafeFixes = append(d.SafeFixes, legacyWireFix(edit, "safe"))
	}
	for _, edit := range old.UnsafeFixes {
		d.ReviewFixes = append(d.ReviewFixes, legacyWireFix(edit, "review-required"))
	}
	if old.Suppression != nil && old.Suppression.Suppressed {
		kind := old.Suppression.Mechanism
		if kind == "" {
			kind = "legacy"
		}
		d.Suppressed = &Suppression{Kind: kind, Reason: old.Suppression.Reason}
	}
	return nil
}

func legacyWireLocation(location legacyLocation) Location {
	return Location{
		URI:  location.File,
		Span: &ByteSpan{Start: location.Start.Offset, End: location.End.Offset},
		Range: &Range{
			Start: Position{Line: zeroBased(location.Start.Line), Character: zeroBased(location.Start.Column)},
			End:   Position{Line: zeroBased(location.End.Line), Character: zeroBased(location.End.Column)},
		},
	}
}

func legacyWireFix(edit legacyEdit, kind string) Fix {
	document := DocumentEdit{
		URI: edit.File,
		Edits: []TextEdit{{
			Span:    ByteSpan{Start: edit.Range.Start.Offset, End: edit.Range.End.Offset},
			NewText: edit.NewText,
		}},
	}
	document.Version = legacyWireVersion(edit.Version)
	return Fix{
		Message: "Apply fix",
		Kind:    kind,
		Edit:    WorkspaceEdit{Documents: []DocumentEdit{document}},
	}
}

func legacyWireVersion(raw json.RawMessage) *int32 {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var number int32
	if json.Unmarshal(raw, &number) == nil {
		return &number
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return nil
	}
	value, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return nil
	}
	number = int32(value)
	return &number
}

func zeroBased(value int) int {
	if value > 0 {
		return value - 1
	}
	return 0
}
