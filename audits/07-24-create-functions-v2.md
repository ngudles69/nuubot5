# Create Function Audit

FAIL

Lifecycle ownership is corrected. Two unrelated cleanup findings remain.

## Findings

### Medium: replay suite counter collision

Location: `rtest.sh:5`, `rtest.sh:47`, `rtest.sh:88`

Runtime proof parsing overwrites requested suite `runs` with 794,880.

Why it matters: `./rtest.sh 1 6 9` continues running instead of stopping after
one process.

Required fix: store parsed Runtime runs in a distinct `runtime_runs` variable.

### Low: unused Replay failure state

Location: `internal/replay/replay.go:19`, `internal/replay/replay.go:49`,
`internal/replay/replay.go:59`

`Reader.failed` is written but never read.

Why it matters: it implies failure behavior that does not exist.

Required fix: remove the field and both assignments.

## Lifecycle assessment

| Component | Decision | Reason |
|---|---|---|
| Runtime | `Init` | One concrete Runtime owned directly by BtRunner. |
| Replay | `Init` | One concrete Reader owned directly by BtRunner. |
| BotCycle | `Init` | One concrete Control initialized for each accepted Signal. |
| Signaler | `Init` | One concrete Signaler owns selected calculator implementations. |
| Clock | Keep `Create` | Approved construction boundary for selectable Clock versions. |
| Executor, Risk | Keep `Create` | Configuration factories select concrete implementations. |
| Logging, Market | Keep constructors | They are ordinary constructors, not lifecycle factories. |

## Proof checked

- Focused lifecycle tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed.
- `gofmt -d` returned no output.
- `git diff --check` passed.
- Fresh direct BtRunner execution exited zero.
- Fresh replay served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Fresh Runtime produced 55 Signals and 18 closed cycles.
- Replay completed in 1,088 ms.
- No old Runtime, Replay, BotCycle, or Signaler constructor remains in source.

## Proof missing

- Clean `./rtest.sh 1 6 9` completion awaits the counter fix.

## Assumptions

- Clock remains an approved future selectable construction boundary.
- Dead Replay failure state remains outside the approved lifecycle cleanup.

## Open questions

- None.

Bloat check: unused Replay failure state and one harness variable collision
remain. No fake code, half-wired dependency, fallback, race risk, logic drift,
overcomplicated lifecycle, or low-value optimization remains in the reviewed
constructor paths.
