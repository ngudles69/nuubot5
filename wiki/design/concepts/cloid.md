# CLOID

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Encode and decode one fixed 128-bit Hyperliquid client Order identity.

## Scope

CLOID is a stateless utility.

Account owns CLOID creation for domain Orders.

Nuubot4 exports and tests the codec. Its current Runtime path does not call it.

## Layout

| Field | Bits | Allowed values |
|---|---:|---:|
| `botcycle_id` | 24 | 0 to 16,777,215 |
| `symbol_id` | 16 | 0 to 65,535 |
| `exchange` | 4 | 0 to 15 |
| `network` | 2 | 0 to 3 |
| `side` | 1 | 0 to 1 |
| `reduce_only` | 1 | false or true |
| `purpose` | 8 | 0 to 255 |
| `trade_no` | 21 | 1 to 2,097,151 |
| `batch_no` | 10 | 1 to 1,000 |
| `order_pos` | 10 | 1 to 1,000 |
| `timestamp_s` | 31 | 0 to 2,147,483,647 |

Fields pack from highest to lowest bit in table order.

## Identity

`trade_no` is the BotCycle-local Trade number.

`trade_no` MUST NOT equal or replace datastore `trade_id`.

`batch_no` increments per Trade submission.

`order_pos` identifies one request position inside its batch.

Timestamp is real Unix seconds. It MUST NOT be altered to force uniqueness.

## Responsibilities

- Validate every field before encoding.
- Produce `0x` followed by exactly 32 lowercase hexadecimal characters.
- Decode only the current fixed layout.
- Reject malformed shape and invalid decoded ranges.
- Round-trip every accepted field exactly.

## Does Not

- Own lifecycle or mutable state.
- Allocate `trade_id`.
- Reserve `trade_no`.
- Submit Orders.
- Read configuration or credentials.
- Provide legacy compatibility decoding.

## Invariants

- Encoding MUST never truncate or wrap.
- Decoding MUST reject invalid field ranges.
- There is one current layout.
- Strategy code MUST NOT create or supply CLOIDs.
- Account MUST generate CLOID only after Trade and batch identity exist.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\cloid.md
D:\rust\nuubot4\src\cloid.rs
D:\rust\nuubot4\src\lib.rs
```

Supplemental:

```text
D:\rust\nuubot3\wiki\cloid.md
D:\rust\nuubot3\wiki\account\account.md
```

## Conflict

Nuutrader6 contains a different CLOID layout. Nuubot5 MUST use the Nuubot4 layout unless the user explicitly replaces it.

## Recommendation

Port Nuubot4's codec directly into safe, standard Go. Do not add compatibility or runtime wiring before Account needs it.
