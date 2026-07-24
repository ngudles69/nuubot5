# Simulator Parity

Status: Proposed assessment.
Covers: No implemented source.
Purpose: Separate Hyperliquid behavior, SDK-visible responses, and Nuubot domain evidence.

## Verdict

Nuubot3 needs correction, not replacement.

Its Account, Ledger, Trade, Order, Fill, and reconciliation design remains the strongest domain reference.

Nuutrader6 supplies the proven exchange behavior.

`async_hyperliquid` supplies the client-visible response contract used by Nuutrader6.

Nuubot5 rewrites both contracts in Go.

## Three Contracts

| Contract | Source | Nuubot5 owner |
|---|---|---|
| Exchange behavior | Official API and proven Nuutrader6 behavior | Simulator and Hyperliquid adapter |
| Client-visible response | `async_hyperliquid` 0.4.8 and frozen fixtures | `internal/hyperliquid` protocol types |
| Domain evidence | Nuubot3 Account and Ledger design | Account, Ledger, Trade, Order, Fill |

These contracts may resemble each other.

They are not interchangeable.

Account must not read Simulator internals.

Simulator must not create domain objects.

## Batch Bracket Response

Order results retain request order:

```text
0 entry
1 take profit
2 stop loss
```

A resting entry returns:

```text
entry        resting
take profit  waitingForFill
stop loss    waitingForFill
```

An immediately filled entry returns:

```text
entry        filled
take profit  waitingForTrigger
stop loss    waitingForTrigger
```

`waitingForFill` means the child waits for its parent entry.

`waitingForTrigger` means the child is armed against market data.

Public Order queries may expose both child states as `open`.

Account retains the exact submit result and later reconciles public Order truth.

Neither Simulator nor live submission responses bypass reconciliation.

WebSocket events remain optional dirty hints.

Account passes `persist_mode` into Simulator.

`none` remains memory-only until one successful final export.

`max` persists every Simulator state change.

Recovery reloads Simulator truth and forces recon before any decision.

Forced reconciliation catches missing hints and lifecycle drift.

## Bracket Lifecycle

```text
submit entry, TP, SL
  entry rests
    TP and SL wait for parent Fill
  entry fills
    TP and SL become trigger-active
  TP or SL triggers
    selected child fills
    sibling cancels
    position becomes flat
```

Parent cancellation cancels both waiting children.

Reduce-only children cannot open or reverse exposure.

One child Fill and one sibling cancellation must remain queryable.

Internal arming may update Simulator state.

It must not create exchange-history noise absent from the client-visible contract.

## Output Boundary

Hyperliquid transport returns the `async_hyperliquid`-compatible response shape.

Simulator constructs the same admitted response shape.

Account validates and translates that shape once.

Mutation responses retain:

- outer status;
- response type;
- ordered item statuses;
- CLOID;
- OID;
- filled size;
- average Fill price;
- item error.

HTTP success does not imply Order success.

Every item in `statuses` must be inspected.

One payload-wide pre-validation error may represent the complete batch.

Account expands that error into one ordered rejection per requested item while
retaining the raw response.

Information responses retain the admitted Open Order, Order status, Fill, and account-state shapes.

Simulator-only diagnostics remain outside protocol responses.

Ledger stores normalized domain evidence and optional source evidence.

## Proven Evidence

Resting long and short bracket responses have frozen testnet and Simulator parity.

Both return `resting`, `waitingForFill`, and `waitingForFill`.

Nuubot3 parity captures also prove immediate-filled submission.

Those responses return `filled`, `waitingForTrigger`, and `waitingForTrigger`.

Nuutrader6 implements parent activation, trigger matching, reduce-only protection, and OCO sibling cancellation.

The strict post-trigger fixtures are Simulator-only evidence.

Fresh Hyperliquid testnet proof remains required for post-trigger history and sibling cancellation.

## Reference Files

```text
D:\rust\nuutrader6\src\nuubot\hcbots\simulator.py
D:\rust\nuutrader6\src\nuubot\hcbots\exchange.py
D:\rust\nuubot3\wiki\simparity.md
D:\rust\nuubot3\smoke\simparity.py
D:\rust\nuutrader3-web\research\2026-05-07-simnet-fixture-parity-audit.md
D:\rust\nuutrader3-web\research\fixtures\hype\phase-13-testnet-primitive-suite\repeatable-20260506T141429Z
D:\rust\nuutrader3-web\research\fixtures\hype\phase-13-simnet-primitive-suite\20260507T-strict-parity-002
```

## Proof Ladder

1. Lock Go protocol types to frozen response fixtures.
2. Run deterministic Simulator bracket tests.
3. Compare Simulator and frozen testnet response structures.
4. Compare domain reconciliation from both evidence sources.
5. Run controlled Hyperliquid testnet bracket parity.
6. Admit live mutation only after testnet parity passes.

Testnet proof must cover long and short paths.

It must cover resting entry, immediate Fill, TP Fill, SL Fill, sibling cancellation, and final flat state.

Exact identifiers, timestamps, prices, fees, and balances may differ.

Structure, status meaning, ordering, and lifecycle must match.

## Does Not Claim

- Python implementation parity.
- Exact raw values.
- Mainnet mutation proof.
- Public information endpoints not required by Account.
- Queue position or order-book-depth simulation.
