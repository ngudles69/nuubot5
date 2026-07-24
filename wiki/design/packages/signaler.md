# Signaler Package

Status: Implemented for replay.
Covers: `internal/signaler/*.go`
Purpose: Create one initialized, passive Signaler.

## Ownership

`signaler.Create` owns concrete selection and initialization.

Runtime owns the returned Signaler.

Macross and RSI are complete Signaler implementations.

They are not indicator calculators.

TA remains the indicator abstraction.

## Contract

```go
type Signaler interface {
    Signals(symbol string, atMS uint64, count int) []Package
    Stop()
}
```

Signaler has no `Start`, `Run`, or `Release`.

It owns no timer, execution, or Runtime stop policy.

## Program Flow

```text
Create
  select signaler
  initialize signaler

Macross.Init
  configure macross
  load macross data
  calculate macross signals
  validate packages
  initialize signaler

macross.configure
  parse intervals
  validate config

RSI.Init
  configure rsi
  load rsi data
  calculate rsi signals
  validate packages
  initialize signaler

rsi.configure
  parse interval
  validate config

CreatePackage
  validate standard fields
  validate custom fields
  create package

loadSeries
  load ohlcv

Stop
  stop signaler
```

## Concrete Signalers

Macross owns its timeframes, EMAs, regime alignment, and package fields.

RSI owns its timeframe, RSI, volume confirmation, and package fields.

Each calculates one package for every admitted signal bar.

Each package contains standard triggers plus all useful calculated fields.

## History

`Signals` returns packages at or before `atMS`.

Results are chronological and contain at most `count` entries.

Wrong symbols, stopped Signalers, and nonpositive counts return no packages.

The latest package is the final slice element.

See [Signal](../concepts/signal.md).

## Current Boundary

Replay initialization loads and calculates the complete range.

Live bar ingestion has no current caller or implemented contract.

Add it when Runner owns a real closed-bar feed.

## Approved Target Contract

Status: Approved target design. Not implemented.

The exact BotSpec selects and configures its Signaler.

Config does not select an arbitrary Signaler kind.

Controller owns Signaler for the complete BotGeneration.

Signaler starts once with Controller and stops only when Controller stops.

It remains active:

- Before any BotCycle.
- While BotCycles run.
- While BotCycles stop.
- Between completed BotCycles.

It preserves EMA, KAMA, HMA, regime, and other strategy history across cycles.

## Traffic-Light Meaning

Signaler is a passive strategy signal source.

It may report:

```text
NoAction
StartCycle
StopCycle
```

Signaler never:

- Creates a BotCycle.
- Stops a BotCycle.
- Checks whether one BotCycle is active.
- Checks Account availability.
- Checks capital.
- Calls Controller or any trading child.

Controller interprets every strategy signal.

`NoAction` starts nothing and does not stop an active BotCycle.

`StartCycle` may be ignored when Risk, Account, Meta, or capital
admission blocks it.

`StopCycle` requests exit of the active BotCycle.

Signaler cannot stop Controller.

Symbol, side, Account, capital, and Order sizing are fixed Executor Config.

Signaler continues unchanged after a blocked start.

## Executor Unit

One accepted entry signal starts one complete BotCycle Executor unit.

Every Executor receives the same immutable strategy Signal.

No Executor owns an independent strategy signal subscription.

No signal selects only part of the BotSpec's Executor unit.

Executors may monitor or act differently according to their fixed BotSpec roles.

While a BotCycle runs, another `StartCycle` does nothing.

After BotCycle completion, Controller rechecks the current action on the next
control event.

No fresh crossover or queued Signal is required.

See [Signal](../concepts/signal.md) and [BotSpec](../concepts/bot-spec.md).
