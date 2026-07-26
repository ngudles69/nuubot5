# Go Source Reviews

Change `User Review` to `YES` only after explicit user confirmation.
At review, record the exact review date and source file SHA-256.
Later hash comparison detects source changes after review.
Never infer review completion.

```text
Path                                                   User Review  Review Date  Reviewed SHA-256
-----------------------------------------------------  -----------  -----------  ----------------
cmd/nuubot-bt-bot/main.go                              NO
cmd/nuubot-bt-bot/main_test.go                         NO
cmd/nuubot-bt-sweep/main.go                            NO
cmd/nuubot-cli/main.go                                 NO
cmd/nuubot-report/main.go                              NO
cmd/nuubot-runner/main.go                              NO
cmd/nuubot-server/main.go                              NO
cmd/parity-probe/main.go                               NO
internal/account/account.go                            NO
internal/account/account_test.go                       NO
internal/account/doc.go                                NO
internal/bot/bot.go                                    NO
internal/botcycle/botcycle.go                          NO
internal/botcycle/botcycle_test.go                     NO
internal/botspec/botspec_test.go                       NO
internal/botspec/build.go                              NO
internal/botspec/config.go                             NO
internal/btbot/btbot.go                                NO
internal/btsweep/btsweep.go                            NO
internal/btsweep/btsweep_test.go                       NO
internal/cloid/cloid.go                                NO
internal/cloid/cloid_test.go                           NO
internal/config/config.go                              NO
internal/config/config_test.go                         NO
internal/config/credentials.go                         NO
internal/controller/controller.go                      NO
internal/controller/controller_test.go                 NO
internal/datastore/models.go                           NO
internal/datastore/sweep.go                            NO
internal/datastore/sweep_test.go                       NO
internal/executor/executor.go                          NO
internal/executor/grid.go                              NO
internal/executor/grid_test.go                         NO
internal/executor/observer.go                          NO
internal/executor/observer_test.go                     NO
internal/executor/trade.go                             NO
internal/executor/trade_test.go                        NO
internal/fill/doc.go                                   NO
internal/fill/fill.go                                  NO
internal/fill/fill_test.go                             NO
internal/hyperliquid/client.go                         NO
internal/hyperliquid/client_test.go                    NO
internal/hyperliquid/meta.go                           NO
internal/hyperliquid/meta_test.go                      NO
internal/hyperliquid/state.go                          NO
internal/ledger/doc.go                                 NO
internal/ledger/ledger.go                              NO
internal/ledger/ledger_test.go                         NO
internal/ledger/publish.go                             NO
internal/ledger/store.go                               NO
internal/market/market.go                              NO
internal/meta/doc.go                                   NO
internal/meta/meta.go                                  NO
internal/meta/meta_test.go                             NO
internal/ohlcv/ohlcv.go                                NO
internal/ohlcv/ohlcv_test.go                           NO
internal/order/doc.go                                  NO
internal/order/order.go                                NO
internal/order/order_test.go                           NO
internal/parity/info/clearinghouse.go                  NO
internal/parity/info/clearinghouse_test.go             NO
internal/parity/info/info.go                           NO
internal/parity/parity.go                              NO
internal/parity/parity_test.go                         NO
internal/replay/replay.go                              NO
internal/report/render.go                              NO
internal/report/report.go                              NO
internal/report/report_test.go                         NO
internal/resultpublisher/resultpublisher.go            NO
internal/resultpublisher/resultpublisher_test.go       NO
internal/risk/balanced.go                              NO
internal/risk/risk.go                                  NO
internal/risk/risk_test.go                             NO
internal/setup/setup.go                                NO
internal/signaler/macross.go                           NO
internal/signaler/rsi.go                               NO
internal/signaler/signaler.go                          NO
internal/signaler/signaler_test.go                     NO
internal/simulator/doc.go                              NO
internal/simulator/publish.go                          NO
internal/simulator/simulator.go                        NO
internal/simulator/simulator_test.go                   NO
internal/simulator/store.go                            NO
internal/telemetry/telemetry.go                        NO
internal/toolkit/clock/clock.go                        NO
internal/toolkit/clock/clock_test.go                   NO
internal/toolkit/clock/tickclock.go                    NO
internal/toolkit/clock/timer.go                        NO
internal/toolkit/clock/wallclock.go                    NO
internal/toolkit/logging/logging.go                    NO
internal/toolkit/logging/logging_test.go               NO
internal/trade/doc.go                                  NO
internal/trade/trade.go                                NO
internal/trade/trade_test.go                           NO
```
