# Hyperliquid

Status: Official mutation and Recon JSON protocol implemented. Live mutation transport remains pending.

Covers: `internal/hyperliquid/**`

Purpose: Own Nuubot's Hyperliquid protocol boundary.

## Decision

Hyperliquid support stays inside Nuubot5.

No separate Go SDK repository is planned.

Nuubot independently rewrites Hyperliquid support from the
[official Hyperliquid API documentation](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api).

[Sonirico's `go-hyperliquid`](https://github.com/sonirico/go-hyperliquid)
is the secondary implementation reference.

Python `async_hyperliquid` is the third known-working reference.

The locally inspected installed version is `0.4.8`.

Nuubot does not target library parity.

Nuubot copies or adapts audited reference code only when that reduces work.

The package uses Go standard-library REST and `gorilla/websocket`.

## Decision Record

These decisions prevent repeated SDK and dependency discussions.

| Decision | Reason | Reconsider when |
|---|---|---|
| Keep code in `internal/hyperliquid`. | Nuubot5 is the sole consumer. Changes remain atomic with Account and Venue. | A second Go project needs the same boundary. |
| Rewrite from the official API. | Nuubot needs only its admitted behavior. | Never while the official API remains available. |
| Treat Sonirico as secondary reference. | Its Go code proves useful request shapes and behavior. | Official API behavior conflicts with it. |
| Reuse cleaned Sonirico Meta code. | Its perpetual Meta request works against live mainnet. | Official Meta changes or local proof fails. |
| Preserve attribution for copied code. | Sonirico code is MIT licensed. | No Sonirico code remains. |
| Treat `async_hyperliquid` as response reference. | Nuutrader6 uses its client-visible outputs. | A selected output fails official or local proof. |
| Avoid full Python parity. | Python structure and defects do not define Nuubot correctness. | Never; compare only admitted response outputs. |
| Compare deterministic wire outputs selectively. | Payload, hash, digest, signature, and formatting comparisons can expose mistakes. | The official protocol intentionally differs. |
| Match Simulator public response shapes. | Account processes Simulator and live evidence through one translation path. | The admitted response contract changes. |
| Exclude CCXT. | Its unified, generated, multi-exchange surface exceeds current needs. | A second exchange is approved. |
| Avoid a CCXT-style unified layer. | Venue already owns Nuubot's required common boundary. | Venue cannot admit an approved second exchange cleanly. |
| Use standard-library REST. | Hyperliquid REST is bounded JSON POST behavior. | Proven requirements exceed `net/http`. |
| Use `gorilla/websocket`. | It supplies the required socket primitives without a framework. | Proven lifecycle requirements cannot be implemented safely. |
| Own WebSocket recovery locally. | Existing reference lifecycle has races, deadlocks, and incomplete shutdown. | Never without equivalent focused proof. |

Meta reuse means retaining useful request, type, and decoding knowledge.

Nuubot removes hidden constructor requests, panics, unsafe assertions, and unrelated dependencies.

Reference success is evidence, not automatic admission.

## Where

```text
internal/hyperliquid/
  REST transport
  WebSocket transport
  signing
  request and response types
  Hyperliquid-specific validation
  Venue mapping
```

Account uses Hyperliquid through the common Venue boundary.

The current Simulator implements that boundary.

Account sends official action values and receives detached JSON.

`internal/hyperliquid` validates and translates each response.

Meta uses the Hyperliquid information client for raw exchange metadata.

The selected live transport owner uses the Hyperliquid WebSocket boundary.

Shared versus process-local WebSocket ownership remains TBD.

## In

- Mainnet and testnet endpoint selection.
- REST request transport.
- WebSocket connection lifecycle.
- Hyperliquid request signing.
- Hyperliquid request and response types.
- Perpetual and required spot information calls.
- Batch Order submission and cancellation.
- Open Order, Fill, Order-status, and account-state queries.
- Hyperliquid response validation.
- Hyperliquid errors and rate-limit evidence.
- Mapping between Venue values and Hyperliquid payloads.
- Protocol response types constructed by Simulator for parity proof.

## Implemented Protocol

`protocol.go` owns official values for:

- Order requests and limit or trigger types;
- order actions and batch grouping;
- cancel-by-CLOID actions;
- submit and cancel acknowledgements;
- open Orders;
- exact Order status; and
- user Fills.

`state.go` owns clearinghouse response validation.

Every Nuubot Order carries one mandatory official CLOID.

CLOID remains opaque to Hyperliquid protocol and Simulator.

Venue assigns OID after accepting the request.

Submit responses preserve item order and expose assigned OID.

Open and exact status rows may include CLOID.

Official Fill rows may omit CLOID.

Account resolves those rows through OID and validates TID uniqueness.

`Encode` creates fresh JSON.

Strict decoders reject malformed envelopes, missing identity, and incomplete rows.

Protocol types contain no Ledger, Trade, local Order, role, or purpose fields.

## Out

- Bot, Controller, BotCycle, Executor, Risk, or strategy policy.
- Account, Ledger, Trade, Order, or Fill ownership.
- Meta caching, refresh admission, normalization, or persistence.
- Shared subscription policy or subscriber ownership.
- Simulator behavior.
- Reconciliation decisions.
- Minimum-order policy.
- Credentials storage.
- Mainnet fallback when configuration is missing.
- Automatic mutation retries.
- Background work started by constructors.

## Construction

Constructors allocate local state only.

Constructors do not contact Hyperliquid, start goroutines, panic, or read credentials.

Network work accepts `context.Context` and returns errors.

## Internal Design

- [REST](hyperliquid/rest.md) owns HTTP transport and explicit calls.
- [Exchange](hyperliquid/exchange.md) owns signed mutations and application responses.
- [WebSocket](hyperliquid/websocket.md) owns socket mechanics and recovery.
- [Meta](hyperliquid/meta.md) owns raw Hyperliquid Meta retrieval.
- [Parity Probe](hyperliquid/parity.md) owns permanent API and Simulator evidence.

Add another page only when one implemented concern needs its own contract.

## Reuse Policy

Reuse is preferred when the code is small, safe, required, and proven.

Patch isolated defects in place.

Replace code with tangled ownership, unsafe failure behavior, or larger dependencies than the required function.

Copied MIT code retains required license and attribution.

## Reference Sources

Authoritative protocol source:

- [Official Hyperliquid API documentation](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api)

Supplemental implementation reference:

- [Sonirico's Go Hyperliquid client](https://github.com/sonirico/go-hyperliquid)
- Python `async_hyperliquid`.

```text
D:\rust\references\go-hyperliquid
D:\rust\go-hyperliquid
D:\rust\nuutrader6\.venv\Lib\site-packages\async_hyperliquid
D:\rust\nuubot3\nuubot\exchange\async_hyperliquid.py
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
```

Nuutrader6 is the matching-behavior reference.

`async_hyperliquid` 0.4.8 is the client-visible response reference.

Reference behavior does not override official semantics, Nuubot ownership, or safety rules.

Simulator may reuse admitted Hyperliquid protocol types.

It must not import REST, signing, WebSocket, or live-client behavior.

See [Simulator Parity](concepts/simulator-parity.md).

## Required Proof

- Build and tests pass with `CGO_ENABLED=0` and `-tags noasm`.
- Mainnet and testnet selection is explicit.
- External responses fail closed when malformed.
- Constructors perform no network work.
- Logs contain no credentials, signatures, or full private request bodies.
- Queries, signing, Orders, cancellations, and WebSockets receive focused proof before admission.
