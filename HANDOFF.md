# Handoff

Last updated: 2026-07-24 18:12:25 +08:00

## Focus

Complete the Hyperliquid clearinghouse slice and implementation-ready trading-state design.

## Current status

- Published baseline remains commit `2b230ef`.
- Current worktree contains the trading-state assessment and Hyperliquid clearinghouse slice.
- `internal/hyperliquid` implements bounded testnet clearinghouse-state reads.
- Private wire fields retain official Hyperliquid names.
- Exported values use readable names and exact `decimal.Decimal` values.
- `crossMaintenanceMarginUsed` maps to `MaintenanceMargin`.
- `marginSummary.totalMarginUsed` maps to `Margin.MarginUsed`.
- `cmd/parity-probe` and `internal/parity` are implemented.
- `internal/parity/info` owns `/info` probes.
- Implemented command shape:
  `parity-probe <testnet|simnet> <account> <operation> [arguments...]`.
- Implemented operation: testnet `clearinghouse-state`.
- Simulator remains under `internal/simulator`.
- Simnet probing waits for real Simulator clearinghouse behavior.
- No fake Simulator adapter exists.
- `persist_mode` is `none` or `max`.
- Account passes one configured persistence mode to Ledger and Simulator.
- `none` keeps both in memory until successful result publication.
- `max` persists every accepted Ledger mutation and Simulator state change.
- `none` opens no result database during Account execution.
- ResultPublisher atomically publishes the final database only after success.
- Hyperliquid history omission never deletes local Ledger evidence.
- WebSocket `userEvents` mark Account dirty only.
- Account solely owns reconciliation-dirty state.
- Normal recon skips clean Accounts. Forced recon still queries every Account.
- Only successful recon commits canonical lifecycle, position, and balance truth.
- Permanent design:
  [Hyperliquid Parity Probe](wiki/design/hyperliquid/parity.md).
- Chief-of-staff review and tranche state remains in [DESIGN](wiki/DESIGN.md).

## Active agents

- Root only.

## Blockers

- None.

## Files changed

- `.gitignore`: admits source `internal/config/credentials.go`.
- `internal/config/credentials.go`: existing credentials loader is now trackable.
- `internal/hyperliquid/**`: REST transport, raw payload, decoding, decimals, and tests.
- `internal/parity/**`: permanent harness, `/info` probe, evidence writer, and tests.
- `cmd/parity-probe/main.go`: thin executable boundary.
- `workspace/config/config.toml`: request timeout raised from 2 to 10 seconds.
- `wiki/design/hyperliquid/**`: REST, parity, and captured JSON evidence.
- Trading-state package and concept pages remain modified from the active assessment.
- `wiki/DESIGN.md` catalogs all 24 internal package pages.

## Proof

- Full tests passed for all 29 Go packages with `CGO_ENABLED=0` and `-tags noasm`.
- Full `go vet -tags noasm ./...` passed.
- `go mod tidy` is clean.
- All 10 changed Go files pass `gofmt`.
- All 36 changed Markdown files pass local-link checks.
- `git diff --check` passed with line-ending warnings only.
- Final adversarial re-audit passed with no material finding.
- Malformed HTTP payload proof preserves raw bytes before decoding failure.
- DDL applied twice and rejected cross-Ledger Order, Fill, and CLOID ancestry.
- Approved testnet accounts `tgrid` and `thedge` returned HTTP 200.
- `tgrid`: equity `172.232247`, positions `0`, duration `165 ms`.
- `thedge`: equity `549.237687`, positions `0`, duration `150 ms`.
- Go and `async_hyperliquid` field paths match for both accounts.
- Values match exactly after excluding request-time field `time`.
- Evidence:
  `wiki/design/hyperliquid/json/info/clearinghouse-state/20260724-clearinghouse-baseline/testnet`.

## Proof pending

- Simulator parity, because Simulator is not implemented.
- Non-empty testnet position capture.
- Exchange mutations, signing, WebSocket, Account, Ledger, and recon implementation.

## Decisions

- `cmd/parity-probe/main.go` owns only the process boundary.
- `internal/parity` owns permanent probe orchestration.
- Endpoint probes are grouped by real Hyperliquid endpoint.
- Meta and clearinghouse operations belong under `internal/parity/info`.
- Future exchange operations belong under `internal/parity/exchange`.
- HTTP and WebSocket operations retain protocol vocabulary.
- Empty lifecycle phases remain omitted.
- Raw payloads and normalized state remain separate evidence.
- Any returned raw payload is written before decoding.
- HTTP mutation responses remain acknowledgement evidence until recon.
- Exchange history is bounded. Missing rows never prove local evidence invalid.
- Account filters Fill history from an inclusive cursor and bounded time range.
- A capped or unproven history range blocks cursor advancement.
- ResultPublisher writes `none` evidence only after successful completion.
- Terminal evidence flows by owned value from Ledger and Simulator to ResultPublisher.
- TradeExecutor captures results after shutdown recon and before Account teardown.
- BotCycle collects cached results only after Executors stop.
- Result values alias no mutable child state.
- `parity-probe` has the sole operator exception to write tracked wiki evidence.
- Credentials, API keys, and signing material never enter evidence.

## Pending user approval

- Commit and push the completed tranche.

## Next action

Await user authority to commit and push or begin the trading implementation tranche.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
