# Changelog

Notable changes to `pawnkit-core` are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

The module uses semantic version tags. Before 1.0, breaking changes are called
out explicitly and include migration notes.

## [Unreleased]

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

[Unreleased]: https://github.com/pawnkit/pawnkit-core/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/pawnkit/pawnkit-core/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/pawnkit/pawnkit-core/releases/tag/v0.1.0
