# Changelog

Notable changes to `pawnkit-core` are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

The module uses semantic version tags. Before 1.0, breaking changes are called
out explicitly and include migration notes.

## [0.5.0] - 2026-07-26

### Added

- Added immutable source buffers that share unchanged text across revisions.
- Added line indexes that update persistent buffers without flattening them.

## [0.4.0] - 2026-07-26

### Added

- Added byte-backed line indexes for editor buffers.

## [0.3.0] - 2026-07-26

### Added

- `LineIndex.Apply` updates an immutable line index after one replacement.

## [0.2.1] - 2026-07-25

### Changed

- Added the repository support record with CI validation.

## [0.2.0] - 2026-07-24

### Changed

- Emit version 2 diagnostics and read both published version 1 shapes.

### Migration

- Consumers that inspect diagnostic JSON must accept `schemaVersion: 2`.

## [0.1.1] - 2026-07-23

### Added

- Source and diagnostic type-boundary guidance.
- A compatibility check for the version 1 diagnostic JSON shape.

## [0.1.0] - 2026-07-18

### Added

- `source` types for file identity, byte spans, snapshots, and editor position
  conversion.
- `textedit` validation and application for document and workspace edits.
- `diagnostic` types for findings, related locations, fixes, and suppressions.
- Versioned diagnostic JSON encoding in `protocol`.
- Stable content and structured-value hashes in `hash`.
- Runnable examples for each public package.

[Unreleased]: https://github.com/pawnkit/pawnkit-core/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/pawnkit/pawnkit-core/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/pawnkit/pawnkit-core/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/pawnkit/pawnkit-core/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/pawnkit/pawnkit-core/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/pawnkit/pawnkit-core/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/pawnkit/pawnkit-core/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/pawnkit/pawnkit-core/releases/tag/v0.1.0
