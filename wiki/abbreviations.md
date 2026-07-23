# Go Abbreviations

These spellings are mandatory when the shortened form is used.

## Go Names

- `cfg` = configuration
- `ctx` = context
- `err` = error
- `req` = request
- `resp` = response
- `mgr` = manager
- `ID` = identifier: `botID`, `sweepID`
- `URL` = uniform resource locator: `apiURL`
- `HTTP` = Hypertext Transfer Protocol: `httpClient`
- `JSON` = JavaScript Object Notation: `rawJSON`
- `DB` = database: `sweepDB`
- `SQL` = Structured Query Language: `sqlQuery`

## Trading Names

- `qty` = quantity
- `BBO` = best bid and offer
- `OHLCV` = open, high, low, close, volume
- `EMA` = exponential moving average
- `RSI` = relative strength index
- `PnL` = profit and loss
- `CLOID` = client order ID: `CLOID` when exported, `cloid` when unexported
- `tf` = timeframe only inside short local calculations; public names MUST use
  `timeframe`
- `ts` = timestamp only inside short local calculations; public names MUST use
  `timestamp`
- `ms` = milliseconds
- `us` = microseconds

## Rules

- Unexported Go names MUST use `camelCase`; exported names MUST use
  `PascalCase`.
- Initialisms MUST keep one spelling: `botID`, not `botId`; `rawJSON`, not
  `rawJson`.
- Database and wire names MUST retain spelling required by the external
  contract, such as `close_time_us`.
- Go identifiers MUST NOT use snake_case.
- A shortened name MUST NOT be invented merely to save characters.
- `required` MUST be written in full. `req` MUST NOT mean `required`.
- Go keywords such as `map`, `range`, `type`, `func`, and `go` MUST NOT be used
  as identifiers.
- Exchange payload fields that the domain does not use MUST remain in the raw
  payload; they MUST NOT become first-class domain fields.
