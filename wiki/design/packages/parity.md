# Parity Package

Status: Testnet clearinghouse-state probe implemented.
Covers: `internal/parity`
Purpose: Admit and run one permanent parity-probe request.

## Ownership

Parity owns target admission, credentials selection, operation dispatch, and
terminal probe statistics.

Endpoint packages own their probe mechanics.

Production Hyperliquid and Simulator packages own tested behavior.

## Lifecycle

`ParseInput` admits the command shape.

`Init` loads immutable inputs and selects one real target.

`Run` executes one operation.

Empty `Start`, `Loop`, and `Stop` phases are prohibited.

## Canonical Design

See [Hyperliquid Parity Probe](../hyperliquid/parity.md).
