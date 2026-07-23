# Source and diagnostic type boundaries

PawnKit uses three kinds of source and diagnostic types. They look similar,
but they are not interchangeable.

## Core types

Use these inside a process when data can keep its PawnKit identity:

| Package | Types | Use |
|---|---|---|
| `source` | `FileID`, `URI`, `Offset`, `Span`, `Position`, `Range` | Files, byte spans, and editor positions |
| `textedit` | `Edit`, `DocumentEdit`, `WorkspaceEdit` | Validated edits tied to `source.Span` |
| `diagnostic` | `Diagnostic`, `RelatedLocation`, `Fix`, `Suppression` | Findings shared between PawnKit libraries |

`source.FileID` is process-local. Convert it to a URI before data leaves the
process.

## Core protocol types

The types in `protocol` are the JSON form of the core types. They remain
separate because they replace file IDs with URIs, make derived ranges
optional, and carry a schema version.

The version 1 shape is frozen in
`protocol/testdata/diagnostic_v1.json`. A wire change needs a new schema
version and fixture.

## Repository-specific types

Some owners need a different representation:

| Owner | Types | Classification |
|---|---|---|
| `pawn-parser` | Token positions, spans, parser diagnostics | Parser model |
| `pawn-analysis` | Preprocessor byte ranges and file indexes | Stage result; `ToCore` converts diagnostics |
| `pawnfmt` | Operation-local formatting range | Formatter API |
| `pawnlint` | Rule ranges and analyzer results | Lint model |
| `pawnlint/pkg/externalrule` | Request, diagnostics, and edits | Versioned external-rule protocol |
| `pawnlsp` | LSP positions, ranges, diagnostics, and edits | LSP wire protocol |
| `pawndebug` | DAP source locations | DAP wire protocol |
| `pawndoc` | Generated documentation locations | Documentation output |
| `pawntest` | Test-result locations | Test-report protocol |
| `pawnkit-cli` | SARIF locations | SARIF output |

These types should convert at their owner boundary. Do not alias a parser
offset, LSP position, DAP line, or SARIF region to a core type merely because
its fields have similar names.

## Adding a new type

Use a core type when its file identity and units match exactly. Keep a local
type when it belongs to a versioned protocol or has different units,
indexing, optional fields, or lifetime.

If a public JSON shape changes, add a version and a compatibility fixture in
the repository that owns the protocol.
