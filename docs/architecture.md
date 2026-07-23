# Architecture

`pawnkit-core` sits at the bottom of the PawnKit Go dependency graph. It defines the source, edit, and diagnostic types used by higher-level tools, but it does not import those tools.

```text
pawnkit-core
    ^
    |
pawn-parser, pawn-analysis, pawn-project, pawnfmt, pawnlint, pawnlsp, ...
```

`pawnkit-spec` owns shared formats and language contracts. `pawnkit-core` owns the Go primitives used to implement them. Neither repository depends on the other.

## Packages

| Package | Responsibility |
|---|---|
| `source` | File identity, snapshots, byte spans, line indexes, and position conversion |
| `textedit` | File and workspace edits, overlap checks, and application |
| `diagnostic` | Findings, related locations, fixes, tags, and suppressions |
| `protocol` | Versioned JSON encoding for diagnostics |
| `hash` | Stable content and cache-key hashes |

## How data moves

1. A tool reads a file and creates a `source.Snapshot`.
2. Parser or analysis code records locations as byte-based `source.Span` values.
3. Findings use `diagnostic.Diagnostic`; fixes use `textedit.WorkspaceEdit`.
4. `protocol` replaces process-local file IDs with URIs before serialization.
5. A CLI or editor host validates and applies the edit, then decides how to write it.

The last step belongs to the caller. Core computes new content but does not write files or contact an editor.

## Design choices

### Byte offsets inside the process

Byte offsets do not depend on an editor's character encoding. `source.LineIndex` converts them to UTF-8, UTF-16, or UTF-32 positions when needed.

### URIs across process boundaries

`source.FileID` is a small, process-local handle. It is useful as a map key but meaningless in another process. The protocol package serializes a `source.URI` instead.

### No global registry

A long-running tool should create one `source.Registry` and pass it to the packages that share file IDs. Core does not provide a process-wide singleton.

### Validation before mutation

Workspace edits are checked as a complete set before content is returned. Writing files, handling rollback, and asking an editor to apply changes remain the host's responsibility.

## Ownership

Core owns source locations, edits, diagnostics, their wire format, and small hashing helpers. These concerns live elsewhere:

| Concern | Owner |
|---|---|
| Pawn parsing | `pawn-parser` |
| Preprocessing and semantics | `pawn-analysis` |
| Projects, manifests, and toolchains | `pawn-project` |
| SA-MP and open.mp API data | `pawn-api` |
| Shared schemas | `pawnkit-spec` |
| Lint policy | `pawnlint` |
| User workflows | `pawnkit-cli` |

Logging and tool-specific caches stay with each tool.

Source and diagnostic lookalikes are classified in
[type-boundaries.md](type-boundaries.md). Wire protocols keep their own types.

## Adding something to core

Core is intentionally small because every downstream repository pays for its API changes. A new package belongs here only when all of these are true:

- at least two independent PawnKit repositories need the same concept;
- the API is small enough to support conservatively;
- no repository already owns the concern;
- it does not pull a higher-level PawnKit module into core;
- untrusted input and performance-sensitive code have suitable tests.

CLI helpers, logging frameworks, project models, parsers, and generic utility packages do not belong here.

Open an issue before writing a new package. Name the downstream consumers and explain why a private implementation in either repository would create an incompatible contract.
