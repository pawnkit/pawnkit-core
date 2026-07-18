package protocol

// Position is the JSON wire form of a line/character position. Character
// counts follow whatever encoding was requested when encoded (see
// [EncodeOptions.Encoding]); readers must agree on it out of band.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is the JSON wire form of a half-open line/character range.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// ByteSpan is the JSON wire form of a half-open byte range within one
// file's content.
type ByteSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Location is the JSON wire form of a source location.
type Location struct {
	URI string `json:"uri"`

	// Span is authoritative; Range is derived from it and may be omitted.
	Span *ByteSpan `json:"span,omitempty"`

	// Range is present only when the encoder had a line index available.
	Range *Range `json:"range,omitempty"`
}

// RelatedLocation is the JSON wire form of a diagnostic.RelatedLocation.
type RelatedLocation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// TextEdit is the JSON wire form of a textedit.Edit.
type TextEdit struct {
	Span    ByteSpan `json:"span"`
	NewText string   `json:"newText"`
}

// DocumentEdit is the JSON wire form of a textedit.DocumentEdit.
type DocumentEdit struct {
	URI string `json:"uri"`

	// Version is omitted when the edit was computed with textedit.AnyVersion.
	Version *int32 `json:"version,omitempty"`

	Edits []TextEdit `json:"edits"`
}

// WorkspaceEdit is the JSON wire form of a textedit.WorkspaceEdit.
type WorkspaceEdit struct {
	Documents []DocumentEdit `json:"documents"`
}

// Fix is the JSON wire form of a diagnostic.Fix.
type Fix struct {
	Message string `json:"message"`

	// Kind is "safe" or "review-required".
	Kind string `json:"kind"`

	Edit WorkspaceEdit `json:"edit"`
}

// Suppression is the JSON wire form of a diagnostic.Suppression.
type Suppression struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason,omitempty"`
}

// Diagnostic is the stable, versioned JSON wire form of a
// diagnostic.Diagnostic. Field names and SchemaVersion are part of the
// cross-repository wire contract; do not rename or repurpose without a
// SchemaVersion bump.
type Diagnostic struct {
	SchemaVersion int `json:"schemaVersion"`

	Code   string `json:"code"`
	Source string `json:"source"`

	// Severity is one of "error", "warning", "info", "hint".
	Severity string `json:"severity"`

	Message string   `json:"message"`
	Primary Location `json:"primary"`

	Related []RelatedLocation `json:"related,omitempty"`
	Notes   []string          `json:"notes,omitempty"`
	Help    string            `json:"help,omitempty"`
	DocsURL string            `json:"docsUrl,omitempty"`
	Tags    []string          `json:"tags,omitempty"`

	SafeFixes   []Fix `json:"safeFixes,omitempty"`
	ReviewFixes []Fix `json:"reviewFixes,omitempty"`

	Suppressed *Suppression `json:"suppressed,omitempty"`
}
