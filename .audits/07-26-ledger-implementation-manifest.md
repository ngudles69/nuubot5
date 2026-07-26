# Ledger Reconciliation Implementation Manifest

Date: 2026-07-26

## Result

Chunk 1 inventory is complete.

- Source, wiki, and `HANDOFF.md` remain untouched by Chunk 1.
- No replay ran.
- Compilation is not a Chunk 1 acceptance condition.
- Git `HEAD` at inventory time: `abee5d5abf47696c4c32c78359600d616ed91732`.
- Later chunks must preserve unrelated work and prove frozen byte identity.

## Canonical Owners

- Source and runtime evidence own implemented truth.
- `wiki/PROJECT.md` owns accepted terminal proof.
- `wiki/baselines/macross-grid-bot.md` owns dated Grid baseline history.
- `rtest.sh` owns exact Observer control counts and semantic gates.
- `trtest.sh` owns Trade relational and semantic gates.
- `grtest.sh` owns Grid relational and semantic gates.
- Canonical docs and retained successful evidence own exact Trade and Grid literals.
- `wiki/PERFORMANCE.md` owns accepted performance history and timer meanings.
- `HANDOFF.md` owns restart state, not accepted baseline literals.

Resolved conflicts:

- Observer signal packages are `2,208` from `rtest.sh`; `wiki/PROJECT.md` still says `2,207`.
- Trade maximum drawdown is `4.200462813402` USDC from `wiki/PROJECT.md`.
- `HANDOFF.md` value `4.21244716452` USDC is not the release gate.
- Any unexplained terminal mismatch fails before performance interpretation.

## Frozen Collision Boundary

Current worktree bytes are frozen with SHA-256.

```text
SHA-256                                                           File
53fa5c131f97f0d12d8e5531426a1e65da3721cf72d8a6b880fd238717e4e550  internal/resultpublisher/resultpublisher.go
a238468f1bf290e41f6b1391509a9edc9448bdf6834eb5e1b9c004cd94e24dab  internal/resultpublisher/resultpublisher_test.go
2db46502579f26e99dd38a97f3c7db6de7679dc7fd943af61253d702b6578172  internal/btrunner/btrunner.go
dc14b22f90cb858537fa8b0838a6763d72676446aad5b9321da084094d17307b  cmd/nuubot-btrunner/main.go
151c7d9e21043bbcd5d4a3103c1f82d1bc82f116b1534a024ac1cda65f236eb5  cmd/nuubot-btrunner/main_test.go
0e62ed635f37971b6f48a9083e91161c1f72ef300bf834ef522c2398d541ddbe  cmd/nuubot-cli/main.go
55a921c5741679bd51e154f95b68635272622c700be06cac585a58f9a5e7c3ea  cmd/nuubot-report/main.go
0e62ed635f37971b6f48a9083e91161c1f72ef300bf834ef522c2398d541ddbe  cmd/nuubot-runner/main.go
0e62ed635f37971b6f48a9083e91161c1f72ef300bf834ef522c2398d541ddbe  cmd/nuubot-server/main.go
3a4e079a45f74042b63c35c685d35439620db62bda3029ed8be7d08987592a73  cmd/parity-probe/main.go
db92631c8bc61f2f8d2d2c5fc3bb6231916dae5aa6fe91e747db1902ba5e19b1  internal/report/render.go
eb337a7a3a9df4180da13c65b0e26963f49f18bb4bd4d776818a4cd4090c0977  internal/report/report.go
6d262a03f3f61de61867d46b4e501961823c24d337b0445509afd8cb56f9b529  internal/report/report_test.go
ABSENT                                                             internal/runreport/render.go
ABSENT                                                             internal/runreport/runreport.go
ABSENT                                                             internal/runreport/runreport_test.go
```

Deleted tracked paths retain these `HEAD` blob identities:

```text
Git blob                                  File
cce6efcec66a1dbceac02c2ec398f413429fb1d9  internal/runreport/render.go
7610d06df964253e0853a92be3d65ef52f59d944  internal/runreport/runreport.go
10797098b9743e6ea37d16cfcc54798163c88c79  internal/runreport/runreport_test.go
```

Later hash proof must use current filesystem bytes.

Git warns tracked frozen files may convert LF to CRLF when touched. Any conversion fails byte identity.

## Exact Current Diffs

Chunk 1 exact command:

```sh
git --no-pager diff -- internal/ledger internal/account internal/resultpublisher internal/btrunner cmd
```

Its tracked patch identity is Git blob `5132a584f5551deda73dd5256e1b95b656445dc9`.

Its exact file and line-count inventory is:

```text
Status  Added  Deleted  File
M           3        3  cmd/nuubot-btrunner/main.go
M           8        8  cmd/nuubot-report/main.go
M           7        7  internal/btrunner/btrunner.go
M         212        0  internal/ledger/ledger_test.go
M          33        2  internal/ledger/store.go
M           4        4  internal/resultpublisher/resultpublisher.go
M           7        7  internal/resultpublisher/resultpublisher_test.go
```

Current Ledger and Account tracked patch identity is Git blob `8627d5d31451d8c3c0dcf9f593c3712f0f1a80fd`.

Current Ledger and Account bytes:

```text
SHA-256                                                           File
6c03eba86ff0839623d8eafe4600508f6e13ea0ba67803f152119db0a25960e9  internal/ledger/ledger_test.go
e788c65e03277ce48df736994dc2fde45ced1fdbe263a2dc69c3da462f96dc66  internal/ledger/store.go
64438e6ce806bca7f0d64a3877b87dbed2ef50a64975271ffab1940e9556c6ad  internal/account/account.go
c55c449f195152d356e1041fa479f8c00b8d781a90d55237d3adeb69b5ed022f  internal/account/account_test.go
```

The unrelated tracked rename patch identity is Git blob `c63137bb9edfc3c72d1d2db2d56761bf13f38d92`.

Unrelated frozen rename state:

```text
Status  Added  Deleted  File
M           3        3  cmd/nuubot-btrunner/main.go
M           8        8  cmd/nuubot-report/main.go
M           7        7  internal/btrunner/btrunner.go
M           4        4  internal/resultpublisher/resultpublisher.go
M           7        7  internal/resultpublisher/resultpublisher_test.go
D           0      267  internal/runreport/render.go
D           0      473  internal/runreport/runreport.go
D           0      181  internal/runreport/runreport_test.go
A           —        —  internal/report/render.go
A           —        —  internal/report/report.go
A           —        —  internal/report/report_test.go
```

The three `internal/report` files are untracked. Git patch output excludes them until added.

## In-Scope Blockers

### Chunk 2

`internal/ledger/ledger_test.go:430` contains:

```go
const characterizedLedgerJSON = "096020a24a61486d24842ed5d166fa84ba11b622da1d08ed1bdb1c7d27322108"TO_CAPTURE"
```

Read-only `gofmt -d` confirms:

```text
internal/ledger/ledger_test.go:430:99: expected ';', found TO_CAPTURE
internal/ledger/ledger_test.go:430:109: string literal not terminated
```

Chunk 2 owns placeholder removal, reviewable characterization fixtures, and restored Ledger test compilation.

### Chunk 5

`internal/ledger/store.go:69-72` contains malformed partial hook formatting:

```go
if err = s.beforeStatement(); err != nil {
return err
}
_, err = transaction.Exec(`
```

Read-only `gofmt -d` requires indentation only for this block.

Chunk 5 owns this formatting repair and real SQL fault-boundary proof.

## Unrelated Frozen Failures

- No compile failure is attributed to frozen rename code.
- Compile-state probing stopped before package analysis because the sandbox denied the external Go build cache.
- A second external cache path under `D:\tmp` was also denied.
- This environmental failure is separate from Chunk 2 and Chunk 5 source defects.
- Frozen rename compilation remains unverified, not green.
- Any later rename compile failure is an external blocker and cannot be repaired under Ledger scope.

Observed external error:

```text
open C:\Users\PC\AppData\Local\go-build\...\...-a: Access is denied.
failed to initialize build cache at D:\tmp\nuubot5-go-cache: mkdir D:\tmp\nuubot5-go-cache: Access is denied.
```

Chunk 1 has no compile gate.

## Terminal Baseline Manifest

### Observer — Sweep 6 Bot 9

Owners: `rtest.sh`, `wiki/PROJECT.md`, and successful result database.

- Ticks: `7,948,800`.
- Controller runs: `794,880`.
- Signal packages: `2,208`.
- Skipped starts: `724`.
- Cycles: `64` started, zero rejected, `64` closed.
- Stop-loss exits: `17`.
- Telemetry samples: `794,881`.
- Stop reason: `parent_stop`.
- Semantic completion: process zero, nonempty report JSON, and exact controller stop line.
- Account cursor: not applicable.

### Trade — Sweep 9 Bot 13

Literal owners: `wiki/PROJECT.md` and retained successful result evidence.

`trtest.sh` owns relational and semantic gates only.

- Ticks: `7,948,800`.
- Controller runs: `794,880`.
- Cycles: `193`.
- Trades: `193`.
- Orders: `626`.
- Fills: `386`.
- Stop Orders: `47`.
- Retries: `0`.
- Capital: `1,000` USDC.
- Net PnL: `-3.90459332761` USDC.
- Ending equity: `996.09540667239` USDC.
- Maximum drawdown: `4.200462813402` USDC.
- Telemetry samples: `794,881`.
- Every terminal cursor equals that Account's final successful reconciliation timestamp.
- The last cycle cursor equals `backtest_result.last_ms`.
- Semantic completion includes completed flag one and exact Config match.
- Semantic completion includes cycle equity carry and terminal equity match.
- Semantic completion includes zero false-equity samples and nondecreasing maximum drawdown.
- Semantic completion includes zero legacy `close` Orders and no `.partial`.
- Integrity must be `ok`; foreign-key rows must be zero.

### Grid — Sweep 10 Bot 14

Literal owners: current `wiki/baselines/macross-grid-bot.md` and retained successful result evidence.

`grtest.sh` owns relational and semantic gates only.

- Ticks: `7,948,800`.
- Controller runs: `794,880`.
- Signal packages: `2,208`.
- Cycles: `50`.
- Trades: `1,982`.
- Orders: `4,697`.
- Fills: `2,636`.
- Cancellations: `2,061`.
- Stop Orders: `733`.
- Retries: `0`.
- Round trips: `585`.
- Net PnL: `-57.420074089999999993851` USDC.
- Ending equity: `942.579925910000000006149` USDC.
- Maximum drawdown: `75.791979199999999992245` USDC.
- Telemetry samples: `794,881`.
- Every terminal cursor equals that Account's final successful reconciliation timestamp.
- Each terminal cycle cursor equals its terminal Account observation.
- Semantic completion includes `1,500` Grid levels and `100` boundaries.
- Semantic completion includes `1,400` initialized active levels.
- Semantic completion includes zero active Orders and zero nonflat Accounts.
- Semantic completion includes exact Config match, zero legacy `close`, and no `.partial`.
- Integrity must be `ok`; foreign-key rows must be zero.

## Chunks 2–19 Exact Write Sets

### Chunk 2 — Freeze Reviewable Ledger Oracles

- `internal/ledger/ledger_test.go`
- `internal/ledger/testdata/characterization/*.json`

### Chunk 3 — Freeze Account and Publication Oracles

- `internal/account/account_test.go`
- `internal/account/testdata/characterization/*.json`
- `internal/resultpublisher/ledger_characterization_test.go`
- `internal/resultpublisher/testdata/ledger_characterization/*.json`

### Chunk 4 — Add Green Failure and First-Error Characterization

- `internal/ledger/ledger_test.go`
- `internal/account/account_test.go`
- `internal/executor/trade_test.go`
- `internal/executor/grid_test.go`
- `internal/botcycle/botcycle_test.go`
- `internal/controller/controller_test.go`
- `internal/btrunner/btrunner_test.go`

### Chunk 5 — Prove Real SQL Fault Boundaries

- `internal/ledger/store.go`
- `internal/ledger/ledger_test.go`

### Chunk 6 — Define Opaque Candidates and Generation Invalidation

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account.go`
- `internal/account/account_test.go`

### Chunk 7 — Share Exact Trade and Order Calculations

- `internal/trade/trade.go`
- `internal/trade/trade_test.go`
- `internal/order/order.go`
- `internal/order/order_test.go`

### Chunk 8 — Stage Observational Indexes and Collision Proof

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Chunk 9 — Build Exact Deltas Off-Path

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/order/order.go`
- `internal/order/order_test.go`
- `internal/fill/fill.go`
- `internal/fill/fill_test.go`

### Chunk 10 — Isolate Complete Store Ownership

- `internal/ledger/store.go`
- `internal/ledger/publish.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/resultpublisher/ledger_characterization_test.go`

### Chunk 11 — Cut CreateTrade Incremental Store

- `internal/ledger/store.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Chunk 12 — Cut AddOrders Incremental Store

- `internal/ledger/store.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Chunk 13 — Cut RecordSubmit Incremental Store

- `internal/ledger/store.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Chunk 14 — Prove Recon Dirty Store Off-Path

- `internal/ledger/store.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account_test.go`

### Chunk 15 — Prepare Off-Path Candidate Commit and Memory Publication

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account_test.go`

### Chunk 16 — Atomic Recon Routing Cutover

- `internal/account/account.go`
- `internal/account/account_test.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Chunk 17 — Switch Ledger Identity Reads

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Chunk 18 — Switch Account Missing-Status Reads

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account.go`
- `internal/account/account_test.go`

### Chunk 19 — Add Exact Scaling Benchmarks

- `internal/ledger/recon_benchmark_test.go`
- `internal/ledger/ledger_test.go`

## Chunk 1 Proof

- Sole write: `.audits/07-26-ledger-implementation-manifest.md`.
- Exact Chunk 1 diff command ran.
- Frozen current bytes are recorded.
- Current collision diffs are recorded.
- In-scope and unrelated blockers are separate.
- No source, wiki, `HANDOFF.md`, replay, commit, or push action occurred.
