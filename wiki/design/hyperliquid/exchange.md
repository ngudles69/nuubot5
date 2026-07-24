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
