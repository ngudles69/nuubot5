# Meta Package

Status: Implemented for mainnet perpetual Meta. Spot Meta is pending.
Covers: `internal/meta/*.go`
Purpose: Own Hyperliquid symbol reference metadata and normalized trading constraints.

## Canonical Sources

- Nuutrader6 reference: `src/nuubot/hcserver/hc_meta.py`

## Scope

Meta is the reference table for Hyperliquid symbols.

The current implementation admits mainnet perpetual symbols only.

Tables live in the configured shared Nuubot database.

Meta changes rarely. Symbols may still be added, retired, delisted, or have
their exchange constraints changed.

## Owner and Children

Setup calls Meta freshness admission.

Meta owns exchange fetching, validation, normalization, persistence, and
symbol lookup.

Meta uses NuubotDB and the [internal Hyperliquid information client](../hyperliquid/meta.md).

## Responsibilities

- Fetch the complete Hyperliquid mainnet perpetual Meta dataset.
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

Stored Meta identity is:

```text
network + kind + symbol
```

Runtime and Account network settings never select the Meta source.

Meta always refreshes from mainnet.

Tests needing changed Meta edit their local SQLite fixture manually.

## Setup Freshness Contract

Every Setup caller checks mainnet Meta freshness.

Meta refreshes when its dataset is empty or its last successful refresh is at
least 24 hours old.

The first Setup caller after expiry performs the refresh. Fresh callers perform
no exchange request.

Freshness belongs to the complete dataset for one network. It must not depend
on individual row update timestamps.

Only a successful complete refresh advances the dataset refresh timestamp.

The SQLite immediate writer transaction serializes concurrent refresh admission.

## Program Flow

```text
EnsureFresh
  validate Meta request
  open shared database
  prepare Meta tables
  claim network refresh
  read dataset freshness
  refresh stale dataset
  load admitted symbol
  commit Meta admission
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
- Setup requests mainnet regardless of Runtime network.
- Missing symbols become retired only after a successful full refresh.
- Final rounded order notional respects the configured USDC 11 floor.

Refresh failure fails Setup. Stale Meta is not silently admitted.
