# Create Function Audit

PASS

No remaining `Create` function is empty or better moved into `Init`.

## Assessment

| Function | Decision | Reason |
|---|---|---|
| `market.CreateBBO` | Keep | Validates untrusted values and returns one trusted immutable value. |
| `logging.Create` | Keep | Constructs the project logger around an injected standard-library writer. |
| `clock.Create` | Keep | Binds the required logger; `RegisterTimer` owns separately validated timer policy. |
| `replay.Create` | Keep | Opens its owned reader and returns one ready resource in one step. |
| `signaler.Create` | Keep | Selects and validates the calculator before Runtime can request its OHLCV requirements. |
| `createMacross`, `createRSI` | Keep | Construct validated concrete calculators behind the Signaler factory. |
| `risk.Create`, `createBalanced` | Keep | Select and construct the configured Risk without an empty lifecycle phase. |
| `executor.Create`, `createObserver` | Keep | Select and validate the configured Executor before BotCycle starts it. |
| `botcycle.Create` | Keep | Constructs a dynamic BotCycle and its configured Executors in one operation. |

Adding `Init` to these paths would create two-step construction, hide configured
implementation selection, or split one complete resource-opening operation.

## Proof checked

- Inspected every exported and private `Create` function and every caller.
- Checked owning package design pages.
- `go test -tags noasm ./...` passed.
- Focused Runtime and BtRunner tests passed.
- `gofmt -d` returned no output.
- `git diff --check` passed.
- No `runtime.Create`, `create runtime`, or `runtime create` reference remains.

## Proof missing

- Real BtRunner replay was not run because construction behavior did not change.

## Assumptions

- Scope covers constructor indirection, not unrelated implementation quality.
- Explicit factory boundaries for Signaler, Executor, and Risk remain approved.

## Open questions

- None.

Bloat check: no fake code, unused constructor, dead stub, half-wired dependency,
fallback, race risk, logic error, remaining inefficient indirection,
overcomplicated constructor, or low-value optimization was found.
