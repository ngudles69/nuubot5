# Signaler Package

Status: Implemented for replay.
Covers: `internal/signaler/*.go`
Purpose: Calculate and serve immutable strategy lifecycle state.

## Contract

```go
type Signaler interface {
    Signals(string, uint64, int) []Package
    Stop()
}
```

One Package contains symbol, timestamp, typed Action, regime, risk score, and
diagnostic fields.

Actions are:

- `NoAction`;
- `StartCycle`; and
- `StopCycle`.

There are no entry-direction or close-direction fields.

Executor Config owns side, symbol, Account, role, capital, and order sizing.

## Implementations

Macross emits persistent traffic-light state from closed signal and regime
bars.

RSI emits persistent state from confirmed RSI and volume conditions.

Signaler loads no BotConfig after initialization.

## Invariants

- Packages are ordered by unique timestamp.
- Closed bars preserve no-lookahead behavior.
- Package availability equals bar start plus its configured interval.
- The final fully closed loaded bar produces a Package.
- Signaler performs no lifecycle or trading mutation.
- Controller may reuse the current action after a BotCycle completes.
