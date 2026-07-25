# Handoff

Last updated: 2026-07-25 04:30:55 +08:00

## Focus

Finish and prove the approved three-month Macross TradeBot architecture.

Then inspect Nuutrader6 and prepare the requested Grid Executor plan.

## Published Baseline

- Branch: `main`.
- `HEAD` and `origin/main`: `da1566ea8e1a9041316485eb3bb21aa6109997b1`.
- Message: `docs: define bot controller target architecture`.
- Documentation adversary passed round two without material findings.

## Active Implementation

- Exact database BotSpecID, TOML bytes, and SHA-256 are authoritative.
- TOML remains the only persisted BotConfig representation.
- AppConfig no longer owns Bot behavior.
- BotSpec admits exact typed Config and builds immutable Bot definitions.
- Implemented BotSpecs are `macross_observer_bot` and `macross_trade_bot`.
- Controller replaced Runtime without compatibility code.
- Controller owns persistent Signaler, Risks, and zero or one active BotCycle.
- `max_cycles = 999` permits sequential cycles.
- Maximum concurrent BotCycles remains structurally fixed at one.
- Signal uses `NoAction`, `StartCycle`, and `StopCycle`.
- Risk uses immutable RiskInput and typed decisions.
- Executors own fixed side, capital, sizing, and distinct resources.
- Controller carries terminal Account equity into the next cycle.
- Every Risk assesses the same immutable input before Controller acts.
- Stronger Risk decisions dominate weaker decisions.
- Resource equity uses the exact Venue, network, Account, and symbol tuple.
- ResultPublisher writes Controller, cycle, Executor, Signal, Risk, and trading evidence.
- Results preserve exact admitted BotConfig TOML and hash.
- Final publication preserves both none-mode and maximum-mode Account children.
- Simulator replay loads no private credentials.

## Local Database

- Database: `workspace/db/nuubot.db`.
- Recoverable pre-hardcut backup: `workspace/db/nuubot.pre-botspec-20260725.db`.
- Thirteen Bot rows contain exact TOML and valid hashes.
- Sweep 6 Bot 9 uses `macross_observer_bot`.
- Sweep 9 Bot 13 uses `macross_trade_bot`.
- Stored templates now use `max_cycles = 999`.
- Latest integrity check: `ok`.

## Verified Proof

- Full Go tests and vet pass with `CGO_ENABLED=0` and `-tags noasm`.
- TradeBot processed 7,948,800 ticks and 794,880 control passes.
- TradeBot completed 193 cycles, 193 Trades, 626 Orders, and 386 Fills.
- TradeBot capital was 1,000 USDC.
- TradeBot net PnL was -3.90459332761 USDC.
- TradeBot ending equity was 996.09540667239 USDC.
- TradeBot maximum drawdown was 4.200462813402 USDC.
- Cycle 2 starting equity equals Cycle 1 terminal equity.
- Result BotConfig equals the exact stored database TOML and hash.
- Result integrity and foreign-key checks passed.
- TradeBot passed 2 of 2 and 10 of 10 fresh processes.
- Observer passed 2 of 2 and 10 of 10 fresh processes.
- TradeBot 10x log: `workspace/logs/nuubot5-trtest-s9-b13-10-20260724T202751Z.log`.
- Observer 10x log: `workspace/logs/nuubot5-rtest-s6-b9-10-20260724T202915Z.log`.
- Post-audit TradeBot 1x passed in 9,393 ms process and 4,248 ms replay.
- Post-audit TradeBot log: `workspace/logs/nuubot5-trtest-s9-b13-1-20260724T204016Z.log`.
- Post-audit Observer 1x passed in 1,788 ms process and 1,694 ms replay.
- Post-audit Observer log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T204033Z.log`.
- Implementation audit round one found maximum-mode final publication data loss.
- The accepted finding received one focused failing test and owning-path fix.
- Implementation audit round two passed with no material finding or bloat.

## Completed

- Approved TradeBot architecture implementation.
- Full three-month TradeBot proof.
- TradeBot and Observer 2x and 10x stability.
- Final adversarial implementation audit.
- Post-audit static and replay proof.

## TODO

- Inspect Nuutrader6 Grid behavior read-only.
- Present the exact Nuubot5 Grid Executor design for user approval.

## Pending User Approval

- Grid Executor implementation.
- Final source commit and push.

## Grid Executor Request

- Inspect `D:\rust\nuutrader6` after TradeBot closeout.
- Proposed range uses starting price minus 5 percent through plus 5 percent.
- Proposed level count is 32.
- Proposed deployed capital is 95 percent of assigned Executor capital.
- Every level must satisfy Venue minimum order value, such as 11 USDC.
- Every Grid Executor start logs its validated level table before order placement.
- Each row includes level, price, side, size, notional, and intended action.
- Present canonical ownership, sizing semantics, files, outcome, and proof before editing.

## Deferred

- Live cross-process Account claims.
- Multi-source replay merge.
- Physical Account and global risk.
- Periodic telemetry.
- Server monitoring and recovery.

## Next Action

Inspect Nuutrader6 Grid behavior and prepare the approval plan.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
