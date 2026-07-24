# Hyperliquid REST

Status: Clearinghouse-state read implemented. Remaining calls pending.

Covers: `internal/hyperliquid` REST transport.

Purpose: Send bounded Hyperliquid JSON payloads and translate admitted responses.

Protocol source:
[official Hyperliquid API documentation](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api).

## In

- Go standard-library `net/http`.
- Explicit mainnet or testnet base URL.
- Caller-owned `context.Context`.
- Explicit positive request timeout.
- JSON request encoding.
- HTTP status handling.
- Bounded response reading.

## Out

- WebSocket connections.
- Request signing.
- Credentials loading.
- Constructor network calls.
- Panics or fatal process exits.
- Full request or response body logging.
- Automatic retries for mutations.
- Implicit network selection.

## Flow

```text
New
  validate network
  configure HTTP client

ClearinghouseState
  request clearinghouse payload
  decode clearinghouse payload

ClearinghouseStatePayload
  validate address
  encode request payload
  post request payload

Post
  create request
  send request
  read response payload
  validate response

DecodeClearinghouseState
  decode response payload
  translate account state
```

## Construction

Client construction stores an admitted base URL and HTTP client.

Construction performs no DNS lookup, connection, request, authentication, or background work.

## Logging

The REST Client performs no logging.

Logs must not include credentials, signatures, private actions, or complete response bodies.

## Retry

The first implementation performs no automatic retries.

Safe query retry policy may be added after rate-limit and timeout behavior is proven.

Mutations require request-identity and unknown-outcome design before any retry.

## Clearinghouse State

The first implemented call sends:

```json
{
  "type": "clearinghouseState",
  "user": "<address>",
  "dex": ""
}
```

Private wire types retain Hyperliquid field names.

Exported Nuubot types use readable names and `decimal.Decimal`.

| Hyperliquid | Nuubot |
|---|---|
| `marginSummary.accountValue` | `Margin.Equity` |
| `marginSummary.totalMarginUsed` | `Margin.MarginUsed` |
| `marginSummary.totalNtlPos` | `Margin.Notional` |
| `marginSummary.totalRawUsd` | `Margin.RawUSD` |
| `crossMarginSummary` | `CrossMargin` |
| `crossMaintenanceMarginUsed` | `MaintenanceMargin` |
| `assetPositions[].position.szi` | `Positions[].SignedSize` |
| `entryPx` | `EntryPrice` |
| `liquidationPx` | `LiquidationPrice` |
| `positionValue` | `Notional` |
| `unrealizedPnl` | `UnrealizedPnL` |

`RawUSD` remains untranslated because its stronger semantic meaning is unproven.

Live captures belong under `wiki/design/hyperliquid/json`.

The permanent [Parity Probe](parity.md) records the raw payload before decoding.

## Required Proof

- Context cancellation stops an active request.
- Positive timeouts are required.
- Non-success status returns an error.
- Oversized responses fail.
- REST transport logs no payloads.
- Clearinghouse state succeeds for approved testnet accounts.
- Go and `async_hyperliquid` output values match.

## Proven Baseline

Capture: `20260724-clearinghouse-baseline`.

Both approved testnet accounts returned HTTP 200.

Go and `async_hyperliquid` matched every field and value except the expected
request-time field.

```text
tgrid   positions=0 equity=172.232247 duration_ms=165
thedge  positions=0 equity=549.237687 duration_ms=150
```

Evidence:
`json/info/clearinghouse-state/20260724-clearinghouse-baseline/testnet`.
