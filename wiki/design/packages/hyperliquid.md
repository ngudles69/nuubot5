# Hyperliquid Package

Status: Public Info endpoint implemented; WebSocket and Exchange remain unavailable.
Covers: `internal/hyperliquid`
Purpose: Own Nuubot's Hyperliquid protocol boundary.

## Ownership

Hyperliquid owns transport, exact wire fields, response admission, and readable
Nuubot translations.

It owns no Account, Simulator, trading policy, or credentials storage.

## Boundaries

```text
endpoint-info.go
  public credential-free REST endpoint
  Meta and clearinghouse-state

endpoint-ws.go
  Runner-owned lifecycle stub

endpoint-exchange.go
  future credentialed Account-owned trading endpoint
```

## Implemented

- Explicit mainnet or testnet Info construction.
- Bounded JSON `Post`.
- Public perpetual Meta retrieval and translation.
- Public clearinghouse-state retrieval by address.
- Exact decimal translation into `AccountState`.

Runner owns one shared Info object using the configured network.

Meta refresh owns a separate Info object hardcoded to mainnet.

Exchange remains a reservation file. WebSocket Start fails explicitly as unimplemented.

## Canonical Design

See [Hyperliquid](../hyperliquid.md), [REST](../hyperliquid/rest.md), and
[Exchange](../hyperliquid/exchange.md).
