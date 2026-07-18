# Contributing to pawnkit-core

PawnKit is maintained by volunteers, so reviews may take a little time.

Small, well-tested changes are welcome. You do not need to work across the
whole PawnKit toolchain to improve one package or fix one edge case.

Core APIs are used across PawnKit, so changes here need a clear owner and a stable reason to exist. Read [docs/architecture.md](docs/architecture.md) before proposing a package or a large public API.

## Local checks

PawnKit Core requires Go 1.26 or later. The dev container includes Go, Task, and golangci-lint.

```sh
go build ./...
task check
```

`task check` runs formatting, vet, lint, and tests. The individual commands are available as `task fmt`, `task vet`, `task lint`, and `task test`.

## Tests

Add a regression test with every bug fix. Public behavior changes should cover both the expected case and invalid input.

Run the race detector when changing shared state:

```sh
CGO_ENABLED=1 go test -race -shuffle=on ./...
```

Run the relevant fuzz target when changing source positions, file URIs, or edits:

```sh
go test ./source -run=^$ -fuzz=FuzzLineIndexPosition -fuzztime=30s
go test ./source -run=^$ -fuzz=FuzzFileURIFilename -fuzztime=30s
go test ./textedit -run=^$ -fuzz=FuzzApply -fuzztime=30s
```

Protocol changes must keep the frozen fixtures under `protocol/testdata`. Add a fixture for a new schema version instead of editing an existing version.

## Pull requests

Keep package comments short and explain behavior that callers cannot infer from the API. If a change adds generated files, document how to reproduce them and add a CI check for stale output.

Run `task check` before opening a pull request.

## Releases

Releases use semantic version tags. Before 1.0, breaking changes still need a migration note in [CHANGELOG.md](CHANGELOG.md). The full policy is in [docs/compatibility.md](docs/compatibility.md).

Report vulnerabilities through the private route in [SECURITY.md](SECURITY.md), not a public issue.
