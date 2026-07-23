# Meta Package

Status: Reserved.
Covers: `internal/meta/doc.go`
Purpose: Own validated market instrument metadata.

## Canonical Sources

- Nuutrader6 reference: `src/nuubot/hcserver/hc_meta.py`

## Responsibilities

- Fetch venue instrument metadata.
- Validate external response shapes.
- Normalize perpetual and spot definitions.
- Preserve venue identifiers, precision, leverage, and status.
- Store and load normalized metadata.

## Does Not

- Own market events.
- Own accounts, orders, trades, or fills.
- Execute orders.
- Contain strategy policy.

## Required Proof

- Invalid venue responses fail validation.
- Normalized identifiers and precision match venue truth.
- Stored metadata reloads without information loss.
