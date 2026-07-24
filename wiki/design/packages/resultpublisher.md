# ResultPublisher Package

Status: Implemented.
Covers: `internal/resultpublisher/*.go`
Purpose: Publish successful memory-only BtRunner results as one completed SQLite database.

## Ownership

BtRunner calls ResultPublisher after replay verification and Runtime shutdown.

ResultPublisher owns the temporary file and final rename.

Ledger and Simulator own their table serialization.

## Program Flow

```text
Publish
  select memory-only Account results
  prepare temporary result path
  publish Account children
  publish completed result
```

`persist_mode = max` requires no terminal export.

`persist_mode = none` writes `bot_<id>.db.partial`, closes it, then replaces
`bot_<id>.db`.

Failure removes only the partial file.

## Proof

Sweep `9`, Bot `13` published 50 Ledgers, 50 Trades, 151 Orders, 100 Fills,
and 50 Simulator states.

SQLite integrity and foreign-key checks passed.
