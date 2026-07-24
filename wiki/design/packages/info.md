# Parity Info Package

Status: Clearinghouse-state probe implemented.
Covers: `internal/parity/info`
Purpose: Capture `/info` payloads and their Nuubot translations.

## Ownership

Info calls the selected production client.

It writes raw payload, normalized state, and one evidence report.

It does not implement Hyperliquid or Simulator behavior.

## Canonical Design

See [Hyperliquid Parity Probe](../hyperliquid/parity.md).
