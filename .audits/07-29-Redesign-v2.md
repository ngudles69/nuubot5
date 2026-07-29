# Redesign v2

`CreateXXX` creates and initializes the object.

`Start` remains separate.

## Backtest

file: cmd/nuubot-backtest/main.go

    main
        options = parseArguments(os.Args)
        backtest.Execute(context, options)

file: internal/backtest/execute.go

    backtest.Execute
        log = OpenBotLog(options)
        profile = CreateProfile(options)
        config = LoadConfig(options)

        replay = CreateReplay(config.Replay)
        clock = CreateReplayClock(config.Clock)
        market = CreateMarketData(config.Markets)
        store = CreateStore(config.Store)
        venue = CreateSimulator(config.Venue, store)

        nuubot = CreateNuubot(config, clock, market, store, venue)

        strategy = CreateStrategy(config.Strategy, nuubot)

        profile.Start()
        store.Start()
        venue.Start()
        market.Start()
        strategy.Start()
        clock.Start()

        loop
            tick = replay.Next()
            bars = market.Ingest(tick)
            venue.Ingest(tick)
            due = clock.Advance(tick.Time)

            for each completed bar
                strategy.OnBar(bar)

            if due
                status = strategy.OnTick(tick)

            if replay is complete or status is exit
                break

        clock.Stop()
        strategy.Stop()
        market.Stop()
        venue.Stop()
        replay.Stop()
        store.Stop()

        result = strategy.Result()
        WriteReport(result)
        profile.Stop()

## Strategy

file: internal/strategy/macrosstrade.go

    CreateMacrossTrade(config, nuubot)
        acct1 = CreateAccount(config.Accounts[0], nuubot)
        acct2 = CreateAccount(config.Accounts[1], nuubot)
        signaler = CreateSignaler(config.Signaler, nuubot)
        risk = CreateRisk(config.Risk, nuubot)
        executor1 = CreateExecutor(config.Executors[0], acct1, nuubot)
        executor2 = CreateExecutor(config.Executors[1], acct2, nuubot)

        return MacrossTrade(
            config,
            acct1,
            acct2,
            signaler,
            risk,
            executor1,
            executor2,
        )

    MacrossTrade.Start
        acct1.Start()
        acct2.Start()
        executor1.Start()
        executor2.Start()

    MacrossTrade.OnBar(bar)
        signaler.Ingest(bar)

    MacrossTrade.OnTick(tick)
        signal = signaler.Signal()
        decision = risk.Assess(signal, acct1, acct2)
        result1 = executor1.Next(tick, decision)
        result2 = executor2.Next(tick, decision)

        if result1 is exit or result2 is exit
            return exit

        return continue

    MacrossTrade.Stop
        executor1.Stop()
        executor2.Stop()
        acct1.Stop()
        acct2.Stop()

## Live

file: cmd/nuubot-live/main.go

    main
        options = parseArguments(os.Args)
        live.Execute(context, options)

file: internal/live/execute.go

    live.Execute
        log = OpenBotLog(options)
        config = LoadConfig(options)

        clock = CreateLiveClock(config.Clock)
        market = CreateMarketData(config.Markets)
        store = CreateStore(config.Store)
        venue = CreateExchange(config.Venue, store)
        websocket = CreateWebSocket(config.Venue)

        nuubot = CreateNuubot(config, clock, market, store, venue)

        strategy = CreateStrategy(config.Strategy, nuubot)

        store.Start()
        venue.Start()
        market.Start()
        strategy.Start()
        websocket.Start()
        clock.Start()

        loop
            event = websocket.Next()

            if event is market tick
                bars = market.Ingest(event.Tick)
                venue.Ingest(event.Tick)

                for each completed bar
                    strategy.OnBar(bar)

            if clock.ControlDue()
                status = strategy.OnTick(market.Latest())

            if status is exit
                break

        clock.Stop()
        websocket.Stop()
        strategy.Stop()
        market.Stop()
        venue.Stop()
        store.Stop()

## Backtest versus Live

    Backtest
        input = Replay
        clock = ReplayClock
        venue = Simulator
        end = replay complete or Strategy exit
        output = result report

    Live
        input = WebSocket
        clock = LiveClock
        venue = Exchange
        end = external stop or Strategy exit

    Shared
        Config
        MarketData
        Store
        Nuubot
        Strategy
