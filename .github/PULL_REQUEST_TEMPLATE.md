## Summary

<!-- What does this change do, and why? -->

## Scope check

- [ ] The change belongs in core under [the ownership rules](../docs/architecture.md#ownership).
- [ ] This change does not introduce a dependency on another PawnKit
      repository.
- [ ] Any new external dependency is justified in this PR's description.

## Testing

- [ ] `task check` passes locally.
- [ ] Race tests pass if shared state changed.
- [ ] New/changed public behaviour has table-driven tests.
- [ ] Any fixed bug has a regression test.
- [ ] If `source` or `textedit` changed: the relevant fuzz target was run
      locally for at least a few seconds with no failures.
- [ ] If the `protocol` wire format changed: `DiagnosticSchemaVersion` was
      bumped and a new fixture was added (existing fixtures were not
      edited), per `docs/compatibility.md`.

## Breaking changes

<!-- None, or describe the break and the migration path. -->
