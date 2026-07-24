# Hyperliquid Parity Probe

Status: Testnet clearinghouse-state probe implemented. Simulator path pending.

Covers: `cmd/parity-probe`, `internal/parity`, `internal/parity/info`

Purpose: Permanently exercise production Hyperliquid and Simulator paths and preserve API-drift evidence.

## Command

```text
parity-probe <network> <account> <operation> [arguments...]
```

Implemented:

```powershell
parity-probe testnet tgrid clearinghouse-state
parity-probe testnet thedge clearinghouse-state baseline
```

The optional clearinghouse argument names the capture.

Capture names admit only letters, numbers, hyphens, and underscores.

`simnet` becomes executable only after Simulator implements real
clearinghouse-state behavior.

No fake Simulator adapter or placeholder response is allowed.

## Ownership

```text
cmd/parity-probe/
  main.go

internal/parity/
  parity.go
  info/
    info.go
    clearinghouse.go

internal/hyperliquid/
  live and testnet behavior

internal/simulator/
  simnet behavior
```

`cmd/parity-probe` owns the process boundary.

`internal/parity` owns input admission, target selection, operation dispatch,
and terminal evidence summaries.

`internal/parity/info` owns probes for Hyperliquid's `/info` endpoint.

Future `/exchange` probes belong in `internal/parity/exchange`.

Meta and clearinghouse probes remain files inside `internal/parity/info`.

Parity code never implements exchange or Simulator behavior.

## Lifecycle

One-shot REST probes use:

```text
Init
Run
```

WebSocket probes may use:

```text
Init
Start
Loop
Stop
```

Empty lifecycle phases are omitted.

Protocol operations retain protocol terms:

```text
HTTP       Get Post
WebSocket  Connect Read Write Disconnect
Payload    request payload response payload
```

## Program Flow

```text
main
  open parity log
  parse input
  initialize parity probe
  run parity probe
  log result

Probe Init
  select target
  load shared config
  load credentials
  select account
  initialize Hyperliquid client
  select capture

ClearinghouseState
  reserve evidence directory
  get clearinghouse payload
  preserve any returned raw payload
  decode clearinghouse payload
  write normalized payload
  write result report
```

## Evidence

```text
wiki/design/hyperliquid/json/
  info/
    clearinghouse-state/
      <capture>/
        <network>/
          <account>/
            raw.json
            normalized.json
            report.json
            async-hyperliquid.json
            comparison.json
```

This operator-only path is the approved exception to the mutable
[`workspace/`](../concepts/filesystem.md) contract.

`raw.json` preserves the received payload bytes.

`normalized.json` contains translated Nuubot values.

`report.json` records target identity, operation, status, time, and duration.

Failed requests or decoding write a failure report.

Any received payload is preserved before decoding.

Failed normalization leaves `normalized.json` absent.

Reference captures may add `async-hyperliquid.json` and `comparison.json`.

An existing account capture directory fails closed without overwriting evidence.

Files contain no API keys or signing material.

Simnet never loads Hyperliquid credentials.

## Comparison

The harness eventually runs identical scenarios through:

```text
testnet  internal/hyperliquid
simnet   internal/simulator
```

Compare payload shape, field meaning, decimal text, ordering, state
transitions, Orders, Fills, positions, and balances.

Ignore admitted nondeterminism such as timestamps, exchange-generated
identities, market prices, fees, and capture duration.

## Required Proof

- Invalid network, operation, account, and capture names fail.
- Requests use production Hyperliquid or Simulator code.
- Raw payload bytes survive capture unchanged.
- Existing evidence cannot be overwritten.
- Normalized output uses exact decimal values.
- Evidence contains no credentials.
- Both approved testnet accounts succeed.
- Go and `async_hyperliquid` match except admitted nondeterminism.
- Simulator comparison starts only after real Simulator behavior exists.
