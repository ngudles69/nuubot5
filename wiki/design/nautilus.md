# NautilusTrader Intent Comparison

Status: Current Venue comparison.
Covers: Nuubot5 Venue, Nuubot3 Exchange, and NautilusTrader Hyperliquid execution.
Purpose: Reuse proven external intent without copying another implementation.

## Authority

Nuubot5 source owns implemented behavior.

Nuubot3 and NautilusTrader are implementation references only.

Nuubot does not copy their code or architecture line for line.

## Function Comparison

```text
Intent                    Nuubot3                     NautilusTrader                         Nuubot5
------------------------  --------------------------  -------------------------------------  --------------------
Connect Venue             init                        _connect                               Venue.Connect
Connect simulator         init                        _connect                               Simulator.Connect
Disconnect Venue          close                       _disconnect                            Venue.Disconnect
Disconnect simulator      close                       _disconnect                            Simulator.Disconnect

Submit one Order          place_order                 _submit_order                          PlaceOrders([one])
Submit Order batch        place_orders                _submit_order_list                     PlaceOrders
Modify Order              none                        _modify_order / modify_order            none
Cancel one Order          cancel_order                _cancel_order                          CancelOrders([one])
Cancel Order batch        cancel_orders               _batch_cancel_orders                   CancelOrders
Cancel all Orders         none                        _cancel_all_orders                      future Account helper
Set leverage              none                        external Hyperliquid API                SetLeverage

Read Open Orders          get_open_orders             info_open_orders                       GetOpenOrders
Read Order History        get_user_order_history      info_historical_orders                 GetOrderHistory
Bulk Order reports        Order History               historical Order status reports        GetOrderHistory
Read exact Order status   get_order_status            info_order_status                      GetOrderStatus
Read Fill history         get_user_fills              info_user_fills                        GetFillHistory
Read Account state        get_account_state           info_clearinghouse_state               GetAccountState
Read Position reports     through Account state       request_position_status_reports        through Account state
Read Active Asset Data    none                        activeAssetCtx through DataClient       Meta
```

## Bulk Order Reports

NautilusTrader bulk Order reports are framework-normalized reconciliation
reports.

Hyperliquid supplies `historicalOrders`.

Nuubot uses `GetOrderHistory` for that intent.

Nuubot needs no second bulk-report Venue method.

## Recommendation

Keep the current Venue surface:

```text
Connect
PlaceOrders
CancelOrders
SetLeverage
GetOpenOrders
GetOrderHistory
GetFillHistory
GetOrderStatus
GetAccountState
Disconnect
```

Venue remains pass-through routing.

Future `CancelAllOrders` belongs in Account as convenience orchestration over
known Orders.

Add modify, events, or fee queries only when a production caller needs them.

## Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/exchange/exchange.py`
- Nuubot3 Simulator: `D:/rust/nuubot3/nuubot/exchange/simulator.py`
- NautilusTrader Hyperliquid execution:
  `https://github.com/nautechsystems/nautilus_trader/blob/develop/nautilus_trader/adapters/hyperliquid/execution.py`
- NautilusTrader Hyperliquid HTTP:
  `https://github.com/nautechsystems/nautilus_trader/blob/develop/crates/adapters/hyperliquid/src/http/client.rs`
- Hyperliquid Info:
  `https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint`
- Hyperliquid Exchange:
  `https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/exchange-endpoint`
