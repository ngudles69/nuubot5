# Hyperliquid Meta

Status: Approved design. Perpetual retrieval proven in reference code.

Covers: Raw Hyperliquid Meta requests inside `internal/hyperliquid`.

Purpose: Retrieve and validate exchange-owned symbol metadata.

## In

- Explicit perpetual Meta request.
- Required spot Meta request before the Meta package is complete.
- Typed response decoding.
- Universe entries.
- Size decimals.
- Maximum leverage.
- Margin-table identity and tiers.
- Isolation and delisting flags.
- Collateral-token identity.
- Raw exchange response evidence when required by persistence design.

## Out

- Twenty-four-hour freshness admission.
- NuubotDB reads or writes.
- Symbol normalization.
- Price and size policy.
- Minimum-order policy.
- Retirement decisions.
- Setup coordination.
- Background refresh.
- Constructor network calls.

The [Meta package](../packages/meta.md) owns every excluded responsibility.

## Perpetual Flow

```text
Meta caller
  POST {"type":"meta"}
  validate response shape
  decode universe
  decode margin tables
  return typed perpetual Meta
```

## Current Evidence

The audited reference method successfully retrieved live mainnet perpetual Meta.

Observed response:

```text
assets: 232
margin tables: 7
BTC size decimals: 5
BTC maximum leverage: 40
ETH size decimals: 4
ETH maximum leverage: 25
```

This evidence proves endpoint access. Nuubot still requires its own boundary validation and tests.

## Required Proof

- Valid perpetual Meta decodes without information loss.
- Missing or malformed universe fails.
- Malformed margin-table tuples fail without panic.
- Unknown fields do not break compatible responses.
- Context cancellation reaches the REST transport.
- Live mainnet retrieval returns admitted BTC and ETH metadata.
- Spot Meta receives separate proof before admission.
