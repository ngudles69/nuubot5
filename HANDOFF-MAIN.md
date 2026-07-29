# Nuubot5 Main Handoff

Last updated: 2026-07-29

## Focus

Finish the flat Account Store integration.

The Venue and Simulator protocol expansion is complete.

The Account, Ledger, Trade, Order, Fill, Recon, Venue, and Simulator production
stack is connected well enough for Observer, Trade, and Grid backtests.

Persistence is partially restored. Account writes work. Account reload remains
unfinished.

Do not commit or push without explicit user authority.

Ignore `workspace/backups/**`.

Shared `AGENTS.md` and `HANDOFF.md` now route by exact worktree.

Main state lives in `AGENTS-MAIN.md` and `HANDOFF-MAIN.md`.

Server targets were not copied. They will arrive through the server-worktree
sync.

## DONE

- Hardcut Fill, Order, Trade, Ledger, and Account to flat production records.
- Remove Input, Record, State, clone, compatibility, and Batch-number models.
- Delete the old tests under `internal/account/**` for later replacement.
- Join the flat Account stack internally.
- Complete and close the Account-stack adversarial audit.
- Restore Recon around the flat Ledger.
- Add current-Recon OID repair for downloaded Fills.
- Add Executor-owned Order Level.
- Route Account through one Venue interface.
- Route simnet Venue calls to Simulator without direct Account access.
- Reconnect Observer, Trade, and Grid.
- Add flat Account and Simulator SQL tables.
- Add dirty Trade, Order, Fill, Simulator Order, and Simulator Fill persistence.
- Preserve domain tables during ResultPublisher publication.
- Pass current Observer, Trade, and Grid backtests.
- Save Grid Baseline 1 and Baseline 2.
- Define the canonical result comparison format.
- Split shared routers from main-worktree instructions and handoff state.
- Commit the complete Account-stack hardcut and flat Store work as `3e98a69`.
- Push `3e98a69` to `origin/main`.
- Rename Simulator sections to Venue Interface and Lifecycle, Matching Engine,
  and Helpers without changing behavior.
- Rename Simulator lifecycle to Connect and Disconnect.
- Rename Venue lifecycle to Connect and Disconnect.
- Hardcut Venue reads to Get Open Orders, Get Order History, Get Fill History,
  Get Order Status, and Get Account State.
- Add Hyperliquid-shaped Order History reconciliation.
- Add persisted Simulator leverage and margin mode through Set Leverage.
- Split Simulator submission facts from mutable Venue outcome facts.
- Organize Simulator domain functionality into Matching Engine and Persistence.
- Remove the incorrect one-Fill-per-batch-per-BBO restriction.
- Save the Nuubot3 and NautilusTrader intent comparison.
- Pass Observer, Trade, and Grid after the Venue lifecycle hardcut.
- Ignore local `workspace/backups/**` artifacts.

## Domain Model

Ledger owns flat maps and relationship indexes.

Order owns no Fills. Trade owns no Orders.

Every Trade, Order, and Fill has its own memory identity.

Every Fill keeps Exchange-supplied CLOID, OID, and Venue TID unchanged.

CLOID is the Order Venue identity and encodes `(LedgerID, OrderID)`.

Venue TID deduplicates Fills.

OID resolves Fill ownership only when Exchange omitted CLOID.

No Batch entity, Batch number, Batch key, or Batch lifecycle exists.

Every Order stores Executor-owned `Level`.

Grid Orders use their Grid level. Other Executors use Level zero.

Canonical designs:

- `wiki/design/account.md`
- `wiki/design/trade.md`
- `wiki/design/concepts/trading-schema.md`
- `wiki/design/concepts/cloid.md`

## Trade Closure

One Trade may contain any number of entry, take-profit, stop-loss, and close
Orders.

DCA entries and staged or partial exits are valid.

Entry and exit counts need no symmetry.

A Trade closes only when all Orders are closed and signed Fill size is zero.

Zero size does not close a Trade while any Order remains open.

An Order with a fee-incomplete Fill remains open.

Exchange Order evidence owns Venue status.

Account and Ledger record Exchange facts. Executors own trading intent.

## Recon

Canonical owner: `wiki/design/concepts/recon.md`.

Current flow:

1. Prepare attempt.
2. Download current Order evidence.
3. Download Fill history.
4. Download current Account state.
5. Update Fill records.
6. Update Order records.
7. Search deferred OID-only Fills against newly indexed Orders.
8. Update touched Trade records.
9. Update Account Snapshot.
10. Persist and publish.
11. Finalize outcome.

Fill updates precede Order updates.

The OID search reuses the same Fill and Order update functions.

Only matched Fills, distinct owning Orders, and touched Trades reapply.

Every executed search logs `Recon-OIDSearch found nothing` or exact found
counts.

Missing Open Order evidence invents nothing.

Exact status queries resolve missing active Orders.

Fill cursor is inclusive.

Missing fees use bounded repair windows.

No Ledger clone or rollback exists.

## Venue and Simulator

Canonical owners:

- `wiki/design/concepts/venue.md`
- `wiki/design/packages/venue.md`
- `wiki/design/packages/simulator.md`

Account calls only the Venue interface.

Venue owns Simulator connection, calls, and disconnection.

Account never imports, constructs, or calls Simulator.

Simulator receives config values, not Account or Store objects.

Simulator subscribes to MarketData, matches Orders, and updates Simulator-owned
Exchange state.

Simulator sends no direct callbacks to Account.

Account currently discovers Simulator changes through clean Recon sweeps.

Future Venue WebSocket support may carry protocol-shaped `userEvent` updates.

Mainnet, testnet, signing, authentication, and live WebSocket work remain
future scope.

Nuubot5 source owns current Venue behavior.

Nuubot3, Nuutrader6, and NautilusTrader supply reusable intent only.

Venue routes calls and contains no Account or trading business logic.

Active Asset Data intent belongs to Meta.

Future Cancel All Orders belongs in Account as convenience orchestration.

## Store

Current table names:

```text
account
ledger
trade
order
fill
simulator
simulator_order
simulator_fill
```

`order` is quoted in SQL.

Columns map directly to current production structs.

Account and Simulator use separate SQLite connections to the same runtime
database.

Each connection uses one open connection, immediate transactions, foreign
keys, and a 30-second busy timeout.

Clean domain rows never reach SQLite.

Dirty flags clear only after successful commit.

Failed commits retain dirty rows and fail publication.

`none` exports all rows once during terminal shutdown.

`max` writes changed rows during execution.

Backtests recreate the partial result database every run.

ResultPublisher appends report tables, then atomically publishes the database.

## Store Proof

Current Venue expansion proof:

```text
go build -tags noasm ./internal/account ./internal/venue ./internal/simulator ./internal/executor
stale Venue lifecycle name scan
git diff --check
```

All passed. Latest behavioral proof follows.

Latest backtest proof:

```text
Observer  PASS  workspace/logs/nuubot5-stest-s6-b9-1-20260729T054523Z.json
Trade     PASS  workspace/logs/nuubot5-stest-s9-b13-1-20260729T054539Z.json
Grid      PASS  workspace/logs/nuubot5-stest-s11-b15-1-20260729T054607Z.json
```

Trade and Grid execution counts and financial results exactly match Baseline 2.

Grid Baseline 2 database integrity: `ok`.

Foreign-key check: clean.

```text
Table              Rows    Coverage
account               50   Cycles 1-50
ledger                50   50 Ledgers
trade              1,980   50 Ledgers
order              4,693   50 Ledgers
fill               2,632   50 Ledgers
simulator               1   Latest root only
simulator_order       201   Mixed cycle rows
simulator_fill        158   Mixed cycle rows
```

Account-side counts exactly match the Grid report.

## Known Simulator Persistence Defect

Do not fix this as a standalone Store patch.

Simulator identity currently uses Account and Symbol only.

Every cycle reinitializes Simulator and restarts Venue Order and Fill IDs.

Later cycles overwrite matching IDs and leave stale higher IDs.

Observed root next OID was 72 while stored rows reached 201.

Observed root next TID was 31 while stored rows reached 158.

The planned Simulator architectural change will replace this identity model.

Until then, Simulator persistence rows are not trusted across cycles.

## Baselines

Canonical comparison format:
`wiki/design/concepts/comparison.md`.

Grid baselines:
`wiki/baselines/macross-grid-bot.md`.

Baseline 2 name: `Post ALTOFRVS Change`.

ALTOFRVS means Account, Ledger, Trade, Order, Fill, Recon, Venue, and Simulator.

Future result displays show Baseline 1, Baseline 2, Current Run, and Diff versus
Baseline 2.

Completed round trips are not a baseline comparison metric.

### Observer Baseline 2

```text
Suite                    8,992 ms
BtBot                    8,536 ms
Replay loop              5,629 ms
Ticks                    7,948,800
Cycles                   63
```

Report:
`workspace/logs/nuubot5-stest-s6-b9-1-20260729T041943Z.json`.

### Trade Baseline 2

```text
Suite                    21,149 ms
BtBot                    19,258 ms
Replay loop              14,937 ms
Trades                   192
Orders                   623
Fills                    384
Net PnL                  -3.87659936049
Ending equity            996.12340063951
Maximum drawdown         4.21802556724
```

Report:
`workspace/logs/nuubot5-stest-s9-b13-1-20260729T041350Z.json`.

### Grid Baseline 2

```text
Suite                    38,828 ms
BtBot                    36,028 ms
Replay loop              32,089 ms
Trades                   1,980
Orders                   4,693
Fills                    2,632
Net PnL                  -57.440213499999999998272
Ending equity            942.559786500000000001728
Maximum drawdown         75.537686959999999996921
```

Report:
`workspace/logs/nuubot5-stest-s11-b15-1-20260729T042021Z.json`.

## Audit Evidence

- `.audits/07-29-account-stack-coherency.md`
- `.audits/07-29-account-stack-coherency-v2.md`
- `.audits/07-29-account-stack-coherency-v3.md`
- `.audits/07-29-account-venue-coherency-v1.md`

The Account-stack adversarial audit is closed.

No further hypothetical-error audit is requested.

## TODO

1. Add durable Account and Ledger reload for `max`.
2. Rebuild Ledger indexes, active sets, counters, and Summary during reload.
3. Validate Account Store identity and schema before restore.
4. Add Account-row dirty tracking.
5. Persist final Recon statistics without one-attempt lag.
6. Keep failed Store publication from appearing successful.
7. Update `stest.sh` maximum-mode queries to the eight current table names.
8. Measure clean transaction cost and a 30-row dirty transaction.
9. Align remaining Store, Simulator, and ResultPublisher documentation.
10. Repeat Observer, Trade, and Grid proof after Store completion.

## Deferred

- Simulator persistence identity fix through the planned architecture change.
- Simulator `userEvent` through the future Venue event path.
- Mainnet and testnet transport, signing, and authentication.
- Live WebSocket reconciliation.
- Persistence features not required by current backtest proof.
- Live-support features after backtest logic and persistence stabilize.
- Historical equity and balance retention tiers.
- Cross-process Account claims.
- Multi-source replay merge.
- Global physical-Account risk.
- Runner periodic telemetry persistence.
- Server monitoring and recovery.

## PENDING USER APPROVAL

- None.

## Next Action

Continue Account Store reload and dirty-row correctness.

Do not fix Simulator cross-cycle identity until its architectural change.
