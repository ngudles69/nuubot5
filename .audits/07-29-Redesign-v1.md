


file: cmd/nuubot-backtest
    parse arguments, including optional -pp
    runner = CreateRunner(arguments)
    runner.Execute(options)

file: backtest/main.go

    backtest.Execute
        validate options
        open bot log
        start whole-run profile
        create Backtest Run
        Init
        Start
        Loop
        Stop
        collect result
        write report
        stop profile
        log completion

    Run
        load stored Config
        create replay reader
        create replay Clock
        create MarketData
        create Simulator Venue
        open Stores
        create Strategy from Config.Strategy
        pass infrastructure into Strategy

    Start
        start Clock
        start Strategy

    Loop
        read next historical tick
        update MarketData
        advance Clock
        call Strategy.Next when due
        exit when replay or Strategy ends

    Stop
        stop Strategy
        stop MarketData
        stop replay reader
        stop Clock
        close Stores

file: strategy/MacrossTrade.go

    OnInit
        config = get passed in config e.g. nuubot.config or nuubot
        acct1 = CreateAccount(config.??)
        acct2 = CreateAccount(config.??)
        signaler = CreateSignaler(config.signaler)
        risk = CreateRisk(config.risk)
        executor1 = CreateExecutor(config.executor[0], acct2)
        executor2 = CreateExecutor(config.executor[1], acct2)

    OnBar(bar)
        signaler.ingest(bar)

    OnTick(tick)    <-- same as bbo

        read current market state??
        read Signal??
        assess Risk??
        executor1.OnTick(tick)
        executor2.OnTick(tick)
        return continue or exit

    OnDeInit
        stop Executors
        return final Strategy result


=========================================
Everything below this line is standard infra
=========================================

  Store  ?  part of nuubot?

  Clock  ?  part of Runtime - added to nuubot?

  WebSocket  ?  part of Runtime added to nuubot?
