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
