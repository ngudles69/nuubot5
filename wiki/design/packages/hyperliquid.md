# Hyperliquid Package

Status: Clearinghouse-state REST slice implemented.
Covers: `internal/hyperliquid`
Purpose: Own Nuubot's Hyperliquid protocol boundary.

## Ownership

Hyperliquid owns transport, exact wire fields, response admission, and readable
Nuubot translations.

It owns no Account, Simulator, trading policy, or credentials storage.

## Implemented

- Explicit mainnet or testnet Client construction.
- Bounded JSON `Post`.
- Raw perpetual clearinghouse payload retrieval.
- Exact decimal translation into `AccountState`.

## Canonical Design

See [Hyperliquid](../hyperliquid.md), [REST](../hyperliquid/rest.md), and
[Exchange](../hyperliquid/exchange.md).
