# Compatibility

`pawnkit-core` is pre-1.0. Public Go APIs may change between minor versions.

After 1.0, normal semantic versioning applies: patches fix bugs, minor releases add compatible APIs, and major releases may break callers.

## Go versions

The module supports the two most recent stable Go releases. CI tests the version in `go.mod` and the current stable release.

## Diagnostic JSON

`protocol.DiagnosticSchemaVersion` identifies the JSON contract. The current version is 1.

An optional field may be added without changing the schema version if existing fields keep the same meaning. Renaming, removing, repurposing, or changing the type of a field requires a new schema version.

`protocol.DecodeDiagnostic` rejects unsupported versions. It does not guess. During a future transition, core will keep a reader for at least the previous stable schema.

Each schema version has a frozen fixture under `protocol/testdata`. Add a new fixture for a new version; do not rewrite an old one.

## Diagnostic codes

The tool producing a diagnostic owns its `Code`. Core treats codes as opaque strings and does not reserve them. Once a tool publishes a code, it should not reuse that code for a different problem.

## Positions

`source.Span` and `textedit.Edit` use byte offsets. A line and character position is meaningful only with its `source.Encoding`, because UTF-8, UTF-16, and UTF-32 count characters differently.

Code exchanging editor positions must track or negotiate the encoding. LSP clients do this through `positionEncodingKind`.

## Downstream modules

PawnKit repositories version independently. A downstream module should pin a released `pawnkit-core` version that provides the APIs it uses; it should not assume every PawnKit repository shares one version number.
