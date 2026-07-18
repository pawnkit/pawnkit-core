# pawnkit-core

`pawnkit-core` contains the small Go types shared by PawnKit tools: source positions, text edits, diagnostics, and their JSON wire format.

This keeps a diagnostic or workspace edit consistent as it moves between the parser, linter, CLI, and editor. The module has no dependency on another PawnKit repository.

## Status

The module is pre-1.0. Breaking changes are recorded in
[CHANGELOG.md](CHANGELOG.md); the version policy is in
[docs/compatibility.md](docs/compatibility.md).

## Install

```sh
go get github.com/pawnkit/pawnkit-core
```

Requires Go 1.26 or later.

## Choose a package

| Package | Purpose |
|---|---|
| [`source`](source) | File IDs, URIs, byte spans, snapshots, and editor position conversion |
| [`textedit`](textedit) | Validated, version-aware file and workspace edits |
| [`diagnostic`](diagnostic) | Findings, related locations, fixes, and suppressions |
| [`protocol`](protocol) | Versioned JSON encoding for diagnostics |
| [`hash`](hash) | Stable hashes for content and cache keys |

Runnable examples for each package live under [`examples/`](examples).

## Quick start

```go
package main

import (
	"fmt"

	"github.com/pawnkit/pawnkit-core/diagnostic"
	"github.com/pawnkit/pawnkit-core/source"
)

func main() {
	reg := source.NewRegistry()
	file := reg.Intern(source.FileURI("/gamemodes/main.pwn"))

	span, _ := source.NewSpan(file, 10, 25)

	d := diagnostic.New(
		"pawnlint:deprecated-func",
		"pawnlint",
		diagnostic.SeverityWarning,
		"SetPlayerColor is deprecated",
		span,
	)

	if err := d.Validate(); err != nil {
		panic(err)
	}

	fmt.Println(d.Message)
}
```

## Contracts worth knowing

- Source spans use byte offsets. Convert them through `LineIndex` when an editor needs UTF-8, UTF-16, or UTF-32 positions.
- `source.FileID` is process-local. Serialized diagnostics use a URI instead.
- `textedit` computes new content but does not write files.
- Each type documents whether its zero value is usable.

## What does not belong here

Parsers, project manifests, API data, CLI frameworks, and tool-specific caches belong in their owning repositories. [docs/architecture.md](docs/architecture.md#scope-and-boundaries) explains the boundary.

## Links

- PawnKit organisation: <https://github.com/pawnkit>
- Architecture and scope: [docs/architecture.md](docs/architecture.md)
- Performance budgets and benchmarking: [docs/performance.md](docs/performance.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Changelog: [CHANGELOG.md](CHANGELOG.md)
