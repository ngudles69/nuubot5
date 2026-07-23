# Signaler Package

Status: Implemented.
Covers: `internal/signaler/*.go`
Purpose: Create and run one configured Signal strategy behind a stable factory.

## Canonical Source

- `D:/rust/nuubot4/src/signaler.rs`

## Scope & Responsibilities

SignalerFactory selects one concrete Signaler. Runtime knows only the common
Signaler contract.

- Concrete Signalers own configuration, Bars requirements, and Signal calculation.
- New strategies add one concrete file and one factory case.

## Program Flow

```text
SignalerFactory(kind, log, ctx) -> Signaler

init
  concrete = select kind
  concrete.init(log, ctx)

start
  concrete.start()

run(now)
  return every Signal available before now

stop
  concrete.stop()

---

domain
  macross = 1h EMA 9/21 cross filtered by closed 4h EMA 200 regime
```

## Notes

- Current configured Signaler is Macross.
- Signals release only after their source Bar closes.
