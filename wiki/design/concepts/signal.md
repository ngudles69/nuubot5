# Signal

Status: Implemented.
Covers: `internal/signaler`, `internal/controller`
Purpose: Carry one immutable strategy package through Controller, BotCycle, and Executors.

## Shape

One Package contains:

- symbol;
- availability timestamp;
- typed Action;
- regime;
- risk score; and
- arbitrary BotSpec-defined custom fields.

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

Controller reads the standard Action and arbitrates lifecycle.

Controller passes the complete unchanged package into active `BotCycle.Run`.

BotCycle passes the complete unchanged package to supported running Executors.

Each Executor reads only the standard or custom fields required by its fixed BotSpec.

Signaler calls none of those owners.

## Evidence

Controller records each newly observed Package timestamp and Action once.

An active BotCycle receives the complete current package on every Controller
pass. Custom Executor handling must preserve persistent traffic-light semantics.

ResultPublisher stores ordered Signal decisions.
