# Redesign v3

## Core Model

This is a Config plus Strategy pairing model.

Config parameterizes bricks.

Strategy code defines the graph, interaction rules, and unrestricted strategy-specific logic.

    Runner
        owns runtime infrastructure
        creates and drives one Strategy

    Strategy
        implements the shared lifecycle
        composes only the bricks it needs

    Executor
        exclusively receives one Account

    Account
        hides Venue, Ledger, persistence, and reconciliation

## Runner

    Runner interface
        Init()
        Start()
        Run()
        Stop()

Commands and Server use the same lifecycle.

    runner = CreateRunner(context, options)
    runner.Init()
    runner.Start()
    runner.Run()
    runner.Stop()

Command runs one blocking Runner.

Server stores the Runner handle, runs it asynchronously, and calls `Stop`.

Runner contains no command parsing, process exit, or Server job management.

## Backtest Runner

    cmd/nuubot-backtest
        runner = backtest.CreateRunner(context, options)
        run standard Runner lifecycle

    BacktestRunner.Init
        profile = runharness.NewProfile(...)
        profile.Start()
        nuubot = setup.Setup(...)
        nuubot.Runtime = nuubot.App.Backtest
        reader.Init(...)
        nuubot.Clock = clock.Create(clock.Tick)
        nuubot.Clock.Init(...)
        nuubot.MarketData = market.CreateMarketData()
        strategy = CreateStrategy(nuubot.BotSpec, nuubot)
        clock.RegisterTimer(strategy.Run)

    BacktestRunner.Start
        clock.Start()
        strategy.Start()

    BacktestRunner.Run
        bbo = reader.Next()
        marketData.IngestBBO(bbo)
        clock.Advance(bbo.TimestampMS)
        timer calls strategy.Run()
        repeat until replay or Strategy exits

    BacktestRunner.Stop
        clock.Stop()
        reader.Stop()
        strategy.Stop()
        marketData.Stop()
        publish result
        profile.Stop()

## Live Runner

    cmd/nuubot-live
        runner = live.CreateRunner(context, options)
        run standard Runner lifecycle

    LiveRunner.Init
        nuubot = setup.Setup(...)
        nuubot.Runtime = nuubot.App.Live
        nuubot.Clock = clock.Create(clock.Wall)
        nuubot.Clock.Init(...)
        nuubot.MarketData = market.CreateMarketData()
        nuubot.Info = hyperliquid.NewInfo(...)
        nuubot.WebSocket = hyperliquid.NewWebSocket(...)
        strategy = CreateStrategy(nuubot.BotSpec, nuubot)
        clock.RegisterTimer(strategy.Run)

    LiveRunner.Start
        WebSocket.Start()
        strategy.Start()
        clock.Start()

    LiveRunner.Run
        WebSocket updates MarketData
        WallClock timer calls strategy.Run()
        supervise external stop and runtime failure

    LiveRunner.Stop
        clock.Stop()
        WebSocket.Stop()
        Info.Stop()
        strategy.Stop()
        marketData.Stop()

## Strategy

    Strategy interface
        Start()
        Run()
        Stop()
        Result()
        Telemetry()

Each Strategy has one exact typed Config.

Missing required Config fails Strategy creation.

Config stores Account names. Credentials remain infrastructure.

Config does not define a generic component graph.

Each Config plus Strategy pairing decides its deployment bindings.

Initial Nuubot6 implementation keeps Symbol and Network in Config TOML.

    Config
        Symbol is required
        Network is required

Runner keeps the current replay Symbol consistency check.

Future extension:

    Config supplies value
        value is fixed
        Bot cannot replace it

    Config omits value
        Bot must supply it

This future binding change is not part of the architecture implementation.

Example fixed pairing:

    GoldStrategyConfig
        Symbol = GOLD

    CreateGoldStrategy
        require resolved Symbol = GOLD
        reject BTC

## MacrossTrade Pairing

    MacrossTradeConfig
        GridAccountName
        HedgeAccountName
        Signaler1
        Signaler2
        Risks
        GridExecutor
        HedgeExecutor

    CreateMacrossTrade(config, nuubot)
        signaler1 = CreateSignaler(config.Signaler1)
        signaler2 = CreateSignaler(config.Signaler2)
        risks = CreateRisks(config.Risks)
        return MacrossTrade(config, nuubot, signaler1, signaler2, risks)

MacrossTrade creates Account and Executor pairs when it admits an execution cycle.

    gridAccount = CreateAccount(config.GridAccountName)
    hedgeAccount = CreateAccount(config.HedgeAccountName)

    gridExecutor = CreateGridExecutor(config.GridExecutor, gridAccount)
    hedgeExecutor = CreateHedgeExecutor(config.HedgeExecutor, hedgeAccount)

MacrossTrade assigns roles and controls how all bricks react.

MacrossTrade may add any custom logic.

Each Executor stops its exclusive Account.

## Observer Pairing

    ObserverConfig
        Market

    Observer
        subscribes to MarketData
        prints BBO

Observer needs no Account, Executor, Risk, or Signaler.

## Current Code Disposition

    KEEP
        replay.Reader
        TickClock and WallClock
        MarketData
        Info
        profiling, control Store, and result publication
        Account, Venue, Ledger, and Simulator internals
        Grid and Trade Executor policy
        Signaler calculations
        Risk implementations
        credentials loader
        datastore Bot record for the first implementation

    MODIFY
        Controller becomes concrete Strategies
        BotCycle coordination moves inside MacrossTrade
        Strategy creates Accounts and assigns them to Executors
        BotSpec becomes exact Config plus Strategy pairing
        Signaler creation separates calculation from replay-only loading
        Result and telemetry become Strategy-shaped
        Server owns active Runner handles
        Observer Executor becomes Observer Strategy

    TWEAK
        backtest.Run and live.Run implement Runner
        Loop becomes Run
        add Backtest and Live Runner factories
        timer calls Strategy.Run instead of Controller.Run
        Executor creation accepts one Account
        remove Account creation from Executors
        setup.Setup returns the paired Strategy Config
        Nuubot stores Strategy Config instead of generic BotSpec
        Runtime Controller interval becomes Strategy interval
        commands call Runner instead of package Execute
        WebServer routes call Server Runner operations
        sweep parameter paths follow Strategy Config

    DELETE
        generic Controller architecture
        public BotCycle architecture layer
        duplicated Backtest and Live lifecycle driving
        Observer Executor kind after Observer Strategy exists

    SEPARATE FUNCTIONAL GAPS
        Hyperliquid WebSocket Start is not implemented
        Live WebSocket does not feed MarketData
        credentials are not connected to Account
        MarketData has BBO but no live Bar pipeline

These functional gaps are not silently included in the architecture change.

## Implementation Order

    1. Standardize Runner lifecycle.
    2. Add Strategy contracts and Config factory.
    3. Move Controller logic into MacrossTrade.
    4. Move BotCycle coordination into MacrossTrade.
    5. Move Account creation into Strategy.
    6. Reshape Strategy result, telemetry, and reports.
    7. Add Observer Strategy.
    8. Add Server Runner ownership.
    9. Delete Controller and BotCycle.
    10. Prove Nuubot5 Backtest parity.

## Proof

    command can run Backtest
    command can construct Live Runner
    Server can start and stop Backtest Runner
    Server can own and stop Live Runner
    MacrossTrade preserves current behavior
    Account remains opaque
    Nuubot5 and new Backtest results match

Functional Live proof requires the separate WebSocket and credentials work.
