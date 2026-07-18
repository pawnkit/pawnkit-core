# Performance

Source indexing and edit application run in formatters, linters, and language servers. A small regression here is repeated across every file they touch.

## Run the benchmarks

```sh
go test ./... -run=^$ -bench=. -benchmem
```

`task bench` runs the same command. Keep `-benchmem`; allocations matter in long-running editor processes.

The current benchmarks cover a 50,000-line mixed-encoding file and 10,000 non-overlapping edits in a 200 KB source file. They are synthetic, so use them for before-and-after comparisons rather than claims about a particular project.

## Compare a change

Run the relevant package several times and compare the samples with [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```sh
go test ./source/... -run=^$ -bench=. -benchmem -count=10 > old.txt
# make the change
go test ./source/... -run=^$ -bench=. -benchmem -count=10 > new.txt
benchstat old.txt new.txt
```

## Current development baseline

Results vary by machine. These numbers are useful mainly as a rough check that a benchmark has not changed by an order of magnitude.

| Benchmark | Time/op | B/op | allocs/op |
|---|---:|---:|---:|
| `NewLineIndex` (50k lines) | ~0.9 ms | ~2.0 MB | 24 |
| `PositionUTF16` | ~34 ns | 0 | 0 |
| `PositionUTF8` | ~33 ns | 0 | 0 |
| `LineAt` | ~39 ns | 0 | 0 |
| `ApplyManyEdits` (10k edits) | ~0.29 ms | ~0.59 MB | 4 |

Do not fail CI on a single noisy benchmark result. Re-run it, compare distributions, and investigate a repeatable change.
