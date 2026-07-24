# Signaler

## Covers

- `internal/ohlcv/ohlcv.go`
- `internal/signaler/signaler.go`
- `internal/signaler/macross.go`
- `internal/signaler/rsi.go`
- `internal/runtime/runtime.go`

## Intent

Signaler MUST select one calculator, load complete closed OHLCV bars, calculate
indicators once, and release ordered Signals without lookahead.

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
Signaler.Init(config, source, start, end)
  select Macross or RSI calculator
  resolve exact timeframe and warmup requirements
  load required OHLCV
  calculate indicators wholesale
  generate all Backtest Signal candidates
  validate timestamp order
  retain OHLCV and Signals

Signaler.Start()
  require prepared state
  admit release requests

Signaler.Run(now)
  release next Signal only after now crosses the next-row start
  set availableMS to now

Signaler.Stop()
  close admission
  log calculated, released, and pending counts
```

## Data Contract

Every OHLCV value MUST contain aligned:

```text
startMS, open, high, low, close, volume
```

The loader MUST reject missing files, nulls, invalid OHLCV, gaps, duplicates,
wrong types, and incorrect range length.

`SignalMS` MUST identify the closed bar that produced the Signal.
The next row start MUST prove closure.

`AvailableMS` MUST record the release observation time, not a theoretical bar
end. Signals MUST release in strictly increasing `AvailableMS` order.

Live calculators MUST generate only the concrete Signaler's required frame
tail. Tail length is strategy-owned.

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

Signaler MUST report intervals, rows loaded, Signals calculated, Signals
released, and Signals pending. Runtime MUST separately report received and
active-cycle-skipped Signals.
