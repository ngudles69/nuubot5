# Hyperliquid REST

Status: Approved design. Implementation pending.

Covers: `internal/hyperliquid` REST transport.

Purpose: Send bounded Hyperliquid JSON requests and return validated response bytes.

Protocol source:
[official Hyperliquid API documentation](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api).

## In

- Go standard-library `net/http`.
- Explicit mainnet or testnet base URL.
- Caller-owned `context.Context`.
- Thirty-second default request timeout.
- Optional injected `http.Client` for tests and approved configuration.
- JSON request encoding.
- HTTP status handling.
- Bounded response reading.
- Hyperliquid error decoding.

## Out

- WebSocket connections.
- Request signing.
- Domain mapping.
- Credentials loading.
- Constructor network calls.
- Panics or fatal process exits.
- Full request or response body logging.
- Automatic retries for mutations.
- Implicit network selection.

## Flow

```text
caller method
  encode request
  create request with context
  send through owned HTTP client
  read bounded response
  reject HTTP or Hyperliquid error
  return response for typed decoding
```

## Construction

Client construction stores an admitted base URL and HTTP client.

Construction performs no DNS lookup, connection, request, authentication, or background work.

## Logging

Debug logs may include method, endpoint class, status, duration, and response size.

Logs must not include credentials, signatures, private actions, or complete response bodies.

## Retry

The first implementation performs no automatic retries.

Safe query retry policy may be added after rate-limit and timeout behavior is proven.

Mutations require request-identity and unknown-outcome design before any retry.

## Required Proof

- Context cancellation stops an active request.
- Default timeout is present.
- Injected HTTP clients work.
- Non-success status returns an error.
- Oversized responses fail.
- Debug logging excludes bodies.
- Perpetual Meta succeeds against live mainnet.
