# Signaler

## Covers

- `internal/bars/bars.go`
- `internal/signaler/signaler.go`
- `internal/signaler/macross.go`
- `internal/signaler/rsi.go`
- `internal/runtime/runtime.go`

## Intent

Signaler MUST load complete closed OHLCV bars, calculate indicators once, and
release ordered Signals without lookahead.

## Ownership

Runtime MUST own one `Signaler`. `Signaler` MUST own:

- the selected calculator;
- required Bars;
- calculated Signals;
- the next unreleased Signal index;
- its lifecycle state.

The calculator interface is valid because Macross and RSI are both current
implementations of the same boundary.

## Program Flow

```text
signaler.New(config)
  select Macross or RSI calculator

Signaler.BarsNeeded()
  return exact timeframe and warmup requirements

Signaler.Prepare(bars)
  calculate all Signals once
  validate timestamp order
  retain Bars and Signals

Signaler.Start()
  require prepared state
  admit release requests

Signaler.Next(now)
  release next Signal only when availableMS < now

Signaler.Stop()
  close admission
  log calculated, released, and pending counts
```

## Data Contract

Every Bars value MUST contain aligned:

```text
startMS, endMS, open, high, low, close, volume
```

The loader MUST reject missing files, nulls, invalid OHLCV, gaps, duplicates,
wrong types, and incorrect range length.

`SignalMS` MUST identify the closed bar that produced the Signal.
`AvailableMS` MUST be strictly later than `SignalMS`. Signals MUST be released
in strictly increasing `AvailableMS` order.

## Implementations

### Macross

Macross MUST calculate:

- 1h EMA 9/21 crossover;
- backward-aligned closed 4h EMA 200 regime filter.

It MUST NOT use an unclosed 4h bar.

### RSI

RSI MUST calculate:

- 1h RSI;
- volume moving-average confirmation.

Indicator calculations MUST remain in the concrete calculator. Runtime MUST
receive only Signals.

## Evidence

Signaler MUST report timeframes, bars loaded, Signals calculated, Signals
released, and Signals pending. Runtime MUST separately report received and
active-cycle-skipped Signals.
