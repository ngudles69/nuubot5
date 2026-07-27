# Nuubot Runner Implementation Plan

Status: PENDING PLAN AUDIT

Purpose: Add live Runner execution without changing BtBot behavior or duplicating trading policy.

## Planning and Audit Write Boundary

Every planner or auditor task must name its exact approved output file or files before work starts.

The role may write only those approved outputs.

If no output file is approved, the role is read-only.

For this task, the planner may write only `NUUBOT_RUNNER_PLAN.md`.

The plan auditor is read-only.

The plan auditor writes no audit file.

Audit findings return to the planner.

Only the planner may revise the approved plan file.

Unless the user later authorizes a named file, this scope is prohibited:

```text
Source edits; Test edits; Script edits; Config edits; Wiki edits; Handoff edits; Implementation commands.
```

Read-only code assessment roles are code inspectors.

Coder means implementation authority.

Auditor means read-only review against objectives and proof.

## Mandatory Coder Execution Gate

This gate applies before Target Change #1 and before every later Target Change.

The coder must re-read every current file named by that target.

The coder must compare current code with the target's Before section.

If they match, the coder may implement only that surgical target.

If they differ, the coder makes no edit.

The coder reports the exact conflict, affected symbols, upstream impact, downstream impact, and proof impact.

The planner revises this file.

The plan auditor re-audits the revision.

Coding resumes only after the user accepts the audited plan.

Concurrent work is never overwritten, reverted, reformatted, or treated as precedent without review.

Before every shared-core target, the coder captures current BtBot proof.

After every shared-core target, the coder repeats the same proof and compares semantic results.

Any unexplained difference stops the target.

## Assessment Basis

### Nuubot5 Canonical Sources

- `HANDOFF.md`
- `wiki/PROJECT.md`
- `wiki/USER.md`
- `wiki/SOUL.md`
- `wiki/ARCHITECTURE.md`
- `wiki/DESIGN.md`
- Current Nuubot5 code at inspection time

### Nuubot3 Reference Boundary

Use Git objects, not its inconsistent worktree.

- Commit `2f6906fd33493f8587ff8aaf39d8ebef8ebab4c6`
- `nuubot/runner/runner.py`
- `nuubot/runner/btrunner.py`
- `nuubot/runtime/runtime.py`
- `nuubot/runtime/botcycle.py`
- `nuubot/account/account.py`
- `nuubot/exchange/bars.py`
- `nuubot/exchange/ws.py`
- `nuubot/core/clock.py`
- `wiki/runner-lifecycle.md`

Committed snapshots from `76ee05a` may clarify ownership and cleanup:

- `nuubot/runner/runner.py.backup`
- `nuubot/runtime/runtime.py.backup`
- `nuubot/runner/btrunner0.py`
- `nuubot/runner/btrunner1.py`
- `nuubot/runner/service.py`

Do not copy recovery, pause, resume, control-plane, alternate replay-loader, or current-worktree behavior.

### Nuutrader6 Reference Boundary

Use it only for these concerns:

- Live data and execution network separation.
- Account Venue selection.
- Simulator mechanics.
- Reconciliation windows.
- Lifecycle and STOP sequencing.
- Public stream supervision.

Useful references:

- `src/nuubot/services/data_engine.py`
- `src/nuubot/hcbots/account.py`
- `src/nuubot/hcbots/exchange.py`
- `src/nuubot/hcbots/simulator.py`
- `src/nuubot/hcbots/recon.py`
- `src/nuubot/hcbots/runtime.py`
- `src/nuubot/hcbots/store.py`
- `src/nuubot/hcbots/ghbot/executor.py`

Do not copy its missing Controller, Signaler, or Risk architecture.

Do not copy `Any` Venue typing, direct feed-policy calls, aggressive cancellation inference, incomplete fee handling, or automatic recovery.

## Confirmed Facts

- `cmd/nuubot-runner/main.go` is only an `Under Construction.` reservation.
- `internal/runner` does not exist.
- Both Runner locations require genuinely new implementation.
- `cmd/nuubot-bt-bot/main.go` is the current command clarity and lifecycle example.
- BtBot owns current Parquet replay, TickClock, Controller, proof, reporting, and publication.
- `setup.Setup` is replay-specific.
- `datastore.Bot` always carries `ReplayInput`.
- `botspec.Build` reads replay fields directly.
- `signaler.Create` loads complete Parquet OHLCV ranges.
- Signalers expose no live Bar bootstrap or completed-Bar ingestion.
- `market.BBO` contains one price.
- Controller has no Bar ingestion, user-event dirty path, or freshness gate.
- Controller uses dynamic maps.
- Account directly owns `simulator.Simulator`.
- Account request, response, reconciliation, and result types leak Simulator types.
- BotSpec, Executor, and Account repeat Simulator and simnet gates.
- Hyperliquid implements Meta and clearinghouse-state REST only.
- Signing, exchange mutations, Order queries, Fill queries, and WebSocket are missing.
- WallClock owns one goroutine and runs callbacks synchronously.
- No Runner event lane exists.
- Ledger and Simulator `max` reload child state only.
- Full Runner recovery does not exist.
- BalancedRisk always returns `Allow`.
- BalancedRisk is not protection.
- Nuubot5 has active uncommitted work across shared packages.

## Inferences

- One source-neutral Bar contract is required for shared Signaler calculations.
- A two-price BBO is required for correct live and Simulator behavior.
- Runner needs process-local public feeds for Server-independent operation.
- Feed and Clock callbacks need one serialized Runner lane.
- Simnet can ship before signing exists.
- A crash-retained filesystem claim can fail closed without Server or ProcessStore.
- Live run evidence needs a Runner-owned database distinct from BtBot result semantics.
- Testnet needs current official signing vectors before mutation authorization.

## Current Conflicts

- Runner docs require fixed Trade, Order, and Fill capacity.
- Account and Ledger use dynamic maps and reject fixed reserves.
- Runner docs conflate live mode with testnet, paper, and simulator networks.
- The user requires only `bt` and `live` process modes.
- Setup always loads current mainnet Meta.
- BotSpec and Signaler require ReplayInput and Parquet paths.
- Venue design requires one Account-owned interface.
- Account still calls concrete Simulator methods.
- Simulator docs deny a Venue interface.
- Executor and BotSpec reject every non-Simulator resource.
- Hyperliquid exchange, signing, and WebSocket designs are unimplemented.
- Runner telemetry persistence has no implemented owner or schema.
- Standalone status, claims, and recovery remain unresolved.
- `market.BBO` lacks bid and ask.
- Runner startup docs subscribe after bootstrap, leaving a possible market-data gap.
- Telemetry docs claim Recon telemetry composition not present in current code.
- Trading-schema prose describes stale complete-tree persistence behavior.
- Mainnet Meta use conflicts with testnet execution safety until identity parity is proven.

## User Decisions

- Process execution mode is explicit `bt` or `live`.
- BtBot remains `bt`.
- Runner remains `live`.
- Simnet and testnet are live execution networks.
- Mainnet execution is excluded.
- Simnet uses mainnet public market data.
- Simnet uses Simulator execution.
- Simnet loads no credentials.
- Testnet uses testnet public market data.
- Testnet uses Hyperliquid testnet execution.
- Testnet resolves only referenced credentials.
- Both modes share Controller, Signaler calculations, Risk, BotCycle, Executor, Account, Ledger, lifecycle, and domain evidence.
- Bt initial data remains current Parquet and OHLCV.
- Live initial data comes from the live market-data boundary.
- Live admits only validated closed Bars.
- Runner owns WallClock, feeds, event lane, supervision, heartbeat, telemetry, and shutdown.
- Feed callbacks never call trading policy.
- Account owns one minimal Venue boundary.
- Simulator and Hyperliquid testnet implement that boundary.
- Direct Simulator calls and repeated gates receive one hardcut.
- Simnet ships and proves first.
- First release fails closed after crash.
- Recovery is a separate target and approval gate.

## Assumptions

- Live Runner initially loads one stored Bot by globally unique `bot_id`.
- BtBot keeps `(sweep_id, bot_id)` identity.
- `physical_account_id` remains the credential reference.
- Simnet may use its configured account name without a credential.
- Live execution requires `persist_mode = max`.
- One Runner owns one local run database.
- One Runner owns process-local sockets.
- Public market subscriptions require no credential.
- Testnet mutation needs an explicit command opt-in.
- No mainnet endpoint may receive private requests.
- Existing pure-Go restrictions remain mandatory.

Stop and reassess if any assumption conflicts with current code or accepted design.

## Non-Goals

- Server implementation.
- RunnerControl.
- Shared WebSockets.
- DataEngine.
- Pause or resume.
- Automatic restart.
- Mainnet execution.
- New strategy behavior.
- New Risk behavior.
- New Executor behavior.
- Physical-account portfolio Risk.
- Shared capital reservation.
- Feed history retention.
- Generic event bus.
- Generic Venue package.
- Compatibility adapters.
- Dual old and new paths.
- Alternate BtBot replay loaders.

## System Map

```text
bt process
  cmd/nuubot-bt-bot
    internal/btbot
      setup bt admission
      Parquet OHLCV and ReplayReader
      TickClock
      shared Controller

live process
  cmd/nuubot-runner
    internal/runner
      live admission and claims
      mainnet or testnet public market boundary
      optional testnet user stream
      serialized event lane
      WallClock and heartbeat
      shared Controller
      live status, telemetry, and shutdown evidence

shared Controller
  Signaler calculation core
  BalancedRisk
  zero or one BotCycle
    Executor
      Account
        Account-owned Venue
          Simulator simnet
          Hyperliquid testnet
        Ledger
```

## Shared-Core Invariant

Bt and live use the same source-neutral calculation and trading objects.

Outer owners differ only in data acquisition, Clock, supervision, and terminal publication.

BtBot keeps its current Parquet/OHLCV bootstrap.

Runner never imports replay behavior.

Shared packages never detect BtBot or Runner.

Ledger and Venue never detect process mode.

No target may retain parallel replay-only and live-only trading policy.

## Mode and Network Matrix

```text
Process  Execution  Market data  Venue         Credentials  Allowed
bt       simnet     Parquet      Simulator     none         yes
bt       testnet    none         none          none         no
bt       mainnet    none         none          none         no
live     simnet     mainnet      Simulator     none         yes
live     testnet    testnet      Hyperliquid   referenced   yes, opt-in
live     mainnet    none         none          none         no
```

`bt` and `live` are process modes.

`simnet`, `testnet`, and excluded `mainnet` are execution networks.

Data network is derived once:

```text
simnet  -> mainnet public data
testnet -> testnet public data
```

No fallback may change either network.

## Dependency Graph

```text
TC01 baseline proof
  -> TC02 shared Bar and Signaler core
  -> TC03 source-neutral admission
  -> TC04 bid/ask BBO hardcut
  -> TC05 Controller ingress and runtime inputs
  -> TC06 Account Venue hardcut
  -> TC07 simnet gate cleanup
  -> TC08 claims and live persistence
  -> TC09 public live market boundary
  -> TC10 internal Runner simnet
  -> TC11 Runner command
  -> TC12 simnet release proof
  -> TC13 Hyperliquid testnet Venue
  -> TC14 testnet reconciliation and safety
  -> TC15 testnet release proof
  -> TC16 recovery design gate
```

No target may start before every dependency passes.

## Target-Change Index

- Target Change #1: Freeze BtBot and recorded-event proof.
- Target Change #2: Create one shared closed-Bar Signaler core.
- Target Change #3: Separate Bt and live admission.
- Target Change #4: Hardcut BBO to bid and ask.
- Target Change #5: Add source-neutral Controller ingress and runtime inputs.
- Target Change #6: Hardcut Account to one minimal Venue.
- Target Change #7: Remove repeated Simulator gates and finish simnet admission.
- Target Change #8: Add standalone claims, crash fence, and live persistence.
- Target Change #9: Add process-local public live market transport.
- Target Change #10: Implement simnet `internal/runner`.
- Target Change #11: Implement `cmd/nuubot-runner`.
- Target Change #12: Prove the simnet release.
- Target Change #13: Implement the Hyperliquid testnet Venue.
- Target Change #14: Complete testnet reconciliation and safety.
- Target Change #15: Prove the opt-in testnet release.
- Target Change #16: Gate separately approved recovery.

## Data Flows

### Bt

```text
stored Sweep Bot
  -> setup Bt admission
  -> current Parquet OHLCV bootstrap
  -> shared Signaler Bootstrap
  -> current ReplayReader close prices
  -> BBO with bid=ask=close
  -> Controller.IngestBBO
  -> TickClock.Advance
  -> Controller.Run
  -> existing STOP, Result, RunReport, ResultPublisher
```

Bt keeps BBO-before-timer ordering.

Bt keeps current tick counts, timestamps, Signal packages, cycles, domain evidence, finance, and publication behavior.

### Live Simnet

```text
stored live Bot
  -> acquire Account-symbol claims
  -> open Runner run database
  -> start mainnet public subscription into closed admission buffer
  -> fetch exact validated closed-Bar bootstrap
  -> merge buffered completed Bars without gaps
  -> obtain fresh mainnet BBO
  -> shared Signaler Bootstrap
  -> Controller Start
  -> open serialized event admission
  -> WallClock Start
  -> mainnet BBO and completed Bars
  -> Simulator Venue
  -> shared Account reconciliation and Ledger evidence
```

No credentials file is opened.

### Live Testnet

```text
stored live Bot
  -> require explicit testnet mutation opt-in
  -> resolve exact referenced testnet credential
  -> acquire Account-symbol claims
  -> open Runner run database
  -> start testnet public and user subscriptions into closed buffer
  -> fetch exact validated closed-Bar bootstrap
  -> prove fresh BBO, zero owned Orders, zero position, Meta, funds, and symbol safety
  -> shared Signaler Bootstrap
  -> Controller Start
  -> WallClock Start
  -> signed testnet Venue mutations
  -> REST reconciliation remains authoritative
```

No mainnet private endpoint is constructed.

## Live Startup Order

1. Parse one `bot_id` and optional testnet opt-in.
2. Open operational logging.
3. Load strict AppConfig.
4. Load exact stored live BotConfig.
5. Reject mainnet.
6. Derive public data network.
7. Resolve credentials only for testnet.
8. Build source-neutral BotDefinition.
9. Acquire every Account-symbol claim.
10. Create one live generation and run database.
11. Create process-local transports.
12. Start subscriptions with Controller admission closed.
13. Capture one fixed bootstrap boundary.
14. Fetch complete required closed Bars.
15. Merge buffered completed Bars.
16. Reject gaps, duplicates, stale Bars, or identity conflicts.
17. Obtain one fresh BBO per required symbol.
18. Initialize shared Signaler, Risk, Controller, and resources.
19. Run clean-slate and capital gates.
20. Start Controller.
21. Open Runner event admission.
22. Register one WallClock heartbeat.
23. Start WallClock last.
24. Persist `running` only after every dependency succeeds.

Any failure unwinds initialized children in reverse.

Claims remain after any ambiguous or mutation-capable failure.

## Steady-State Event Order

Transport callbacks only validate, timestamp, and enqueue typed events.

They never call Controller, Account, Executor, Risk, or Signaler.

One Runner loop consumes:

```text
completed Bar
  -> validate identity, closure, continuity, and freshness
  -> Controller.IngestBar

BBO
  -> validate identity, bid, ask, ordering, and freshness
  -> Controller.IngestBBO

user event
  -> validate account and network
  -> Controller.MarkAccountDirty

heartbeat
  -> read WallClock once
  -> Controller.Run
  -> collect Controller and Recon telemetry
  -> persist one Runner heartbeat

stop request
  -> close strategy admission
  -> run ordered STOP
```

The lane is FIFO by accepted sequence.

The first implementation uses one fixed bounded channel.

Overflow is fatal.

No event is silently dropped or coalesced.

Stop and reassess if observed traffic cannot fit the proven bound.

## Failure Paths

- Invalid external data fails before local mutation.
- Bar gap blocks policy admission.
- Stale BBO blocks new mutation decisions.
- Subscription failure reaches Runner.
- WallClock callback failure reaches Runner through `Err`.
- Event-lane overflow stops Runner.
- Controller failure stops Runner.
- First two Recon failures skip policy and retain last trusted generation.
- Third consecutive Recon failure starts STOP.
- Unknown mutation outcome creates no retry.
- Unknown mutation outcome remains dirty and reconciliation-pending.
- Persistence failure publishes no successful state.
- Claim conflict prevents Venue initialization.
- Crash leaves claims and nonterminal run evidence.
- Fresh start refuses crash-retained claims.
- STOP failure keeps claims.
- Testnet cleanup failure returns nonzero and keeps evidence.

## STOP Order

```text
Runner
  close new Bar, cycle, and mutation admission
  stop WallClock heartbeat admission
  cancel transport producers
  wait for owned transport goroutines
  close event-lane admission
  run Controller.Stop on the serialized owner

Controller
  stop active BotCycle
  stop Risks
  stop Signaler

BotCycle
  stop Executors in reverse order

Executor
  force Account reconciliation
  cancel owned active Orders
  reconcile
  close owned exposure
  reconcile
  wait for complete closure Fills and fees
  prove zero active Orders and zero position
  capture Account Result
  stop Account

Account
  stop Venue
  stop Ledger

Runner
  persist terminal status and telemetry
  close run database
  release exact claims only after successful flat proof
  return Result
```

Every Stop remains idempotent.

The first Controller stop reason wins.

Cleanup errors join the primary error.

Linux SIGINT and SIGTERM use the same path.

Windows interrupt uses the same cancellation path.

No forced process exit bypasses cleanup after Runner ownership starts.

## Persistence and Recovery Boundary

BtBot publication stays unchanged.

Live Runner uses one generation database:

```text
workspace/db/runners/bot_<bot_id>/generation_<generation_id>.db
```

Live Account uses `persist_mode = max`.

Runner owns lifecycle and heartbeat rows.

Account and Ledger own domain rows.

Simulator owns Simulator state.

Credentials enter no row.

External calls never occur inside SQLite transactions.

First release does not restore Controller, Signaler, BotCycle, or Executor policy.

Any nonterminal generation or retained claim blocks fresh start.

Operators must not delete claims until external truth is verified flat and order-free.

Recovery remains unavailable until Target Change #16 receives separate approval.

## Testnet Safety Gates

- Execution network must equal `testnet`.
- Command must include the explicit testnet mutation opt-in.
- Every resource must name `hyperliquid`, `testnet`, one credential reference, and one symbol.
- Mainnet is rejected before credential loading.
- The selected credential network must equal `testnet`.
- Duplicate credential matches fail.
- Empty address or key fails.
- Placeholder credentials fail.
- Current testnet Meta must match the traded symbol and asset identity.
- Missing Meta fails.
- Inactive, delisted, close-only, or unsupported symbols fail.
- Missing current BBO fails.
- Stale current BBO fails.
- Existing open Orders fail fresh start.
- Existing position fails fresh start.
- Insufficient funds or margin fails.
- Duplicate Account-symbol claim fails.
- BalancedRisk limitation is printed and persisted.
- Unknown mutation outcomes stop new mutations.
- Cleanup must prove flat and order-free.

## Credential Secrecy

Simnet never opens `credentials.toml`.

Testnet loads the catalog only after network and opt-in validation.

Account selects the exact referenced credential.

Only the selected credential reaches the Hyperliquid Venue.

Controller definition, Results, telemetry, logs, errors, tests, and artifacts contain no secret.

No full signed action, signature, private request body, address-key pair, or credential structure is logged.

Tests use synthetic unmistakably fake secrets.

Proof scans tracked files, logs, databases, JSON, and failure output.

Any secret-like value stops release.

## Unknown Mutation Outcomes

Account records created Order intent before external mutation.

Explicit item success or rejection updates ordered acknowledgement evidence.

Transport timeout, cancellation, malformed response, or lost response is unknown.

Unknown is neither success nor rejection.

Account marks the resource dirty.

Account performs no automatic retry.

Reconciliation searches exact CLOID and Venue identity.

New mutation remains blocked while owned unknown evidence is unresolved.

STOP may issue only evidence-safe cleanup actions.

If cleanup outcome is unknown, claims remain.

## Delayed Fee Reconciliation

Fill discovery and pending-fee repair remain independent.

```text
discovery  committed Fill cursor -> observed time
repair     bounded windows around unresolved timestamps
```

Both tracks use inclusive boundaries.

Both split capped responses.

Results merge by Venue TID.

The discovery cursor never clears a pending-fee anchor.

A filled acknowledgement without TID uses its acknowledgement time as repair anchor.

An observed TID replaces that anchor with the Fill timestamp.

Fee presence uses `HasFee`.

Zero fee and negative rebate remain valid.

Existing execution identity never changes during fee enrichment.

STOP waits for closure quantity and fee completion.

Every physical Fill query appears in Recon telemetry.

## Feed Freshness

Bars must be closed by the public source boundary.

Bootstrap must contain every required interval and warmup row.

Completed Bars must be continuous per symbol and interval.

Duplicate identical Bars are idempotent.

Changed duplicate Bars fail.

Forming Bars never enter Signaler state.

BBO requires positive bid and ask with bid not above ask.

Source timestamp cannot move backward.

Arrival timestamp is recorded separately.

New cycles and new Orders require a BBO within the approved freshness threshold.

STOP waits bounded time for a fresh BBO when cleanup needs price.

Failure to obtain freshness stops new mutations and preserves claims.

The exact threshold belongs in AppConfig and must be proven against observed testnet traffic.

## Exclusive Account-Symbol Claims

Claim identity is:

```text
venue / execution_network / physical_account_id / symbol
```

Direction does not change identity.

Runner acquires every claim before Venue initialization.

Claims use atomic create-only files under:

```text
workspace/claims/account-symbol/<sha256-resource-key>.claim
```

Claim content contains schema version, exact resource, bot ID, generation ID, and creation time.

Claim content contains no credential.

Partial multi-resource acquisition releases only files created by that attempt before mutation capability exists.

Once mutation capability exists, abnormal exit retains every claim.

Graceful release requires terminal persistence, zero Orders, zero position, and complete closure fees.

Crash-retained claims fail closed.

No stale-age heuristic removes claims.

No Server or process liveness guess overrides claims.

## BalancedRisk Limitation

BalancedRisk always returns `Allow`.

It does not evaluate balances, margin, equity, drawdown, liquidation distance, or market conditions.

It is shared by Bt and live.

This plan adds no Risk behavior.

Simnet and testnet status, logs, Results, and operator instructions must state this limitation.

Testnet safety gates are admission and execution safety.

They are not trading Risk.

# Target Change #1

## Objective

Freeze current BtBot behavior and create deterministic shared-core comparison proof.

## Before

- `stest.sh` runs fresh BtBot processes and validates result databases.
- `internal/btbot/btbot.go` owns replay proof and terminal publication.
- No recorded source-neutral event fixture compares Bt and live drivers.
- Timing and memory fields vary between runs.

## New

- Add one semantic BtBot baseline manifest.
- Add one deterministic recorded-event fixture covering Bars, BBOs, heartbeats, user hints, STOP, and Results.
- Compare domain values while excluding declared nondeterministic timing and memory.
- Create exact before and after evidence directories.

## Reason

Every shared-core change needs a stable regression oracle before live work starts.

## Canonical Owner

BtBot owns replay proof.

Tests own deterministic comparison.

The proof harness owns artifact capture only.

## Exact Affected Files

- `internal/btbot/btbot_test.go`
- `internal/controller/controller_test.go`
- `internal/controller/testdata/runner_recording.json`
- `stest.sh`
- `wiki/design/packages/btbot.md`
- `wiki/design/packages/controller.md`

## Upstream Impact

- Uses current stored Sweep and Bot fixtures.
- Uses current Parquet and OHLCV paths unchanged.

## Downstream Impact

- Every later shared target depends on this oracle.
- Runner parity uses the same recording.

## Cross-Cutting Impact

- Lifecycle: no production change.
- State: baseline captures exact semantic state.
- Concurrency: deterministic fixture stays synchronous.
- Persistence: checks SQLite integrity and semantic rows.
- Telemetry: compares ordered stable fields.
- Shutdown: captures STOP order and flat proof.
- Security: fixture contains no credential.

## Surgical Steps

1. Capture current `HEAD`, dirty-file inventory, Bot identities, and current result paths.
2. Run one Observer and one trading baseline before edits.
3. Store semantic summaries beneath the exact proof root.
4. Add one recorded-event fixture using public Controller behavior.
5. Add one comparator excluding only named runtime metrics.
6. Keep production packages unchanged.

## Tests and Exact Proof

Automated:

- `./stest.sh -bot 9 -runs 1`
- One current Grid or Trade baseline selected from canonical datastore evidence.
- `go test -count=1 -tags noasm ./internal/btbot ./internal/controller`
- SQLite `integrity_check` and `foreign_key_check`.

Artifacts:

```text
workspace/proof/nuubot-runner/tc01/before/manifest.txt
workspace/proof/nuubot-runner/tc01/before/observer-run.json
workspace/proof/nuubot-runner/tc01/before/trading-run.json
workspace/proof/nuubot-runner/tc01/recorded-events/result.json
workspace/proof/nuubot-runner/tc01/recorded-events/domain.json
```

Manual: none.

Deferred: live driver comparison waits for Target Change #10.

## Non-Goals

- No new runtime abstraction.
- No replay behavior change.
- No performance threshold.

## Stop and Reassess

- Baseline cannot pass current code.
- Canonical trading fixture is unavailable.
- Semantic fields cannot be separated from nondeterministic fields.
- Test work would require production behavior changes.

# Target Change #2

## Objective

Create one source-neutral closed-Bar calculation core while preserving current BtBot output.

## Before

- `signaler.Create` accepts symbol, Parquet source, start, and end.
- `macross.Init` and `rsi.Init` load complete OHLCV.
- `signaler.loadSeries` calls `ohlcv.Load`.
- No `market.Bar` exists.
- Signaler exposes only `Signals` and `Stop`.

## New

- `market.Bar` owns one validated symbol, interval, start, OHLCV, and closed boundary.
- OHLCV converts validated Parquet rows into `market.Bar`.
- Signaler exposes requirements, `Bootstrap`, completed-Bar ingestion, `Signals`, and `Stop`.
- Macross and RSI use one calculation path for batch bootstrap and incremental Bars.
- BtBot still loads current Parquet ranges before Controller starts.
- Remove Signaler filesystem and Parquet knowledge.

## Reason

Bt and live need identical calculations from different data boundaries.

## Canonical Owner

- `internal/market` owns trusted Bar values.
- `internal/ohlcv` owns Parquet decoding.
- `internal/signaler` owns calculation state.
- BtBot owns historical bootstrap acquisition.

## Exact Affected Files

- `internal/market/market.go`
- `internal/market/market_test.go`
- `internal/ohlcv/ohlcv.go`
- `internal/ohlcv/ohlcv_test.go`
- `internal/signaler/signaler.go`
- `internal/signaler/macross.go`
- `internal/signaler/rsi.go`
- `internal/signaler/signaler_test.go`
- `internal/botspec/build.go`
- `internal/btbot/btbot.go`
- Existing affected tests
- `wiki/design/packages/market.md`
- `wiki/design/packages/ohlcv.md`
- `wiki/design/packages/signaler.md`
- `wiki/design/concepts/macross-signaler.md`
- `wiki/design/concepts/rsi-signaler.md`

## Upstream Impact

- OHLCV remains the Bt Parquet boundary.
- BotSpec supplies configuration, not data paths.

## Downstream Impact

- Controller can later ingest live completed Bars.
- Runner can bootstrap the same Signaler without replay types.

## Cross-Cutting Impact

- Lifecycle: Signaler bootstrap completes before Start.
- State: incremental state must equal batch state.
- Concurrency: Signaler remains synchronous.
- Persistence: none added.
- Telemetry: Bar counts and last closed boundary become primitive facts.
- Shutdown: Stop remains idempotent.
- Security: no impact.

## Surgical Steps

1. Add strict `market.Bar` validation.
2. Convert OHLCV output at the boundary.
3. Expose Signaler requirements without loading data.
4. Move existing calculations behind shared bootstrap and append methods.
5. Make BtBot load exact current ranges.
6. Delete Signaler path and time-range parameters.
7. Delete old loaders after all callers move.
8. Preserve Package ordering and availability timestamps.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/market ./internal/ohlcv ./internal/signaler ./internal/botspec ./internal/btbot`
- Batch versus one-Bar-at-a-time Package equality.
- Closed final Bar equality.
- Gap, duplicate-change, forming-Bar, and warmup rejection.

Artifacts:

```text
workspace/proof/nuubot-runner/tc02/before/manifest.txt
workspace/proof/nuubot-runner/tc02/after/manifest.txt
workspace/proof/nuubot-runner/tc02/signaler-parity.json
```

Manual: none.

Deferred: live source proof waits for Target Change #9.

## Non-Goals

- No new indicator.
- No alternate Parquet loader.
- No retained live history beyond Signaler requirements.

## Stop and Reassess

- Incremental calculations differ from current packages.
- OHLCV closure semantics change.
- BtBot semantic proof changes.
- Bar ownership causes an import cycle.

# Target Change #3

## Objective

Separate Bt and live admission without changing stored BotSpec meaning.

## Before

- `datastore.Bot` contains replay fields.
- `datastore.LoadBot` parses replay JSON.
- `setup.Setup` always resolves replay paths.
- `setup.Setup` always fetches mainnet Meta.
- `botspec.Build` reads `Bot.Replay`.
- AppConfig lacks Runner cadence and freshness.

## New

- `datastore.Bot` contains exact shared Bot identity and Config only.
- `datastore.BtInput` contains `Bot` plus `ReplayInput`.
- `LoadBtBot` preserves `(sweep_id, bot_id)` and current replay parsing.
- `LoadLiveBot` loads one globally unique `bot_id` without replay parsing.
- `setup.Bt` preserves current Bt setup behavior.
- `setup.Live` admits one live Bot, derived data network, resources, paths, and credential references.
- AppConfig adds only Runner heartbeat, request timeout, bootstrap timeout, and freshness limits.
- Process mode stays in outer setup calls.

## Reason

Replay data is process input, not BotSpec or strategy identity.

## Canonical Owner

- Datastore owns stored input decoding.
- Setup owns external admission.
- Config owns operational limits.
- BotSpec owns strategy validation.

## Exact Affected Files

- `internal/datastore/models.go`
- `internal/datastore/sweep.go`
- `internal/datastore/sweep_test.go`
- `internal/setup/setup.go`
- `internal/setup/setup_test.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/botspec/build.go`
- `internal/botspec/botspec_test.go`
- `internal/btbot/btbot.go`
- `workspace/config/config.toml`
- `wiki/design/packages/datastore.md`
- `wiki/design/packages/setup.md`
- `wiki/design/packages/config.md`
- `wiki/design/concepts/bot-spec.md`

## Upstream Impact

- Bt command still supplies Sweep and Bot IDs.
- Runner command later supplies only Bot ID.
- Stored BotConfig hash remains exact.

## Downstream Impact

- BotDefinition becomes data-source neutral.
- Runner can select live inputs without replay adapters.

## Cross-Cutting Impact

- Lifecycle: both admissions finish before Controller construction.
- State: no BotSpec identity change.
- Concurrency: Setup starts no goroutine.
- Persistence: live load remains read-only.
- Telemetry: no change.
- Shutdown: no change.
- Security: live admission carries references, never secrets.

## Surgical Steps

1. Split shared Bot identity from Bt replay input.
2. Rename the current loader to `LoadBtBot`.
3. Add one read-only live loader.
4. Split Setup entry points.
5. Move replay path validation into `setup.Bt`.
6. Derive live data network from execution network.
7. Reject live mainnet before credentials.
8. Keep exact BotConfig decoding in BotSpec.
9. Delete old combined APIs after callers move.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/datastore ./internal/setup ./internal/config ./internal/botspec ./internal/btbot`
- Bt admission equality.
- Live load ignores replay-only JSON.
- Simnet derives mainnet data.
- Testnet derives testnet data.
- Mainnet rejects.

Artifacts:

```text
workspace/proof/nuubot-runner/tc03/before/manifest.txt
workspace/proof/nuubot-runner/tc03/after/manifest.txt
workspace/proof/nuubot-runner/tc03/admission-cases.json
```

Manual: none.

Deferred: live network access.

## Non-Goals

- No database schema migration.
- No status write.
- No credential loading.
- No Server dependency.

## Stop and Reassess

- `bot_id` is not globally unique.
- Live Config requires replay-only fields.
- Testnet instrument identity cannot be separated from mainnet Meta.
- Bt result parity changes.

# Target Change #4

## Objective

Hardcut `market.BBO` from one price to bid and ask.

## Before

- `market.BBO` contains `Price`.
- Replay creates BBO from one close.
- Simulator uses one price for crossings and marks.
- Executors use `BBO.Price`.
- Docs describe side-specific behavior not present in code.

## New

- `market.BBO` contains symbol, source timestamp, bid, and ask.
- Validation requires positive finite values and `bid <= ask`.
- `Midpoint` is the canonical mark and strategy reference.
- Buy execution uses ask.
- Sell execution uses bid.
- Replay sets `bid = ask = close`.
- Remove `Price` after all callers move.

## Reason

Live execution and Simulator parity require real best bid and ask semantics.

## Canonical Owner

- Market owns the trusted BBO.
- Simulator owns crossing and Fill-price behavior.
- Executors consume explicit reference prices.

## Exact Affected Files

- `internal/market/market.go`
- `internal/market/market_test.go`
- `internal/replay/replay.go`
- `internal/simulator/simulator.go`
- `internal/simulator/simulator_test.go`
- `internal/account/account.go`
- `internal/account/account_test.go`
- `internal/executor/observer.go`
- `internal/executor/trade.go`
- `internal/executor/grid.go`
- `internal/executor/*_test.go`
- `internal/botcycle/botcycle.go`
- `internal/controller/controller.go`
- Existing affected tests
- `wiki/design/packages/market.md`
- `wiki/design/packages/simulator.md`
- `wiki/design/concepts/replay.md`
- `wiki/design/concepts/ingestbbo.md`

## Upstream Impact

- Replay close remains the exact historical value.
- Future WebSocket decoder supplies real bid and ask.

## Downstream Impact

- Account marks use midpoint.
- Simulator Fill prices become side-correct.
- Existing Bt behavior remains exact because bid equals ask.

## Cross-Cutting Impact

- Lifecycle: none.
- State: transient BBO shape changes.
- Concurrency: immutable value remains copy-safe.
- Persistence: BBO remains transient.
- Telemetry: freshness may record both prices later.
- Shutdown: cleanup pricing becomes explicit.
- Security: none.

## Surgical Steps

1. Replace the BBO constructor and fields.
2. Convert replay close into equal bid and ask.
3. Replace every `Price` caller with midpoint, bid, or ask by meaning.
4. Update Simulator crossing and Fill price.
5. Delete the old field and constructor shape.
6. Search for every stale one-price assumption.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/market ./internal/replay ./internal/simulator ./internal/account ./internal/executor ./internal/botcycle ./internal/controller ./internal/btbot`
- Spread crossing tests.
- Midpoint mark tests.
- Equal-price replay parity.

Artifacts:

```text
workspace/proof/nuubot-runner/tc04/before/manifest.txt
workspace/proof/nuubot-runner/tc04/after/manifest.txt
workspace/proof/nuubot-runner/tc04/bbo-cases.json
```

Manual: none.

Deferred: external spread parity.

## Non-Goals

- No order-book depth.
- No trade tape.
- No spread strategy.

## Stop and Reassess

- Any Bt finance or lifecycle value changes.
- A caller lacks a clear bid, ask, or midpoint meaning.
- Simulator parity requires deeper market data.

# Target Change #5

## Objective

Add source-neutral Controller Bar ingestion, dirty hints, freshness facts, and runtime-only inputs.

## Before

- Controller only ingests BBO.
- Controller owns the Signaler but cannot feed Bars.
- Controller cannot mark an Account dirty.
- BotDefinition contains no runtime credential inputs.
- Controller has no freshness gate.

## New

- `Controller.IngestBar` forwards one validated completed Bar to Signaler.
- `Controller.MarkAccountDirty` routes one exact resource hint through BotCycle and Executor.
- Controller records last accepted Bar and BBO timestamps.
- `Controller.Run` blocks new cycle and mutation work when required market data is stale.
- Runtime-only inputs carry an optional credential catalog without entering BotDefinition or Results.
- Bt passes empty runtime inputs.

## Reason

Live sources need safe shared-core entry points without transport ownership or secret persistence.

## Canonical Owner

- Controller owns synchronous policy admission.
- Signaler owns Bar calculation.
- Account owns dirty state.
- Runner owns freshness configuration and credentials loading.

## Exact Affected Files

- `internal/controller/controller.go`
- `internal/controller/controller_test.go`
- `internal/botcycle/botcycle.go`
- `internal/botcycle/botcycle_test.go`
- `internal/executor/executor.go`
- `internal/executor/trade.go`
- `internal/executor/grid.go`
- Existing affected tests
- `internal/btbot/btbot.go`
- `wiki/design/packages/controller.md`
- `wiki/design/packages/botcycle.md`
- `wiki/design/packages/executor.md`
- `wiki/design/packages/signaler.md`
- `wiki/design/concepts/live-events.md`

## Upstream Impact

- Bt bootstrap still completes before Controller Start.
- Runner later supplies live runtime inputs.

## Downstream Impact

- Account receives exact credential candidates only when constructed.
- User events can mark only the matching active Account dirty.

## Cross-Cutting Impact

- Lifecycle: ingress rejects invalid Controller state.
- State: adds timestamps and opaque runtime inputs.
- Concurrency: methods remain synchronous.
- Persistence: secrets remain absent.
- Telemetry: freshness facts become observable.
- Shutdown: dirty hints reject after admission closes.
- Security: no secret formatting or Result fields.

## Surgical Steps

1. Add Bar ingestion with strict lifecycle checks.
2. Add exact resource dirty routing.
3. Add freshness facts without transport types.
4. Add concrete runtime-input values passed through ownership layers.
5. Keep BotDefinition immutable and secret-free.
6. Keep Bt call order unchanged.
7. Add no generic bus or callback registry.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/controller ./internal/botcycle ./internal/executor ./internal/btbot`
- Completed-Bar routing.
- Forming and stale Bar rejection.
- Exact dirty resource routing.
- Empty Bt credentials.
- Secret absence in Results and JSON.

Artifacts:

```text
workspace/proof/nuubot-runner/tc05/before/manifest.txt
workspace/proof/nuubot-runner/tc05/after/manifest.txt
workspace/proof/nuubot-runner/tc05/controller-ingress.json
```

Manual: none.

Deferred: external event supervision.

## Non-Goals

- No feed code.
- No Account selection change.
- No Risk change.

## Stop and Reassess

- Credentials would enter BotDefinition.
- Dirty routing exposes mutable Accounts.
- Freshness changes Bt decisions.
- Controller would import transport or datastore packages.

# Target Change #6

## Objective

Hardcut Account from concrete Simulator calls to one minimal Account-owned Venue.

## Before

- `Account` owns `simulator.Simulator`.
- `PlaceResult` exposes `simulator.SubmitResponse`.
- Account creates Simulator requests directly.
- Recon uses Simulator Order, Fill, and Account-state types.
- Stop calls Simulator directly.
- Result always contains Simulator evidence.
- No Venue interface exists in code.

## New

- `account.Venue` is the smallest current Account contract.
- Context reaches every blocking Venue operation.
- Hyperliquid owns admitted protocol request and response values.
- Simulator implements the same values without transport or signing.
- Account holds only `Venue` for runtime calls.
- Simulator terminal evidence remains optional through a narrow result-only capability.
- Hyperliquid Account results contain no fabricated Simulator result.
- Remove old Simulator protocol types and direct calls.

## Reason

One Account path prevents live and Simulator behavior drift.

## Canonical Owner

- Account owns the interface.
- Hyperliquid owns protocol values.
- Simulator owns simulated truth.
- Account owns protocol-to-domain translation.

## Exact Affected Files

- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/account/recon2.go`
- `internal/account/account_test.go`
- `internal/hyperliquid/orders.go`
- `internal/hyperliquid/fills.go`
- `internal/hyperliquid/state.go`
- `internal/simulator/simulator.go`
- `internal/simulator/simulator_test.go`
- `internal/simulator/store.go`
- `internal/simulator/publish.go`
- `internal/resultpublisher/resultpublisher.go`
- `internal/resultpublisher/resultpublisher_test.go`
- Existing affected tests
- `wiki/design/concepts/venue.md`
- `wiki/design/packages/account.md`
- `wiki/design/packages/simulator.md`
- `wiki/design/packages/hyperliquid.md`

## Upstream Impact

- Executor still owns Account.
- Account initialization receives the selected network and credential candidates.

## Downstream Impact

- Ledger receives one normalized evidence shape.
- Hyperliquid can implement the same interface later.
- ResultPublisher handles absent Simulator evidence explicitly.

## Cross-Cutting Impact

- Lifecycle: Venue Init and Stop match Account lifetime.
- State: one canonical protocol shape.
- Concurrency: contexts bound live calls.
- Persistence: Simulator persistence remains owned.
- Telemetry: query timing can wrap interface calls.
- Shutdown: Venue stops before Ledger.
- Security: Venue errors must be redacted.

## Surgical Steps

1. Define the interface beside Account.
2. Move shared protocol values into Hyperliquid.
3. Adapt Simulator methods to the exact contract.
4. Replace every Account direct Simulator call.
5. Replace reconciliation concrete types.
6. Keep optional Simulator Result outside the Venue operation set.
7. Delete moved Simulator types.
8. Delete compatibility aliases and bridges.
9. Search for every remaining concrete runtime call.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/account ./internal/simulator ./internal/hyperliquid ./internal/executor ./internal/resultpublisher ./internal/btbot`
- Interface compile assertions.
- Context cancellation.
- Optional Simulator Result behavior.
- Existing delayed-fee and STOP tests.

Artifacts:

```text
workspace/proof/nuubot-runner/tc06/before/manifest.txt
workspace/proof/nuubot-runner/tc06/after/manifest.txt
workspace/proof/nuubot-runner/tc06/venue-contract.json
```

Manual: none.

Deferred: Hyperliquid implementation.

## Non-Goals

- No `internal/venue` package.
- No Venue factory framework.
- No testnet mutation.
- No second Account path.

## Stop and Reassess

- Interface gains a method without a current Account caller.
- Protocol ownership creates an import cycle.
- Simulator and Hyperliquid need incompatible semantics.
- BtBot proof changes.

# Target Change #7

## Objective

Remove repeated Simulator gates and finish one canonical simnet admission path.

## Before

- BotSpec requires `simulator/simnet`.
- TradeExecutor repeats that requirement.
- GridExecutor repeats that requirement.
- Account repeats that requirement.
- Capability comments say Simulator-only.
- Venue selection is not canonical.

## New

- BotSpec owns static resource pairing validation.
- Initially allowed pairing is only `simulator/simnet`.
- Executors validate execution policy, not Venue selection.
- Account selects Simulator from the admitted resource.
- Account rejects unsupported runtime selection once.
- Simulator requires no credential catalog.
- Capability names describe Venue behavior.
- No parallel legacy gate remains.

## Reason

One canonical gate is smaller and prevents inconsistent live expansion.

## Canonical Owner

- BotSpec owns static Config combinations.
- Account owns runtime Venue selection.
- Executors own only execution policy.

## Exact Affected Files

- `internal/botspec/config.go`
- `internal/botspec/botspec_test.go`
- `internal/executor/executor.go`
- `internal/executor/trade.go`
- `internal/executor/grid.go`
- `internal/executor/*_test.go`
- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/account/account_test.go`
- `wiki/design/packages/botspec.md`
- `wiki/design/packages/executor.md`
- `wiki/design/packages/account.md`

## Upstream Impact

- Existing BotConfig keeps `simulator/simnet`.
- No template meaning changes.

## Downstream Impact

- Target Change #13 extends one BotSpec allowlist.
- Executor code stays network-neutral.

## Cross-Cutting Impact

- Lifecycle: unchanged.
- State: unchanged.
- Concurrency: unchanged.
- Persistence: unchanged.
- Telemetry: Venue identity remains explicit.
- Shutdown: one Account path.
- Security: simnet ignores credentials.

## Surgical Steps

1. Keep one static pairing validator in BotSpec.
2. Delete Executor Venue and network gates.
3. Select Simulator once in Account Init.
4. Delete repeated checks and stale comments.
5. Prove simnet never loads credentials.
6. Search for every `simulator` and `simnet` policy gate.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/botspec ./internal/executor ./internal/account ./internal/btbot`
- Unsupported pair rejection at BotSpec.
- Runtime unsupported selection rejection at Account.
- No Executor network branch.
- Simnet credential-loader tripwire.

Artifacts:

```text
workspace/proof/nuubot-runner/tc07/before/manifest.txt
workspace/proof/nuubot-runner/tc07/after/manifest.txt
workspace/proof/nuubot-runner/tc07/gate-inventory.txt
```

Manual: none.

Deferred: testnet allowlist.

## Non-Goals

- No Hyperliquid implementation.
- No new BotSpec.
- No Config aliases.

## Stop and Reassess

- Executor behavior depends on execution network.
- Account needs a second path.
- Any legacy gate must remain for compatibility.

# Target Change #8

## Objective

Add standalone Account-symbol claims, live generation evidence, and first-release crash fencing.

## Before

- No `internal/runner` package exists.
- No cross-process Account-symbol claim exists.
- No live generation or lifecycle database exists.
- Account and Simulator child persistence cannot resume Controller policy.
- Recovery is unapproved.

## New

- Add atomic create-only claim files.
- Add one random collision-resistant generation ID using `crypto/rand`.
- Add one Runner-owned generation database.
- Persist monotonic lifecycle states and heartbeat rows.
- Require live `persist_mode = max`.
- Retain claims after crash, persistence ambiguity, unknown mutation, or failed STOP.
- Refuse fresh start when any claim or nonterminal generation conflicts.
- Implement no recovery.

## Reason

First release must prevent duplicate ownership and fail closed after crash without Server.

## Canonical Owner

- Runner owns claims and live lifecycle evidence.
- Account owns domain persistence.
- Simulator owns Simulator state.
- Server ProcessStore remains out of scope.

## Exact Affected Files

- `internal/runner/claims.go`
- `internal/runner/claims_test.go`
- `internal/runner/store.go`
- `internal/runner/store_test.go`
- `internal/runner/result.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/botspec/config.go`
- `internal/botspec/botspec_test.go`
- `wiki/design/concepts/filesystem.md`
- `wiki/design/concepts/runner.md`
- `wiki/design/concepts/recovery.md`
- `wiki/design/telemetry.md`

## Upstream Impact

- Live admission supplies exact resources and Bot ID.
- Workspace root supplies claim and database directories.

## Downstream Impact

- Runner Init acquires claims and creates the generation.
- Account receives the generation database path.
- Recovery target later consumes retained evidence.

## Cross-Cutting Impact

- Lifecycle: states are `starting`, `running`, `stopping`, `stopped`, and `error`.
- State: claims are durable ownership fences.
- Concurrency: atomic file creation chooses one winner.
- Persistence: SQLite uses foreign keys and integrity checks.
- Telemetry: one heartbeat row per heartbeat.
- Shutdown: claims release last.
- Security: files contain no secret.

## Surgical Steps

1. Define canonical resource-key serialization.
2. Hash it with SHA-256 for the filename.
3. Atomically create each claim with owner-only permissions.
4. Sync and close each created file.
5. Create a random generation ID.
6. Create one generation database below `workspace`.
7. Write conditional monotonic lifecycle transitions.
8. Add heartbeat persistence with stable indexed identity.
9. Keep abnormal claims.
10. Release only exact owned claims after successful terminal proof.
11. Add no stale timeout or PID guessing.

## Tests and Exact Proof

Automated:

- `go test -count=1 -tags noasm ./internal/runner ./internal/config ./internal/botspec`
- Concurrent claim race has one winner.
- Partial acquisition unwinds before mutation capability.
- Wrong generation cannot release a claim.
- Crash simulation retains claims.
- Fresh start rejects retained claims.
- Lifecycle transitions reject invalid prior state.
- SQLite integrity and foreign-key checks.
- Permissions and secret scan.

Artifacts:

```text
workspace/proof/nuubot-runner/tc08/claim-race.json
workspace/proof/nuubot-runner/tc08/crash-fence.json
workspace/proof/nuubot-runner/tc08/runner-db-integrity.txt
workspace/proof/nuubot-runner/tc08/secret-scan.txt
```

Manual:

- Verify claim behavior on Windows local NTFS.
- Verify claim behavior on Linux local filesystem.

Deferred:

- Network filesystems.
- Claim takeover.
- Recovery.

## Non-Goals

- No ProcessStore.
- No PID liveness inference.
- No automatic cleanup.
- No database migration.

## Stop and Reassess

- `O_CREATE|O_EXCL` cannot prove atomicity on supported local filesystems.
- Live persistence needs Server ownership.
- Multiple writers cannot preserve SQLite integrity.
- A claim would be released without flat proof.

# Target Change #9

## Objective

Add one process-local public live market boundary for exact closed Bars and BBOs.

## Before

- Hyperliquid has bounded REST for Meta and clearinghouse state.
- No candle bootstrap call exists.
- No WebSocket exists.
- No live Bar or BBO decoder exists.
- No process-local feed lifecycle exists.

## New

- Add bounded candle bootstrap through Hyperliquid public REST.
- Add one process-local public WebSocket supervisor.
- Use `gorilla/websocket`, as approved by canonical design.
- Decode typed BBO and completed-Bar events.
- Simnet selects mainnet public endpoints.
- Testnet selects testnet public endpoints.
- Constructors perform no I/O.
- Start owns connection and subscriptions.
- Stop cancels, closes, and joins every goroutine.
- Callbacks only enqueue typed events.

## Reason

Runner must operate while Server is stopped and must not mix feed work with policy.

## Canonical Owner

- Hyperliquid owns protocol transport and decoding.
- Runner owns subscriptions and event consumption.
- Market owns admitted values.

## Exact Affected Files

- `internal/hyperliquid/client.go`
- `internal/hyperliquid/candles.go`
- `internal/hyperliquid/candles_test.go`
- `internal/hyperliquid/websocket.go`
- `internal/hyperliquid/websocket_test.go`
- `internal/hyperliquid/market.go`
- `internal/hyperliquid/market_test.go`
- `go.mod`
- `go.sum`
- `wiki/design/hyperliquid/rest.md`
- `wiki/design/hyperliquid/websocket.md`
- `wiki/design/packages/hyperliquid.md`
- `wiki/design/concepts/live-events.md`

## Upstream Impact

- Setup supplies explicit data network and timeouts.
- Signaler supplies exact Bar requirements.

## Downstream Impact

- Runner receives validated market events.
- Simulator receives mainnet BBOs during simnet.

## Cross-Cutting Impact

- Lifecycle: explicit Init, Start, Err, Stop.
- State: desired subscriptions survive reconnect.
- Concurrency: one supervisor, one reader, one serialized writer.
- Persistence: none.
- Telemetry: connection, reconnect, last-source, and last-arrival facts.
- Shutdown: context interrupts reads and backoff.
- Security: public calls load no credential.

## Surgical Steps

1. Verify current official public REST and WebSocket contracts.
2. Add the approved pure-Go WebSocket dependency.
3. Implement strict bounded candle decoding.
4. Prove exact closed bootstrap boundaries.
5. Implement one socket supervisor.
6. Restore subscriptions once after reconnect.
7. Run callbacks outside locks.
8. Reject forming Bars.
9. Emit no policy calls.
10. Join all owned goroutines on Stop.

## Tests and Exact Proof

Automated:

- `go test -count=1 -tags noasm ./internal/hyperliquid ./internal/market`
- Oversized, malformed, wrong-symbol, wrong-network, and forming-Bar rejection.
- Missing Pong and read timeout reconnect.
- Subscription restore once.
- Cancellation interrupts read and backoff.
- Slow callback cannot deadlock Stop.
- Goroutine count returns to baseline.
- `go test -race -count=1 -tags noasm ./internal/hyperliquid`

Artifacts:

```text
workspace/proof/nuubot-runner/tc09/candle-bootstrap.json
workspace/proof/nuubot-runner/tc09/websocket-reconnect.json
workspace/proof/nuubot-runner/tc09/goroutine-leak.txt
workspace/proof/nuubot-runner/tc09/race.txt
```

Opt-in network:

- Mainnet public bootstrap and BBO observation only.
- Testnet public bootstrap and BBO observation only.
- No private request.

Manual: inspect one captured closed-Bar boundary.

Deferred: shared subscriptions and retention.

## Non-Goals

- No shared DataEngine.
- No private mutation.
- No forming Bar delivery.
- No unbounded reconnect.

## Stop and Reassess

- Official contract differs from canonical docs.
- Testnet lacks required public data.
- A callback must hold transport locks.
- A goroutine survives Stop.
- A dependency requires CGO, unsafe, or native code.

# Target Change #10

## Objective

Implement the new simnet `internal/runner` composition root and serialized event lane.

## Before

- `internal/runner` has no Runner implementation.
- No event lane, supervision loop, feed state, heartbeat, or live Result exists.
- WallClock callbacks could call policy directly.
- Simnet live execution does not exist.

## New

- Add one new `Runner` with `Init`, `Start`, `Loop`, `Stop`, and `Result`.
- Runner owns one WallClock, Controller, public feed, event lane, claims, store, telemetry, and cancellation.
- Feed and Clock callbacks enqueue only.
- One Loop owns every Controller call.
- Startup buffers subscriptions before bootstrap.
- Simnet uses mainnet public data and Simulator Venue.
- One heartbeat drives Controller and telemetry.
- Child failure reaches Runner.
- Stop follows the canonical order.

## Reason

This is the smallest Server-independent live process owner.

## Canonical Owner

`internal/runner` owns live orchestration only.

Shared packages retain trading policy.

## Exact Affected Files

- `internal/runner/runner.go`
- `internal/runner/events.go`
- `internal/runner/result.go`
- `internal/runner/runner_test.go`
- `internal/runner/recorded_test.go`
- `internal/runner/testdata/recorded_events.json`
- `internal/toolkit/clock/wallclock.go`
- `internal/toolkit/clock/clock_test.go`
- `internal/controller/controller.go`
- Existing affected tests
- `wiki/design/concepts/runner.md`
- `wiki/design/concepts/live-events.md`
- `wiki/design/concepts/wall-clock.md`
- `wiki/design/concepts/shutdown.md`
- `wiki/ARCHITECTURE.md`

## Upstream Impact

- Setup supplies live admission.
- Hyperliquid supplies public feed values.
- Claim and store components must already pass.

## Downstream Impact

- Command gains one complete lifecycle object.
- Simnet becomes directly executable.
- Testnet later reuses the same Runner.

## Cross-Cutting Impact

- Lifecycle: one explicit owner and reverse unwind.
- State: Runner-local latest feed facts and sequence.
- Concurrency: one policy lane and owned transport goroutines.
- Persistence: lifecycle and heartbeat rows.
- Telemetry: one row per heartbeat, success or failure.
- Shutdown: idempotent and cross-platform cancellable.
- Security: simnet loads no credentials.

## Surgical Steps

1. Construct a stopped Runner.
2. Bind one fixed bounded event channel.
3. Initialize claims, generation store, feed, Signaler data, Controller, and WallClock.
4. Start feed buffering before bootstrap.
5. Merge bootstrap and buffered completed Bars.
6. Require one fresh BBO.
7. Start Controller.
8. Open event admission.
9. Register one heartbeat.
10. Start WallClock last.
11. Supervise feed, Clock, stop request, and lane.
12. Process all Controller calls on Loop ownership.
13. Stop in exact reverse ownership order.
14. Return immutable Result after Stop.

## Tests and Exact Proof

Automated:

- `go test -count=1 -tags noasm ./internal/runner ./internal/controller ./internal/toolkit/clock`
- Target Change #1 recorded-event parity.
- Initial truth precedes Controller Start.
- Callback never calls policy.
- BBO, Bar, user, heartbeat, and stop FIFO proof.
- Lane overflow fails.
- Feed failure reaches Loop.
- WallClock failure reaches Loop.
- Partial Init and Start unwind.
- Stop idempotence.
- Result unavailable before Stop.
- Simnet credential-loader tripwire.
- `go test -race -count=1 -tags noasm ./internal/runner ./internal/hyperliquid`
- Goroutine and subscription leak proof.

Artifacts:

```text
workspace/proof/nuubot-runner/tc10/recorded-driver-parity.json
workspace/proof/nuubot-runner/tc10/startup-order.json
workspace/proof/nuubot-runner/tc10/event-order.json
workspace/proof/nuubot-runner/tc10/race.txt
workspace/proof/nuubot-runner/tc10/goroutine-leak.txt
```

Manual: none.

Deferred: real network process proof to Target Change #12.

## Non-Goals

- No command implementation.
- No testnet Venue.
- No automatic restart.
- No shared sockets.

## Stop and Reassess

- Any Controller call occurs outside Loop ownership.
- Bootstrap and subscription gap cannot be closed.
- Fixed lane capacity cannot survive measured traffic.
- Bt recorded parity changes.
- Stop leaks a goroutine or claim.

# Target Change #11

## Objective

Implement the new Runner command using BtBot command clarity and lifecycle shape.

## Before

- `cmd/nuubot-runner/main.go` prints `Under Construction.`.
- It has no parse, logging, lifecycle, signal, cleanup, or Result behavior.
- This placeholder is not old Runner behavior.

## New

- Parse `nuubot-runner <bot_id> [--allow-testnet]`.
- Open server log before identity.
- Open Bot log after valid identity.
- Create cancellation from OS signals.
- Run Runner `Init`, `Start`, `Loop`, `Stop`, and `Result` in visible order.
- Preserve primary errors with `errors.Join`.
- Log exact failed boundary.
- Exit nonzero on every failure.
- Handle Result only after successful Stop.
- Keep Section 1, Section 2, and Section 3 layout.

## Reason

The command needs one clear lifecycle shell, not replay behavior or hidden helpers.

## Canonical Owner

- Command owns parsing, logging, OS cancellation, lifecycle calls, and exit status.
- Runner owns live behavior.

## Exact Affected Files

- `cmd/nuubot-runner/main.go`
- `cmd/nuubot-runner/main_test.go`
- `internal/toolkit/logging/logging.go`
- `internal/toolkit/logging/logging_test.go`
- `build.sh`
- `wiki/bin.md`
- `wiki/design/concepts/runner.md`
- `wiki/DESIGN.md`

## Upstream Impact

- Receives one Bot ID and optional testnet authority.
- Uses existing logging conventions.

## Downstream Impact

- Operators can run simnet directly with Server stopped.
- Target Change #15 adds opt-in testnet proof.

## Cross-Cutting Impact

- Lifecycle: explicit phase ordering.
- State: no command-owned domain state.
- Concurrency: signal cancellation only.
- Persistence: delegated to Runner.
- Telemetry: delegated to Runner.
- Shutdown: cleanup always attempted after ownership begins.
- Security: parse errors reveal no credential.

## Surgical Steps

1. Use `cmd/nuubot-bt-bot/main.go` as the clarity example.
2. Copy program-flow ordering, parse boundary, section layout, and error joining.
3. Copy no replay, profiling, Sweep, RunReport, or stdout behavior.
4. Add strict positive Bot ID parsing.
5. Add exact optional testnet flag parsing.
6. Use `signal.NotifyContext`.
7. Call every Runner phase explicitly.
8. Join Loop and Stop errors.
9. Request Result only after clean Stop.
10. Log one terminal result.

## Tests and Exact Proof

Automated:

- `go test -count=1 -tags noasm ./cmd/nuubot-runner ./internal/toolkit/logging`
- Invalid input writes server log and exits nonzero.
- Valid identity switches to Bot log.
- Every lifecycle failure names its phase.
- Loop error still calls Stop.
- Result error returns nonzero.
- SIGINT cancellation reaches Stop.
- Testnet Bot without flag fails before credentials.
- Simnet with testnet flag fails.
- Build uses `CGO_ENABLED=0` and `-tags noasm`.

Artifacts:

```text
workspace/proof/nuubot-runner/tc11/command-cases.json
workspace/proof/nuubot-runner/tc11/sigint-stop.txt
workspace/proof/nuubot-runner/tc11/build.txt
```

Manual:

- Windows Ctrl+C direct process proof.
- Linux SIGINT and SIGTERM direct process proof.

Deferred: forced termination behavior.

## Non-Goals

- No profiling.
- No RunReport JSON.
- No Server control socket.
- No pause, resume, or restart.

## Stop and Reassess

- Command needs domain knowledge.
- Cleanup cannot preserve the primary error.
- Windows or Linux cannot reach the same Stop path.
- A replay-only behavior enters Runner.

# Target Change #12

## Objective

Deliver and prove live simnet before any testnet mutation work.

## Before

- Unit and recorded proofs exist.
- No accepted direct live Runner system proof exists.
- Testnet Venue remains absent.

## New

- Run one direct Server-independent simnet Runner.
- Use mainnet public Bars and BBO.
- Use Simulator execution.
- Load no credentials.
- Prove startup, steady state, STOP, persistence, crash fence, and exact domain evidence.
- Record fixed-width generated reports without hand formatting.

## Reason

Simnet proves shared live orchestration without private mutation risk.

## Canonical Owner

- Runner owns system proof.
- Existing test harness conventions own generated reporting.

## Exact Affected Files

- `runner_test.sh`
- `build.sh`
- `internal/runner/runner_test.go`
- `wiki/PROJECT.md`
- `wiki/ARCHITECTURE.md`
- `wiki/design/concepts/runner.md`
- `wiki/design/telemetry.md`
- `wiki/design/concepts/filesystem.md`

## Upstream Impact

- Requires stable mainnet public connectivity.
- Requires one safe stored simnet Bot.

## Downstream Impact

- Testnet implementation cannot start until this target passes.
- Baseline becomes the live orchestration regression oracle.

## Cross-Cutting Impact

- Lifecycle: complete direct process.
- State: durable generation and domain rows.
- Concurrency: race and leak proof.
- Persistence: integrity and crash fencing.
- Telemetry: heartbeat count and freshness.
- Shutdown: flat and order-free.
- Security: credentials path tripwire.

## Surgical Steps

1. Add one narrow simnet system harness.
2. Build Runner with required flags.
3. Launch Server-independent Runner.
4. Observe bootstrap and steady heartbeats.
5. Request graceful STOP.
6. Validate flat and order-free evidence.
7. Validate generation database.
8. Simulate crash after running.
9. Prove restart refuses retained claims.
10. Remove only test-created claims after manual flat verification.
11. Repeat direct clean run.

## Tests and Exact Proof

Automated:

- Focused unit and integration tests from Targets #1 through #11.
- `./stest.sh -bot 9 -runs 1`.
- `CGO_ENABLED=0 go test -count=1 -tags noasm ./...`.
- `CGO_ENABLED=0 go vet -tags noasm ./...`.
- `go test -race -count=1 -tags noasm ./internal/runner ./internal/hyperliquid`.
- Script syntax.
- Diagnostics.
- `git diff --check`.

Direct simnet:

- `./runner_test.sh -bot <approved_simnet_bot_id> -duration <approved_duration>`.

Artifacts:

```text
workspace/proof/nuubot-runner/tc12/simnet/run.log
workspace/proof/nuubot-runner/tc12/simnet/result.json
workspace/proof/nuubot-runner/tc12/simnet/telemetry.json
workspace/proof/nuubot-runner/tc12/simnet/domain.json
workspace/proof/nuubot-runner/tc12/simnet/sqlite-integrity.txt
workspace/proof/nuubot-runner/tc12/simnet/stop-proof.json
workspace/proof/nuubot-runner/tc12/crash/restart-refusal.txt
workspace/proof/nuubot-runner/tc12/security/credential-tripwire.txt
workspace/proof/nuubot-runner/tc12/race.txt
workspace/proof/nuubot-runner/tc12/leaks.txt
```

Manual:

- Confirm Server is stopped.
- Confirm no credentials file access.
- Confirm claim cleanup follows flat verification.

Deferred:

- Testnet.
- Recovery.

## Non-Goals

- No private request.
- No mainnet execution.
- No automatic crash cleanup.

## Stop and Reassess

- Simnet cannot run without credentials.
- Public feed cannot stay fresh.
- STOP leaves Orders or position.
- Crash permits duplicate start.
- Any BtBot semantic proof changes.

# Target Change #13

## Objective

Implement one Hyperliquid testnet Venue behind the existing Account contract.

## Before

- Hyperliquid supports public REST, Meta, and clearinghouse state.
- Signing is absent.
- Place, cancel, open Orders, Fills, status, and transient Account state are absent.
- BotSpec allows only Simulator simnet.
- Testnet credentials are decoded but weakly validated.

## New

- Implement official testnet signing from current official protocol.
- Implement ordered batch place and cancel.
- Implement open Orders, bounded Fills, exact Order status, and Account state.
- Implement the existing `account.Venue` exactly.
- Add `hyperliquid/testnet` to the one BotSpec pairing allowlist.
- Account selects exact testnet credentials.
- Constructors perform no I/O.
- Mutations never retry automatically.
- Mainnet private endpoint construction is impossible in this release.

## Reason

Testnet must reuse the proven Account path without Simulator branches.

## Canonical Owner

- Hyperliquid owns signing, wire protocol, transport, and response admission.
- Account owns Venue selection and domain translation.
- BotSpec owns allowed resource pairing.

## Exact Affected Files

- `internal/hyperliquid/sign.go`
- `internal/hyperliquid/sign_test.go`
- `internal/hyperliquid/exchange.go`
- `internal/hyperliquid/exchange_test.go`
- `internal/hyperliquid/orders.go`
- `internal/hyperliquid/orders_test.go`
- `internal/hyperliquid/fills.go`
- `internal/hyperliquid/fills_test.go`
- `internal/hyperliquid/client.go`
- `internal/hyperliquid/client_test.go`
- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/account/account_test.go`
- `internal/botspec/config.go`
- `internal/botspec/botspec_test.go`
- `internal/config/credentials.go`
- `internal/config/config_test.go`
- `go.mod`
- `go.sum`
- `wiki/design/hyperliquid.md`
- `wiki/design/hyperliquid/rest.md`
- `wiki/design/hyperliquid/exchange.md`
- `wiki/design/packages/hyperliquid.md`
- `wiki/design/packages/account.md`

## Upstream Impact

- Runner supplies explicit testnet opt-in, context, resource, and credential catalog.
- Official protocol vectors become required inputs.

## Downstream Impact

- Account receives real testnet acknowledgement and query evidence.
- Target Change #14 completes live reconciliation safety.

## Cross-Cutting Impact

- Lifecycle: Client Init and Stop match Account.
- State: no mutable domain object crosses Venue.
- Concurrency: requests use caller contexts and timeouts.
- Persistence: Account records intent before mutation.
- Telemetry: request kind, duration, status, and bounded error.
- Shutdown: cancellations and close Orders use the same Venue.
- Security: no secret, signature, or full private body logging.

## Surgical Steps

1. Re-read current official Hyperliquid signing and exchange contracts.
2. Select the smallest audited pure-Go cryptographic dependencies.
3. Reject CGO, unsafe, assembly, linkname, or native bindings.
4. Prove signing against official vectors before network calls.
5. Add strict wire request and response types.
6. Preserve request order.
7. Expand payload-wide errors to every item.
8. Classify transport ambiguity as unknown.
9. Implement bounded read queries.
10. Implement the Account Venue methods.
11. Extend one BotSpec allowlist.
12. Delete no Simulator proof.

## Tests and Exact Proof

Automated:

- `go test -count=1 -tags noasm ./internal/hyperliquid ./internal/account ./internal/botspec ./internal/config`
- Official signing vectors.
- Stable action hash.
- Ordered mixed batch outcomes.
- Payload-wide error expansion.
- Malformed and incomplete response rejection.
- Timeout classified unknown.
- Context cancellation.
- Mainnet construction rejection.
- Secret scan.

Opt-in testnet:

- Read-only open Orders, state, and exact status first.
- One bounded test Order batch after user-approved testnet account and symbol.
- Cancel every resting Order.

Artifacts:

```text
workspace/proof/nuubot-runner/tc13/signing/vectors.json
workspace/proof/nuubot-runner/tc13/exchange/response-cases.json
workspace/proof/nuubot-runner/tc13/testnet/read-only.json
workspace/proof/nuubot-runner/tc13/testnet/mutation.json
workspace/proof/nuubot-runner/tc13/security/secret-scan.txt
```

Manual:

- Confirm official vector provenance.
- Confirm approved testnet account and disposable symbol.

Deferred:

- Mainnet.
- Read retry policy.
- Leverage changes.

## Non-Goals

- No external Python client.
- No automatic mutation retry.
- No mainnet.
- No generic exchange SDK.

## Stop and Reassess

- Official signing vectors do not pass.
- Required library violates project restrictions.
- Testnet Meta identity is ambiguous.
- Response cannot preserve ordered outcomes.
- Any secret reaches output.

# Target Change #14

## Objective

Complete testnet user hints, unknown outcomes, capped Fill queries, delayed fees, freshness, and STOP safety.

## Before

- Simulator Recon proves core ordering.
- Hyperliquid Fill pagination is absent.
- Unknown live mutation evidence is absent.
- User events cannot mark Account dirty.
- Delayed-fee logic is proven only through Simulator fixtures.
- Testnet STOP proof is absent.

## New

- Subscribe exact testnet user events through the process-local socket.
- User callbacks enqueue only.
- Account dirty routing uses exact resource identity.
- REST Recon remains authoritative.
- Fill discovery splits capped windows and deduplicates inclusive boundaries.
- Pending-fee repair uses independent bounded windows.
- Unknown place and cancel outcomes block new mutations.
- Exact CLOID and Venue identity resolve ambiguity.
- Freshness gates apply before mutation and cleanup pricing.
- STOP waits for flat, order-free, and fee-complete closure evidence.

## Reason

Testnet acknowledgements and streams are hints. Reconciliation must preserve final truth and cleanup safety.

## Canonical Owner

- Hyperliquid owns user-event decoding and query pagination.
- Runner owns subscription and event routing.
- Account owns dirty state and reconciliation.
- Ledger owns durable domain evidence.
- Executor owns ordered cleanup.

## Exact Affected Files

- `internal/hyperliquid/websocket.go`
- `internal/hyperliquid/websocket_test.go`
- `internal/hyperliquid/fills.go`
- `internal/hyperliquid/fills_test.go`
- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/account/account_test.go`
- `internal/ledger/recon.go`
- `internal/ledger/ledger_test.go`
- `internal/executor/trade.go`
- `internal/executor/grid.go`
- `internal/executor/*_test.go`
- `internal/controller/controller.go`
- `internal/controller/controller_test.go`
- `internal/runner/runner.go`
- `internal/runner/runner_test.go`
- `internal/telemetry/telemetry.go`
- `internal/resultpublisher/resultpublisher.go`
- Existing affected tests
- `wiki/design/hyperliquid/exchange.md`
- `wiki/design/hyperliquid/websocket.md`
- `wiki/design/concepts/recon.md`
- `wiki/design/packages/account.md`
- `wiki/design/packages/ledger.md`
- `wiki/design/packages/executor.md`
- `wiki/design/telemetry.md`

## Upstream Impact

- Requires testnet Venue and event transport.
- Requires exact freshness configuration.

## Downstream Impact

- Controller receives only trusted Account snapshots.
- Runner telemetry exposes feed and reconciliation health.
- Testnet release proof can clean mutations safely.

## Cross-Cutting Impact

- Lifecycle: user subscriptions match Runner lifetime.
- State: unknown and pending-fee sets remain durable.
- Concurrency: callbacks enqueue only.
- Persistence: cursor moves only after complete accepted Recon.
- Telemetry: every physical Fill query recorded.
- Shutdown: fees and ambiguity can block claim release.
- Security: user payloads stay out of logs.

## Surgical Steps

1. Add strict user-event decoding.
2. Route exact resource dirty hints through the lane.
3. Implement capped discovery windows.
4. Implement independent repair windows.
5. Merge by Venue TID.
6. Preserve existing execution during fee enrichment.
7. Add explicit unknown mutation evidence.
8. Block new mutations while unknown remains.
9. Reconcile exact identity before cleanup decisions.
10. Apply freshness gates.
11. Keep claims on unresolved STOP.
12. Persist complete Recon telemetry.

## Tests and Exact Proof

Automated:

- Target Change #1 BtBot proof before and after.
- `go test -count=1 -tags noasm ./internal/hyperliquid ./internal/account ./internal/ledger ./internal/executor ./internal/controller ./internal/runner`
- Recorded-event driver parity.
- Lost, duplicated, delayed, and reordered user-event cases.
- Capped Fill-window splitting.
- Inclusive-boundary deduplication.
- Cursor-advanced delayed-fee enrichment.
- Unknown place and cancel outcomes.
- STOP waits for fees.
- Stale feed blocks mutation.
- `go test -race -count=1 -tags noasm ./internal/hyperliquid ./internal/runner`.

Opt-in testnet:

- Delay between filled acknowledgement and fee evidence.
- Disconnect user stream and prove REST repair.
- Timeout simulation followed by CLOID reconciliation.

Artifacts:

```text
workspace/proof/nuubot-runner/tc14/recon/fill-windows.json
workspace/proof/nuubot-runner/tc14/recon/delayed-fee.json
workspace/proof/nuubot-runner/tc14/recon/unknown-mutation.json
workspace/proof/nuubot-runner/tc14/feed/freshness.json
workspace/proof/nuubot-runner/tc14/stop/fee-complete.json
workspace/proof/nuubot-runner/tc14/race.txt
```

Manual: inspect bounded private evidence without copying payloads.

Deferred: automatic read retry and mainnet baselines.

## Non-Goals

- No WebSocket authority over Ledger.
- No inferred cancellation from absence.
- No automatic mutation retry.
- No mainnet.

## Stop and Reassess

- User stream identity cannot be proven.
- Fill caps cannot be split safely.
- Fee enrichment changes execution fields.
- STOP would release claims with unknown evidence.
- BtBot parity changes.

# Target Change #15

## Objective

Prove the complete opt-in testnet release and mutation cleanup.

## Before

- Testnet unit and focused integration proof exists.
- No accepted complete direct Runner testnet proof exists.

## New

- Run one explicitly authorized direct testnet Runner.
- Prove admission, feeds, mutation, reconciliation, telemetry, STOP, cleanup, persistence, and secrecy.
- Prove Server independence.
- Prove no mainnet private request.
- Preserve claims on every failed cleanup.

## Reason

Private mutation requires complete real-path proof before release.

## Canonical Owner

Runner owns end-to-end proof.

The testnet operator owns explicit mutation authority.

## Exact Affected Files

- `runner_test.sh`
- `internal/runner/runner_test.go`
- `wiki/PROJECT.md`
- `wiki/ARCHITECTURE.md`
- `wiki/design/concepts/runner.md`
- `wiki/design/hyperliquid/exchange.md`
- `wiki/design/telemetry.md`
- `wiki/DESIGN.md`

## Upstream Impact

- Requires user-approved testnet credential reference, account, symbol, capital, and time window.
- Requires simnet release proof.

## Downstream Impact

- Testnet can be marked implemented only after every acceptance row passes.
- Recovery remains unavailable.

## Cross-Cutting Impact

- Lifecycle: direct complete process.
- State: exact domain and Venue evidence.
- Concurrency: race and leak checks.
- Persistence: generation database integrity.
- Telemetry: feed, heartbeat, Recon, mutation, and fees.
- Shutdown: flat and order-free.
- Security: complete scan.

## Surgical Steps

1. Confirm explicit user authority for one testnet mutation run.
2. Confirm Server is stopped.
3. Confirm account starts flat and order-free.
4. Launch with exact testnet opt-in.
5. Observe fresh Bars and BBO.
6. Permit one bounded strategy execution.
7. Capture acknowledgement and Recon evidence.
8. Request STOP.
9. Prove cancellations, closure, fees, flatness, and claim release.
10. Repeat one controlled disconnect.
11. Validate databases and telemetry.
12. Scan all outputs for secrets.

## Tests and Exact Proof

Automated:

- Full proof ladder below.
- `./stest.sh -bot 9 -runs 1`.
- Exact Bt semantic parity against Target Change #1.

Opt-in testnet:

- `./runner_test.sh -bot <approved_testnet_bot_id> --allow-testnet -duration <approved_duration>`.

Artifacts:

```text
workspace/proof/nuubot-runner/tc15/testnet/run.log
workspace/proof/nuubot-runner/tc15/testnet/result.json
workspace/proof/nuubot-runner/tc15/testnet/telemetry.json
workspace/proof/nuubot-runner/tc15/testnet/domain.json
workspace/proof/nuubot-runner/tc15/testnet/mutations.json
workspace/proof/nuubot-runner/tc15/testnet/reconciliation.json
workspace/proof/nuubot-runner/tc15/testnet/stop-proof.json
workspace/proof/nuubot-runner/tc15/testnet/sqlite-integrity.txt
workspace/proof/nuubot-runner/tc15/testnet/claim-release.txt
workspace/proof/nuubot-runner/tc15/security/secret-scan.txt
workspace/proof/nuubot-runner/tc15/security/mainnet-private-endpoint-scan.txt
workspace/proof/nuubot-runner/tc15/race.txt
workspace/proof/nuubot-runner/tc15/leaks.txt
```

Manual:

- Confirm account flat before and after.
- Confirm testnet UI matches owned Orders and Fills.
- Confirm BalancedRisk warning is visible.

Deferred: recovery and mainnet.

## Non-Goals

- No performance claim.
- No automatic restart.
- No mainnet.
- No unbounded capital.

## Stop and Reassess

- User mutation authority is absent.
- Account is not clean before start.
- Any mutation remains unresolved.
- STOP is not flat and order-free.
- Fee evidence remains incomplete.
- A secret appears.
- Any mainnet private request occurs.

# Target Change #16

## Objective

Gate recovery behind a separate implementation-grade design and plan audit.

## Before

- First release retains crash claims and nonterminal databases.
- Fresh start fails closed.
- Controller, Signaler, BotCycle, and Executor recovery are absent.
- Account and Simulator child reload is insufficient.
- Automatic restart is excluded.

## New

- No implementation occurs under this target without separate user authorization.
- A recovery plan must define exact persisted ownership and restore order.
- Recovery must be explicit, never automatic.
- Recovery must reconcile every Account before policy admission.
- Recovery must preserve the original generation and claims.
- Recovery must distinguish resume, cleanup-only, and unrecoverable evidence.

## Reason

Partial recovery can duplicate mutations or abandon exposure.

## Canonical Owner

- Runner owns explicit recovery orchestration.
- Controller owns policy restoration.
- Signaler owns calculation state.
- BotCycle and Executor own lifecycle restoration.
- Account and Ledger own external truth.

## Exact Affected Files

No file is authorized by this target yet.

A separate approved recovery plan must name every file before work.

Candidate owners requiring assessment:

- `internal/runner`
- `internal/controller`
- `internal/signaler`
- `internal/botcycle`
- `internal/executor`
- `internal/account`
- `internal/ledger`
- `wiki/design/concepts/recovery.md`

## Upstream Impact

- Consumes retained claim, generation, domain, and external Venue evidence.

## Downstream Impact

- May permit explicit recovery only after complete proof.
- Fresh-start crash refusal remains until then.

## Cross-Cutting Impact

- Lifecycle: restore before admission.
- State: every cursor and active identity needs durable ownership.
- Concurrency: no feed policy before restore completes.
- Persistence: schema and transaction boundaries need approval.
- Telemetry: recovery kind remains distinct.
- Shutdown: partial recovery uses normal STOP.
- Security: credentials resolve again by reference.

## Surgical Steps

1. Make no implementation edit.
2. Reassess current persisted state after Target Change #15.
3. Write a separately authorized recovery plan.
4. Run a read-only plan audit.
5. Obtain user approval.
6. Only then authorize named files and implementation.

## Tests and Exact Proof

Deferred until the separate recovery plan:

- Active BotCycle identity restore.
- Signaler boundary restore.
- Executor intent and unknown mutation restore.
- Account reconciliation before admission.
- Cleanup-only recovery.
- Crash during recovery.
- Repeated explicit recovery rejection.
- Exact claim ownership.
- Race, leak, persistence, STOP, and secrecy proof.

Reserved artifact root:

```text
workspace/proof/nuubot-runner/tc16/
```

Manual: explicit operator recovery authority.

## Non-Goals

- No automatic restart.
- No Server respawn.
- No stale claim deletion.
- No fresh Bot substitution.

## Stop and Reassess

- Always stop before implementation until separate authorization and audit exist.
- Any required state lacks a canonical owner.
- External truth cannot be reconciled before policy.
- Recovery would change BotSpec meaning.

## Acceptance Matrix

```text
Acceptance item                         Automated  Opt-in testnet  Manual  Deferred
Focused unit tests                     yes        no              no      no
Focused integration tests              yes        no              no      no
BtBot regression                       yes        no              no      no
BtBot exact semantic result parity     yes        no              no      no
Recorded-event driver parity           yes        no              no      no
Direct simnet Runner                   yes        no              yes     no
Crash and fail-closed restart          yes        no              yes     no
Direct testnet Runner                  no         yes             yes     no
Testnet ordered outcome parity         no         yes             yes     no
Testnet mutation cleanup               no         yes             yes     no
Race proof                             yes        no              no      no
Goroutine leak proof                   yes        no              no      no
Subscription leak proof                yes        no              no      no
SQLite integrity                       yes        no              no      no
Foreign-key integrity                  yes        no              no      no
STOP zero active Orders                yes        yes             yes     no
STOP zero position                     yes        yes             yes     no
STOP closure fee completion            yes        yes             yes     no
Credential and log scan                yes        yes             yes     no
Mainnet private endpoint exclusion     yes        yes             yes     no
Noasm full tests                       yes        no              no      no
Noasm full vet                         yes        no              no      no
Diagnostics                            yes        no              no      no
Diff check                             yes        no              no      no
Windows interrupt                      no         no              yes     no
Linux SIGINT and SIGTERM               no         no              yes     no
Automatic recovery                     no         no              no      yes
Mainnet execution                      no         no              no      excluded
```

## Full Proof Ladder

Run each rung only after the previous rung passes.

### Rung 1: Target-Focused Tests

- Run each target's named package tests.
- Use `-count=1`.
- Use `-tags noasm`.
- Store output below that target's artifact root.

### Rung 2: Shared-Core Bt Proof

- Run Target Change #1 before proof.
- Apply one shared target.
- Run identical after proof.
- Compare semantic fields.
- Reject unexplained differences.

### Rung 3: Deterministic Driver Parity

- Feed the same recorded Bars, BBOs, user hints, heartbeats, and stop event.
- Compare Signal, Risk, cycles, Orders, Fills, finance, exit reason, and STOP evidence.
- Exclude only declared timing and memory fields.

### Rung 4: Simnet Integration

- Run direct Runner with Server stopped.
- Use mainnet public data and Simulator.
- Prove no credential access.
- Prove heartbeat, freshness, domain evidence, and STOP.

### Rung 5: Crash Fence

- Terminate one simnet process after `running`.
- Confirm claim and nonterminal database remain.
- Confirm fresh start refuses ownership.
- Confirm no automatic cleanup occurs.

### Rung 6: Testnet Read-Only

- Require explicit testnet authority.
- Resolve one referenced credential.
- Query Meta, BBO, Bars, open Orders, state, and status.
- Send no mutation.

### Rung 7: Testnet Mutation

- Require explicit mutation authority.
- Place one bounded ordered batch.
- Capture ordered acknowledgement.
- Reconcile exact domain truth.
- Clean every owned mutation.

### Rung 8: Testnet Failure Cases

- Disconnect user stream.
- Prove REST repair.
- Simulate unknown mutation response.
- Resolve by exact identity.
- Prove delayed fee enrichment.
- Keep claims on unresolved cleanup.

### Rung 9: Concurrency and Leak Proof

- Run race tests.
- Repeat Start and Stop.
- Count owned goroutines and subscriptions before and after.
- Prove no callback holds a transport lock.

### Rung 10: Persistence Integrity

- Run SQLite integrity checks.
- Run foreign-key checks.
- Validate lifecycle monotonicity.
- Validate cursor and domain atomicity.
- Validate no `.partial` or orphan file.

### Rung 11: Full Project Proof

Use the project toolchain with CGO disabled:

```text
CGO_ENABLED=0 go test -count=1 -tags noasm ./...
CGO_ENABLED=0 go vet -tags noasm ./...
```

Also run:

- `./stest.sh -bot 9 -runs 1`
- Shell syntax checks.
- Project diagnostics.
- `gofmt -d` on changed Go files.
- `git diff --check`.
- Stale old-path searches.
- Credential and endpoint scans.

### Rung 12: Platform Shutdown

- Windows direct Ctrl+C.
- Linux direct SIGINT.
- Linux direct SIGTERM.
- Each path must produce the same STOP evidence.

## Exact Proof Artifact Root

All implementation proof belongs below:

```text
workspace/proof/nuubot-runner/
```

Each target owns:

```text
workspace/proof/nuubot-runner/tcNN/
```

Every target writes `manifest.txt` containing:

- Git HEAD.
- Dirty-file inventory.
- Exact commands.
- Start and end time.
- Exit codes.
- Referenced log paths.
- Referenced database paths.
- Automated, opt-in, manual, or deferred classification.

Proof artifacts are runtime evidence.

They are not source, wiki, handoff, or audit files.

## Required Documentation Alignment During Implementation

Documentation changes occur only inside the authorized Target Change.

Rewrite owners coherently.

Do not accumulate patch notes.

Required owners include:

- `wiki/PROJECT.md`
- `wiki/ARCHITECTURE.md`
- `wiki/DESIGN.md`
- `wiki/bin.md`
- `wiki/design/concepts/runner.md`
- `wiki/design/concepts/live-events.md`
- `wiki/design/concepts/shutdown.md`
- `wiki/design/concepts/venue.md`
- `wiki/design/concepts/recovery.md`
- `wiki/design/concepts/filesystem.md`
- `wiki/design/packages/setup.md`
- `wiki/design/packages/datastore.md`
- `wiki/design/packages/botspec.md`
- `wiki/design/packages/signaler.md`
- `wiki/design/packages/ohlcv.md`
- `wiki/design/packages/market.md`
- `wiki/design/packages/clock.md`
- `wiki/design/packages/controller.md`
- `wiki/design/packages/botcycle.md`
- `wiki/design/packages/executor.md`
- `wiki/design/packages/account.md`
- `wiki/design/packages/simulator.md`
- `wiki/design/packages/ledger.md`
- `wiki/design/packages/config.md`
- `wiki/design/packages/meta.md`
- `wiki/design/packages/hyperliquid.md`
- `wiki/design/hyperliquid/rest.md`
- `wiki/design/hyperliquid/websocket.md`
- `wiki/design/hyperliquid/exchange.md`
- `wiki/design/telemetry.md`
- `wiki/coding/STYLE.md`
- `wiki/coding/RULES.md`

Remove the fixed-capacity claim.

Replace mode and network conflation.

Remove stale Simulator-only ownership after the hardcut.

Correct stale replay-coupled Signaler language.

State exactly which Hyperliquid operations are implemented.

Keep Server, RunnerControl, shared sockets, pause, resume, automatic restart, and mainnet deferred.

Do not revive superseded `wiki/logic/runner.md`.

## Plan Audit Checklist

- Does every claim match current code or carry an inference label?
- Does the plan treat Runner command and package as new implementation?
- Does the command use BtBot clarity without replay behavior?
- Does BtBot keep its current Parquet and OHLCV behavior?
- Does every shared target require before and after Bt proof?
- Is there one shared Signaler calculation core?
- Is process mode separate from execution and data networks?
- Is mainnet execution impossible?
- Does simnet use mainnet public data without credentials?
- Does testnet use testnet data and referenced credentials?
- Does Account own one minimal Venue?
- Are direct Simulator calls and duplicate gates removed?
- Does the plan forbid compatibility bridges?
- Do callbacks only enqueue?
- Does one lane own every Controller call?
- Is event overflow fail-closed?
- Does startup close the bootstrap-subscription gap?
- Does STOP preserve flat, order-free, and fee-complete evidence?
- Are unknown mutation outcomes never retried automatically?
- Are delayed fees repaired independently from discovery?
- Are feed freshness rules explicit?
- Are Account-symbol claims exclusive and crash-retained?
- Is first-release recovery unavailable and fail-closed?
- Is BalancedRisk described as non-protective?
- Are credentials absent from logs, Results, telemetry, tests, and databases?
- Is simnet proof required before testnet work?
- Is testnet mutation explicitly opt-in?
- Are race and leak proofs mandatory?
- Are Windows and Linux shutdown proofs named?
- Are Server and excluded features still absent?
- Does each target name exact files, impacts, proof, non-goals, and stop gates?
- Does the planning and audit write boundary match user authority?
- Does the auditor remain read-only?
- Does only the planner revise this approved plan file?

## Audit Status

PENDING PLAN AUDIT
