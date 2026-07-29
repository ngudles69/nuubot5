# Account

Status: Canonical execution and accounting boundary.
Purpose: Execute Executor Order intent and preserve exact Exchange facts.

## Ownership

Executor owns trading business logic.

Executor decides:

- Order role;
- side;
- quantity;
- price;
- type;
- time in force; and
- reduce-only behavior.

Account receives complete Order intent from Executor.

Account executes that intent without changing its business meaning.

Account owns Venue encoding, submission, response handling, and
reconciliation.

Ledger owns flat Trade, Order, and Fill records, relationship indexes,
accounting totals, exposure, fees, and PnL.

Account and Ledger do not:

- correct Executor quantities;
- force reduce-only from Order role;
- infer strategy intent;
- repair invalid Trade design; or
- hide Executor defects.

If Executor creates invalid business intent, the defect belongs to Executor.

Account may reject structurally invalid or Venue-invalid execution requests.

It must not replace an Executor decision with another decision.

## Execution Boundary

```text
Account
└── Venue(config)
    ├── mainnet/testnet
    │   └── Exchange
    └── simnet/backtest
        └── simulated Venue/Exchange
```

Account owns one concrete Venue lifecycle.

Account calls only Venue methods.

Account never imports, constructs, calls, or reads Simulator.

Venue selects behavior from its configuration.

Current Venue supports simnet only.

Venue owns Simulator initialization, protocol calls, and shutdown for simnet.

Future mainnet and testnet Venue behavior owns Exchange transport,
authentication, signing, and protocol events.

Simulator reproduces the complete Venue-visible Exchange unit.

Simulator returns the same protocol responses as the future live path.

Account uses identical decoding and reconciliation paths for every network.

Simulator receives no Account or Ledger reference.

Simulator never calls Account.

Simulator subscribes to MarketData and updates only Simulator-owned Exchange
truth.

Account observes that truth through later Venue reconciliation calls.

Account solely owns reconciliation dirty state.

Simulator MarketData changes do not mark Account dirty.

Dirty or pending Accounts use the short reconciliation cadence.

Clean Accounts use the 60-second sweep.

Future Simulator events must use the approved Venue protocol event path.

They must never become direct Account callbacks.

## Exchange Authority

Account and Ledger preserve exact Exchange identity, status, execution, fee,
timestamp, and payload evidence.

Submission or cancellation intent does not invent Exchange state.

Reconciliation replaces mutable Exchange snapshot fields with current Exchange
values.

Ledger derives accounting results from accepted Exchange facts.

Derived accounting never rewrites Executor intent or Exchange evidence.
