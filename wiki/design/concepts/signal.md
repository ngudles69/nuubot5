# Signal

Status: Implemented.
Covers: `internal/signaler/*.go`
Purpose: Carry immutable, timestamped calculated facts from Signaler to Runtime and Executors.

## Meaning

A Signal package contains facts, not commands.

Signaler calculates and stores it.

Runtime uses only standard entry triggers.

Executor decides whether and how to use custom fields.

## Standard Fields

Every package contains:

```json
{
  "symbol": "BTC",
  "timestamp_ms": 1234567890,
  "enter_long": true,
  "enter_short": false,
  "close_long": false,
  "close_short": false,
  "regime": "bull",
  "risk_score": 24
}
```

Standard field names and types are reserved.

`risk_score` uses zero through 100.

Zero represents the lowest risk.

Both entry triggers cannot be true.

## Custom Fields

Concrete Signalers append calculated fields at the same top level.

```json
{
  "symbol": "BTC",
  "timestamp_ms": 1234567890,
  "enter_long": true,
  "enter_short": false,
  "close_long": false,
  "close_short": false,
  "regime": "bull",
  "risk_score": 24,
  "vol_spike": 1.3,
  "vp_lvz": true,
  "vp_hvz": false,
  "vp_poc": 343.2
}
```

There is no nested `extra_signals` object.

Current Macross packages include bar time, price, and three EMA values.

Current RSI packages include bar time, price, RSI, volume ratio, and threshold facts.

Executor reads only the fields it needs.

## Timestamp

`timestamp_ms` is the earliest safe availability time.

It uses the next admitted bar start.

Signal queries never return a package after the requested time.

## History

Macross and RSI produce one package for every admitted signal bar.

Executor may request the last N packages.

Returned packages are ordered oldest to newest.

The latest package is the final element.

The Signaler may return fewer packages than requested.

Executor owns insufficient-history policy and timestamp guards.

## Runtime Shape

`Package` stores decoded fields behind typed read methods.

Its field map is not exposed.

JSON marshaling produces the same flat object.

Replay performs no JSON decoding in its hot loop.

## Invariants

- Standard fields always exist.
- Custom fields cannot replace standard fields.
- Packages are ordered by increasing availability time.
- Package time never exceeds query time.
- Signaler never tracks Executor consumption.
