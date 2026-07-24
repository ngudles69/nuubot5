# Meta Package

Status: Reserved.
Covers: `internal/meta/doc.go`
Purpose: Own Hyperliquid symbol reference metadata and normalized trading constraints.

## Canonical Sources

- Nuutrader6 reference: `src/nuubot/hcserver/hc_meta.py`

## Scope

Meta is the reference table for Hyperliquid perpetual and spot symbols.

The table belongs inside NuubotDB.

Meta changes rarely. Symbols may still be added, retired, delisted, or have
their exchange constraints changed.

## Owner and Children

Setup calls Meta freshness admission.

Meta owns exchange fetching, validation, normalization, persistence, and
symbol lookup.

Meta uses NuubotDB and the [internal Hyperliquid information client](../hyperliquid/meta.md).

## Responsibilities

- Fetch the complete Hyperliquid perpetual and spot Meta datasets.
- Validate exchange response shapes before persistence.
- Normalize every admitted symbol.
- Preserve network, kind, symbol, asset ID, exchange index, leverage, and status.
- Preserve price decimals and price-rounding constraints.
- Preserve size decimals and size-rounding constraints.
- Preserve minimum and maximum size when the exchange source provides them.
- Preserve raw exchange data for later parity checks.
- Mark previously known missing symbols retired after a successful full refresh.
- Load normalized symbol Meta for downstream order construction.

## Identity

Meta identity is at least:

```text
network + kind + symbol
```

Mainnet and testnet rows must not overwrite each other.

## Setup Freshness Contract

Every Setup caller checks Meta freshness.

Meta refreshes when its dataset is empty or its last successful refresh is at
least 24 hours old.

The first Setup caller after expiry performs the refresh. Fresh callers perform
no exchange request.

Freshness belongs to the complete dataset for one network. It must not depend
on individual row update timestamps.

Only a successful complete refresh advances the dataset refresh timestamp.

Concurrent Setup callers must not perform duplicate refreshes for the same
network. The exact refresh-claim mechanism depends on the NuubotDB design.

## Program Flow

```text
EnsureFresh
  read network dataset state
  return when data is present and younger than 24 hours
  fetch perpetual Meta
  fetch spot Meta
  validate responses
  normalize symbols
  upsert symbols
  mark missing symbols retired
  record successful refresh time

LoadSymbol
  query network, kind, and symbol
  return normalized Meta
```

## Minimum Order Notional

Hyperliquid's minimum order notional is USDC 10.

Nuubot configures `hyperliquid.min_order_notional_usdc = 11`.

The extra USDC 1 buffers price movement plus price and size rounding between
order construction and exchange acceptance.

Meta supplies the precision and rounding constraints. Config supplies the
USDC 11 policy floor.

Order construction must round price and size, recalculate final notional, and
increase size one valid step when rounded notional is below USDC 11.

The buffer reduces minimum-notional rejection risk. It cannot guarantee
acceptance after a larger price move.

## Does Not

- Own market events.
- Own accounts, orders, trades, or fills.
- Execute orders.
- Contain strategy policy.
- Validate account credentials.
- Refresh in a background loop.
- Refresh on every Setup call.

## Nuutrader6 Difference

Nuutrader6 loads Meta only when its table is empty.

Nuubot5 adds caller-driven refresh after 24 hours.

## Required Proof

- Empty Meta triggers one complete refresh.
- Meta younger than 24 hours performs no exchange request.
- Meta at least 24 hours old triggers one complete refresh.
- Concurrent stale callers produce one refresh per network.
- Failed refresh does not advance freshness.
- Invalid venue responses fail validation.
- Normalized identifiers and precision match venue truth.
- Stored metadata reloads without information loss.
- Mainnet and testnet rows remain distinct.
- Missing symbols become retired only after a successful full refresh.
- Final rounded order notional respects the configured USDC 11 floor.

## Open Decision

Decide whether Setup fails or uses existing stale Meta when a refresh fails.
