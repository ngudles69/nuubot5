# Hyperliquid Exchange

Status: Approved response design. Implementation pending.

Covers: Future `internal/hyperliquid` exchange operations and `internal/parity/exchange`

Purpose: Define signed mutation responses, application errors, and parity evidence.

Official sources:

- [Exchange endpoint](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/exchange-endpoint)
- [Error responses](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/error-responses)
- [Rate limits](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/rate-limits-and-user-limits)

## Error Layers

Transport success does not mean Order success.

An HTTP 200 place-order response may contain:

```text
status=ok
response.type=order
response.data.statuses[0].error=Order must have minimum value of $10.
```

The outer `status=ok` means Hyperliquid processed the action envelope.

Each item status reports whether its requested Order rested, filled, or failed.

Account must inspect every item and persist ordered acknowledgement evidence.

The immediate response is acknowledgement evidence.

[Historical Orders](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint)
return at most 2,000 recent rows.

The same official endpoint limits Fill history to 2,000 rows per response.

Time-filtered Fill history exposes only the 10,000 most recent Fills.

Account filters Fill requests by inclusive cursor and bounded time range.

History omission never proves local evidence invalid.

It is not final Order, Fill, position, or balance truth.

Reconciliation by CLOID, OID, Fill, and account state repairs later drift.

## Filled Acknowledgement and Fee Completion

Hyperliquid may acknowledge an immediate execution as filled before fee-bearing Fill evidence becomes available.

Filled acknowledgement makes the Order terminal at Hyperliquid. It does not make Nuubot reconciliation complete.

Nuubot keeps the Order reconciliation-pending until:

- matching Fill evidence exists;
- filled quantity is complete; and
- every Fill has explicit fee metadata.

A filled acknowledgement without a Venue TID does not create a synthetic Fill.

A Fill with a Venue TID but no fee remains fee-incomplete with `Fill.HasFee=false`.

The Fill execution is immutable. It is not an open execution and must not be applied twice.

Later evidence with the same Venue TID may enrich fee, liquidity, and raw metadata only.

Quantity, price, side, identity, and execution timestamp must remain unchanged.

Reconciliation keeps fee-incomplete Fill identities in its working set.

Reconciliation uses two independent inclusive history tracks:

```text
new Fill discovery  committed Fill cursor -> observed time
pending fee repair  bounded windows containing unresolved timestamps
```

The two result sets merge by Venue TID before Ledger admission.

The new-Fill cursor may advance after complete reconciliation. It never replaces or clears pending-fee timestamp anchors.

A filled acknowledgement without Venue TID uses its Order acknowledgement timestamp as a repair anchor.

After a matching TID appears, that Fill timestamp owns its repair anchor.

Nearby anchors share one bounded repair window. Widely separated anchors use separate windows.

Repair does not repeatedly query from the oldest pending timestamp through the present.

Both tracks split capped responses. An old pending Fill must not block discovery of newer Fills.

The pending identity leaves that set only when `Fill.HasFee=true`.

Each Fill preserves durable fee provenance:

```text
HasFee
FeePendingSinceMS
FeeResolvedMS
```

The fields answer whether one Fill received fees initially, later, or not yet:

```text
HasFee=true  FeePendingSinceMS=0  fee arrived initially
HasFee=true  FeePendingSinceMS>0  fee was found later
HasFee=false FeePendingSinceMS>0  fee remains missing
```

`FeeResolvedMS` records when reconciliation changed missing fee metadata to present.

Search attempts belong in cycle telemetry. Recon must not dirty every pending Fill merely to increment an attempt counter.

Grid/Hedge STOP remains reconciling while any position-close Fill is missing or fee-incomplete.

This rule applies to market and marketable-limit closure mechanics.

Fee presence is explicit. Nuubot never infers presence from the fee value.

Zero fees are valid. Negative fees may represent rebates.

## Fee Completion Observability

Every admitted Fill observation has exactly one outcome: added, enriched, or unchanged duplicate.

A conflicting execution or changed existing fee fails reconciliation. It is never counted as unchanged.

Fill transition logs use explicit identities and state:

```text
fill added venue_tid=12345 has_fee=false
fill fee enriched venue_tid=12345 previous=missing fee=12.32
```

Logs contain no complete private Venue payload.

Each reconciliation cycle identifies Venue, network, Account, symbol, and observation timestamp.

Its `fill_queries` array contains one entry for every physical Fill-history request, including failed requests:

```text
kind                 discovery | repair | super | stop
request_started_ms   local request start
start_ms             requested inclusive history start
end_ms               requested inclusive history end
rows_downloaded      raw Venue rows
rows_deduplicated    unique Venue TIDs
rows_for_symbol      rows matching this Account symbol
fills_added          new local Fills
fills_unchanged      unchanged duplicate Fills
fees_enriched        missing-to-present fee transitions
pending_matched      pending identities returned
cap_reached          response reached the Venue row cap
duration_ms          request start through return
error                 bounded failure classification
```

One heartbeat telemetry row may contain multiple query entries. It does not require one database row per request.

Cycle totals record fee-incomplete Fills before and after repair, high-water count, oldest age, cursor movement, cap splits, and rate-limit evidence.

Charts use request start as time, network and kind as series, and rows or duration as values.

This reveals ordinary ranges, isolated spikes, sustained growth, repair effectiveness, and testnet/mainnet differences.

Discovery and repair durations measure immediately before each Venue request through its return.

Fee-resolution lag measures first local observation without fee through the successful missing-to-present enrichment.

Run summaries report initial fee presence, fees found later, fees still missing, repair cycles, and fee-resolution lag distribution.

These fields support separate testnet and mainnet baselines without assuming their behavior matches.

The counts must classify every successfully admitted Fill observation without silent omission.

An unchanged fee-incomplete Fill remains pending. A failed reconciliation does not clear its identity or move its search boundary forward.

## Delayed-Fee Regression Proof

The required test makes fee availability lag behind the Fill execution timestamp and normal discovery cursor.

```text
cycle 1  Fill timestamp=1000  fee unavailable
cycle 2  discovery start=2000  fee now available for timestamp=1000
```

The Venue fixture returns the delayed fee only when a request includes timestamp 1000.

Proof must show:

1. The initial Fill is added once with `Fill.HasFee=false`.
2. The new-Fill discovery cursor advances beyond timestamp 1000.
3. The discovery query cannot return the delayed fee.
4. Pending-fee repair still starts at timestamp 1000.
5. Repair re-sees the same Venue TID with its fee.
6. The existing Fill enriches without duplicating quantity, notional, or fee.
7. Logs identify the Venue TID and missing-to-present transition.
8. Cycle telemetry records one enrichment and both query boundaries.
9. The Order leaves reconciliation-pending only after every owned Fill has `Fill.HasFee=true`.
10. Grid/Hedge STOP waits before enrichment and completes after clean reconciliation.

A one-cursor implementation must fail this test.

The official reference payload is
[`official-min-notional-error.json`](json/exchange/place-order/official-min-notional-error.json).

## Batched Errors

Order and cancel errors usually match the request batch length.

Some payload-wide pre-validation failures return one error for the entire batch.

Nuubot preserves the raw response.

Account expands one payload-wide error across every ordered request result.

That produces exactly one explicit outcome per requested item.

Batch order remains unchanged.

## Minimum Notional

Hyperliquid currently rejects perpetual Orders below USDC 10.

The documented application error is:

```text
Order must have minimum value of $10.
```

Nuubot configures USDC 11 before price and size rounding.

The buffer reduces ordinary minimum-notional rejection.

The response must still be handled because prices and exchange policy can change.

## Rate Limits

Official limits are external policy and may change.

Current documented IP weight:

- Aggregated REST limit: 1,200 per minute.
- `clearinghouseState`: weight 2.
- Exchange action: `1 + floor(batch_length / 40)`.

Address-based mutation limits also apply.

Info requests do not consume the address-based action limit.

Automatic mutation retry is prohibited.

A timeout may hide an accepted mutation.

Recovery must query by CLOID and reconcile Venue truth.

Read-only retry policy remains deferred until observed headers and failures are captured.

## Parity

Future exchange probes belong under:

```text
internal/parity/exchange/
```

Evidence belongs under:

```text
wiki/design/hyperliquid/json/exchange/
```

Required scenarios:

- resting Order;
- immediately filled Order;
- item-level rejection;
- payload-wide rejection;
- cancel success;
- cancel rejection;
- HTTP failure;
- rate-limit response;
- timeout with unknown mutation outcome.

Testnet and Simulator must produce the same admitted ordered outcomes.

Exact identifiers, timestamps, prices, balances, and rate-limit counters may differ.
