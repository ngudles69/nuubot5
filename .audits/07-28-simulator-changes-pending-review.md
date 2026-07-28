# Simulator Changes - Pending Review

Status: **PENDING REVIEW**

Topics 1 through 17 remain pending review.

## Result

```text
Applied pending review: SetFillFeeAvailableForTest deletion
Current assessment:     No other dead Simulator function
Pending review:         Every topic and proposed behavior change
```

## 1. Dead Code - Pending Review

No whole production file is deleted.

Current assessment, pending user review:

- `(*Simulator).SetFillFeeAvailableForTest` - tests were its only caller.

Every other previously proposed deletion has callers:

- `(*Simulator).matchAdded` - called by `PlaceOrders`.
- `(*Simulator).sortedActiveOrders` - called by `OpenOrders`, `match`, and `cancelChildren`.
- `(*Simulator).position` - called by `restore`.
- `(*Simulator).persist` - called by `PlaceOrders`, `CancelOrders`, `Stop`, and `onBBO`.
- `(*Simulator).storedState` - called by `Init` and `persist`.
- `(*Simulator).stage` - called by `PlaceOrders`, `CancelOrders`, and `onBBO`.
- `(*Simulator).commit` - called by `PlaceOrders`, `CancelOrders`, and `onBBO`.
- `(*Simulator).restore` - called by `Init`.
- `restoreOrder` - called by `restore`.
- `restoreFill` - called by `restore`.
- `orderPrice` - called by response generation, Order construction, Fill creation, and recovery.

Delete or rewrite tests that require `SetFillFeeAvailableForTest`.

**KEY DECISION:** Tests never add or control production behavior.

**KEY DECISION:** Used code is not dead. Removing used code requires an approved behavior change and replacement path.

## 2. Prepare Records for Future Persistence

Keep Simulator records complete and stable enough to persist later:

- balance and position;
- open and closed Orders;
- Fills;
- OID and TID;
- CLOID, Venue OID, and Venue TID;
- submitted, updated, and Fill values; and
- raw payloads.

Do not add schema, store calls, database IDs, dirty tracking, recovery, or persistence behavior now.

**KEY DECISION:** Simulator is memory-only in this implementation.

**FUTURE TODO:** Design Simulator persistence separately. A future schema mismatch must fail fast with no repair or compatibility path.

## 3. No Persistence

Delete Simulator persistence behavior from `simulator.go`:

- `persist`
- `storedState`
- `stage`
- `commit`
- `restore`
- `restoreOrder`
- `restoreFill`

Do not add `Persist(mode)`.

Do not change `store.go` during this implementation.

## 4. Change `(*Simulator).Init`

Pseudocode:

```text
Init(config):
    validate config
    initialize balance from config starting balance
    initialize position, Orders, Fills, OID, and TID
    subscribe to MarketData
    return success
```

`Init` performs no database work and no recovery.

**KEY DECISION:** Starting balance comes from Simulator Config.

## 5. Change Simulator Memory

Replace the current combined maps and full history staging with:

```text
openOrdersByOID
closedOrdersByOID
latestOIDByCLOID
fills
positionsBySymbol
```

Moving an Order from open to closed preserves:

```text
Venue OID
CLOID
all original and terminal data
raw payload
```

**KEY DECISION:** Duplicate CLOIDs are accepted. Simulator does not reject them.

**KEY DECISION:** Venue OID is the canonical internal Order key.

## 6. Change Order Submission

Functions:

- `(*Simulator).PlaceOrders`
- `(*Simulator).validateRequest`
- `(*Simulator).newOrder`
- `(*Simulator).matchAdded`

Pseudocode:

```text
PlaceOrders(action):
    validate Simulator lifecycle
    validate each request using Simulator exchange rules

    for each request in order:
        allocate Venue OID
        create Order record

        if valid Venue rejection:
            store closed rejected Order
            return item rejection
            continue

        store open Order
        return one ordered status item

    return exact Hyperliquid-shaped batch response
```

Valid Venue rejection consumes CLOID and Venue OID.

Simulator does not assume what Account validated.

**KEY DECISION:** Simulator owns its request validation because it simulates the Exchange.

**KEY DECISION:** One submitted Hyperliquid batch returns the exact Hyperliquid batch response.

## 7. Change Cancellation

Functions:

- `(*Simulator).CancelOrders`
- `(*Simulator).cancel`
- `(*Simulator).cancelChildren`

Pseudocode:

```text
CancelOrders(action):
    process each requested CLOID in order

    if matching open Order exists:
        close it as canceled
        if it is a bracket parent:
            cancel its open children internally
        append success
    else:
        append item error

    return the exact Hyperliquid batch response
```

Ordinary sibling Orders do not cascade.

Internally canceled bracket children do not create extra response items.

**KEY DECISION:** Cancellation follows exact Hyperliquid batch shape and ordering.

## 8. Change Matching

Functions:

- `(*Simulator).onBBO`
- `(*Simulator).match`
- `(*Simulator).matchAdded`
- `(*Simulator).fill`
- `(*Simulator).armChildren`
- `(*Simulator).cancelChildren`

```text
BBO subscription callback:
    read latest BBO
    store BBO timestamp as Simulator current time
    compare BBO with open Orders

    for each matched Order:
        generate Fill
        update Order
```

That is the complete matching responsibility.

**KEY DECISION:** Simulator matching is BBO comparison, Fill generation, and Order update.

## 9. Change Time

```text
Simulator current time = latest subscribed BBO.TimestampMS
```

The BBO subscription callback stores the timestamp before matching.

Simulator-created submission, rejection, cancellation, Fill, and Order-update timestamps use that stored BBO timestamp.

Do not inject a Clock.

Do not accept caller-supplied current time.

**KEY DECISION:** BBO timestamp is Simulator time.

## 10. Change Prices

Store separately:

```text
LimitPrice
TriggerPrice
FillPrice
```

Delete `orderPrice`.

Each response or matcher reads the exact required field.

**KEY DECISION:** Trigger threshold is not the submitted limit price or execution price.

## 11. Change Query Responses

Functions:

- `(*Simulator).OpenOrders`
- `(*Simulator).Fills`
- `(*Simulator).OrderStatus`
- `(*Simulator).AccountState`

Simulator returns exact Hyperliquid-shaped JSON.

Simulator never returns internal structs.

Fill retention:

```text
memory:   latest 10,000 Fills
response: maximum 2,000 Fills
```

**KEY DECISION:** Production and live Exchange behavior determine tests. Tests never determine production behavior.

## 12. Supported Simulator Behavior

Support:

- `GTC`
- `IOC`
- `na`
- `normalTpsl`
- non-market trigger Orders

Fail loudly:

- `ALO`
- `positionTpsl`
- trigger-market Orders

Add this exact capability list near the top of `simulator.go`.

**KEY DECISION:** Unsupported behavior is documented now, not rediscovered later.

## 13. Change `(*Simulator).Stop`

Pseudocode:

```text
Stop():
    stop MarketData subscription
    mark stopped
    repeated completed Stop returns nil
```

Stop performs no persistence.

## 14. Simulator Boundary

Simulator accepts:

```text
Init(Config)
subscribed BBO updates
PlaceOrders(Hyperliquid action)
CancelOrders(Hyperliquid action)
OpenOrders query
Fills query
OrderStatus query
AccountState query
Stop()
```

Simulator returns:

```text
error for lifecycle or unsupported processing failure
exact Hyperliquid JSON for Exchange operations and queries
```

Simulator internally owns:

```text
starting balance
current BBO and BBO timestamp
position
open and closed Orders
Fills
OID and TID allocation
matching
fees and slippage
```

Simulator does not know or decide what Account, Executor, Controller, Backtest, or Live does.

**KEY DECISION:** Simulator is a called in-memory Exchange simulator, not an application orchestrator.

## 15. File Changed During Implementation

Production:

- `internal/simulator/simulator.go`

Tests:

- Simulator parity tests

## 16. Proof

Required:

```text
42-function inventory remains complete
official Hyperliquid JSON fixtures pass
duplicate CLOID is not rejected
batch placement and cancellation preserve result order
unsupported order types fail loudly
subscription BBO timestamp becomes Simulator current time
matched BBO generates Fill and updates Order
Simulator uses no Clock
Simulator performs no persistence
full tests with -tags noasm
full vet with -tags noasm
```

## 17. Function Review Status

Keep:

- `(*Simulator).validateAccount`
- `(*Simulator).executableQuantity`
- `crosses`
- `newComparisonKey`
- `compareComparisonKeys`
- `closePnL`
- `fillDirection`
- `sideCode`
- `sameSign`
- `validCLOID`

Modify:

- `(*Simulator).Init`
- `(*Simulator).PlaceOrders`
- `(*Simulator).CancelOrders`
- `(*Simulator).Stop`
- `(*Simulator).onBBO`
- `(*Simulator).OpenOrders`
- `(*Simulator).Fills`
- `(*Simulator).OrderStatus`
- `(*Simulator).AccountState`
- `(*Simulator).newOrder`
- `(*Simulator).match`
- `(*Simulator).fill`

Deleted after caller proof:

- `(*Simulator).SetFillFeeAvailableForTest`

Every other function is currently used.

The proposed Keep, Modify, or removal decisions in Topics 2 through 16 require individual review before implementation.

No other production file changes.
