# Signal

Status: Implemented.
Covers: `internal/signaler`, `internal/controller`
Purpose: Carry immutable strategy lifecycle state into Controller.

## Shape

One Package contains:

- symbol;
- availability timestamp;
- typed Action;
- regime;
- risk score; and
- diagnostic fields.

Actions are `NoAction`, `StartCycle`, and `StopCycle`.

Direction, Account, symbol selection, Executor role, capital, and order sizing
never come from Signal.

## Traffic-Light Semantics

Macross reports `StartCycle` while its complete bullish condition remains
true.

It reports `StopCycle` when the condition becomes false after indicators are
ready.

Controller may reuse the current `StartCycle` action after a cycle completes.

It waits until the next control event.

There is no queue and no fresh-crossover requirement.

## Ownership

Signaler calculates.

Controller arbitrates.

BotCycle coordinates.

Executor executes its fixed Config.

Signaler calls none of those owners.

## Evidence

Controller records each newly observed Package timestamp and Action once.

ResultPublisher stores ordered Signal decisions.
