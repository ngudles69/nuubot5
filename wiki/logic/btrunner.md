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

It MUST close an active BotCycle during shutdown after Reader exhaustion and
prove exact completion.

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
  loopErr = runner.Loop()
  stopErr = runner.Stop()
  return both errors without hiding loopErr

btrunner.Init(...)
  setup.Setup(...)
  select effective end date
  create TickClock
  initialize TickClock at replay start
  register Runtime timer callback
  initialize replay.Reader
  initialize Runtime
  calculate expected replay proof

BtRunner.Start()
  start TickClock
  start Runtime
  mark started

BtRunner.Loop()
  for each validated BBO
    Runtime.IngestBBO(BBO)
    record served tick
    TickClock.Advance(timestamp)
      registered timer callback runs Runtime
    stop when Runtime requests stop

  verify exact replay

BtRunner.Stop()
  stop Clock
  stop Reader
  stop Runtime
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
```

## Exclusions

Account, Ledger, Trade, Order, Fill, Simulator, Venue, server, CLI, and web code
MUST NOT be added to BtRunner.
