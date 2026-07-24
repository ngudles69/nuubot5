# Handoff

Last updated: 2026-07-24 20:04:39 +08:00

## Focus

Simulator-first trading transaction flow is implemented and proven.

## Published baseline

- Commit `465f829` is pushed on `main`.
- Current tranche is uncommitted.

## Current state

- Shared database is `workspace/db/nuubot.db`.
- Shared database contains nine Sweeps, thirteen Bots, and mainnet Meta.
- Meta contains 232 perpetual symbols.
- Setup refreshes mainnet Meta after 24 hours.
- Testnet never supplies Meta.
- Sweep 9 and Bot 13 own the TradeExecutor proof.
- Per-Bot results use `workspace/db/sweeps/sweep_9/bot_13.db`.
- Setup, BtRunner, Runtime, and BotCycle retain their user-vetted structures.

## Implemented

- CLOID uses the Nuubot4 128-bit layout.
- Account owns Simulator, Ledger, CLOID creation, normalization, and recon state.
- Ledger owns Trade, Order, Fill, and account snapshots.
- Simulator owns bracket matching, OCO, fees, PnL, and durable child state.
- TradeExecutor owns one Account and one bracket Trade.
- ResultPublisher writes completed memory-mode results through a partial database.
- `none` persists only after successful completion.
- `max` persists every accepted state change.
- `max` reloads Ledger and Simulator child state only.
- Full Bot resume remains pending Runner-owned orchestration cursors.
- TradeExecutor fatally rejects persisted Trades until Runner recovery exists.
- Simulator publishes memory only after successful maximum persistence.
- Account repairs absent created or submitted Simulator Orders during recon.
- BtRunner publishes completion only after ReplayReader and Runtime stop successfully.
- Runtime performs recon before Executor decisions.

## Proof

- Full Go tests pass with `CGO_ENABLED=0` and `-tags noasm`.
- Go vet, module tidy, formatting, diff, and wiki-link checks pass.
- TradeExecutor stability passed 13/13 attempted runs.
- Replay processed 7,948,800 ticks and 794,880 Runtime passes.
- Result contains 50 closed Trades, 151 Orders, and 100 Fills.
- Result contains 50 Simulator states at schema version 2.
- Shared and result database integrity checks pass.
- Result foreign-key checks pass.
- No partial result database remains.
- Final 1x replay took 2,696 ms; process took 4,526 ms.
- Final 2x averaged 2,481 ms replay and 3,011 ms process.
- Final 10x averaged 2,505 ms replay and 3,021 ms process.
- 10x suite took 34,052 ms.
- Audit round one found material durability and recovery-boundary issues.
- Accepted findings were fixed with focused failure and recovery tests.
- Audit round two passed with no material finding or bloat.
- 1x log: `workspace/logs/nuubot5-trtest-s9-b13-1-20260724T120251Z.log`.
- 2x log: `workspace/logs/nuubot5-trtest-s9-b13-2-20260724T120340Z.log`.
- 10x log: `workspace/logs/nuubot5-trtest-s9-b13-10-20260724T120354Z.log`.

## Active work

- None.

## Pending user review

- Account
- Ledger
- Trade
- Order
- Fill
- Simulator
- TradeExecutor
- Meta
- ResultPublisher

## Pending user approval

- Commit and push the completed tranche.

## Blockers

- None.

## Next action

Wait for user authority to commit and push.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
