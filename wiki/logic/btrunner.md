# BtRunner

## Covers

- `cmd/nuubot-btrunner/main.go`
- `internal/btrunner/btrunner.go`
- `internal/setup/setup.go`
- `internal/toolkit/clock/clock.go`
- `internal/runtime/runtime.go`
- `internal/replay/parquet.go`

## Intent

BtRunner MUST run one configured Sweep Bot over one bounded historical range.

It MUST serve validated replay ticks and drive Runtime on the configured clock.

It MUST close an active BotCycle at the effective end date and prove exact
completion.

BtRunner MUST NOT own trading decisions or collect child statistics.

## Ownership

```text
main
`-- BtRunner
    |-- TickClock
    |-- replay.Reader
    `-- Runtime
        |-- Signaler
        |-- []Risk
        `-- BotCycle
            `-- []Executor
```

Each object MUST have one direct owner. Parents MUST stop only direct children.

## Program Flow

```text
main()
  run(args)
  log one terminal error
  exit non-zero on error

run(args)
  parse sweepID and botID
  load config
  configure one process logger
  runner = btrunner.New(...)
  runner.Start()
  runErr = runner.Run()
  stopErr = runner.Stop()
  return both errors without hiding runErr

btrunner.New(...)
  setup.Init(...)
  select effective end date
  create TickClock
  create replay.Reader
  create Runtime
  load required Bars
  prepare Signaler
  calculate expected replay proof

BtRunner.Start()
  start Runtime
  mark started

BtRunner.Run()
  for each validated BBO
    Runtime.Ingest(BBO)
    record served tick
    TickClock.Advance(timestamp)
    if due
      Runtime.MainLoop(timestamp)
      stop when Runtime requests stop

  Runtime.Stop("end_date")
  verify exact replay

BtRunner.Stop()
  stop Runtime
  stop Reader
  stop Clock
  log BtRunner-owned proof
```

## Replay Contract

- Canonical builds MUST use `-tags noasm`.
- Parquet input MUST be treated as untrusted boundary data.
- Every timestamp MUST pass `admitTick`.
- A one-second close MUST fall within microseconds `999000..999999`.
- Admitted ticks MUST normalize to an exact millisecond boundary.
- Consecutive ticks MUST be exactly 1,000 ms apart.
- Invalid OHLCV, nulls, gaps, duplicates, missing files, or wrong types MUST
  fail before Runtime accepts the affected value.

## Completion Contract

BtRunner success requires all of:

- process exit zero;
- `replay_completed=true`;
- served ticks equal expected ticks;
- triggered passes equal expected passes;
- first and last timestamps equal the expected range;
- Runtime terminal statistics are internally consistent;
- all direct children stop successfully.

For Sweep 6 Bot 9, the current accepted result is:

```text
ticks                 7,948,800
passes                  794,880
signals                      55
signals skipped              37
cycles started               18
cycles closed                18
stop-loss exits              17
end-date exits                1
```

## Exclusions

Account, Ledger, Trade, Order, Fill, Simulator, Venue, server, CLI, and web code
MUST NOT be added to BtRunner.
