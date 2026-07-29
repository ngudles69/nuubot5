# Server

Status: Initial embedded WebServer implemented; managers and control surface pending.
Purpose: Provide one Nuubot binary, one Server process, and isolated Bot execution processes.

## Current Development Commands

Development currently uses separate executables:

```text
nuubot-server
nuubot-cli
nuubot-live
nuubot-sweep
nuubot-backtest
nuubot-stest-report
```

These keep current development, testing, profiling, and manual review boundaries simple.

They remain until the combined command is implemented and proven.

## Target Binary

Nuubot eventually compiles into one executable named `nuubot`.

Commands select behavior:

```text
nuubot serve
nuubot bot create
nuubot bot run --bot 4
nuubot backtest run --bot 1
nuubot sweep run --sweep 6
```

One source tree produces separate binaries for each operating system and architecture.

Canonical builds use `CGO_ENABLED=0` and `-tags noasm`.

## Server Process

`nuubot serve` runs one Server process:

```text
nuubot serve
├── WebServer
├── API
├── BotManager
└── SweepManager
```

WebServer, API, BotManager, and SweepManager live inside the Server process.

## Implemented Initial WebServer

`nuubot-server` now runs one standard-library HTTP service on
`127.0.0.1:9898`.

Implemented routes:

```text
GET /          embedded Nuubot Server page
GET /health    JSON health response
GET /assets/   embedded static assets
```

The initial Server contains no BotManager, SweepManager, datastore, HTMX,
ECharts, or control API.

The command opens the Server log, creates one signal context, executes Server,
and reports one terminal error.

`server.Execute`:

```text
validate Server options
create WebServer
run WebServer
log execute completed
```

`webserver.Create`:

```text
validate WebServer inputs
parse home template
create WebServer
```

`WebServer.Run`:

```text
open HTTP listener
start HTTP server
wait for WebServer stop
stop HTTP server
log run completed
```

Cancellation performs bounded graceful HTTP shutdown.

Templates and assets compile into the executable through Go `embed`.

After successful listener binding, WebServer prints and logs:

```text
WebServer started on http://127.0.0.1:9898
```

After graceful shutdown, WebServer prints and logs:

```text
WebServer stopped
```

These are the only approved Server lifecycle lines written directly to stdout.

A console write failure produces a Server-log warning and does not stop HTTP.

## Child Processes

BotManager and SweepManager launch the same executable with different subcommands:

```text
nuubot serve
├── child process: nuubot bot run --bot 4
├── child process: nuubot backtest run --bot 1
└── child process: nuubot backtest run --bot 2
```

Each Runner and BtBot has its own PID, Go runtime, memory, Clock, Controller, logs, database connections, and exit status.

The operating system schedules child processes independently across available CPUs.

One child crash does not crash Server or another child.

## Standalone Execution

Runner and BtBot never require Server.

The same binary can run directly on a backtest machine or VPS:

```text
nuubot backtest run --bot 1
nuubot bot run --bot 4
```

## Concurrency

Functions remain normal blocking functions with explicit context and error returns.

The owning caller decides whether to call directly or through `go`.

The caller owns cancellation, error collection, and shutdown waiting.

There are no separate synchronous and asynchronous implementations of the same operation.

## Ownership

Server owns WebServer, API, BotManager, SweepManager, process supervision, and graceful service shutdown.

Managers own child launch, PID tracking, cancellation, exit handling, and restart policy.

Server and Managers never own Controller, BotCycle, Executor, Account, Ledger, or trading policy.

## Persistence

The central `nuubot.db` stores Bot and Sweep configuration, commands, acknowledgements, process generations, lifecycle status, and health.

Each globally unique Bot ID owns one execution database:

```text
workspace/db/bots/bot_<BotID>.db
```

Backtest replaces that database only after successful terminal publication.

Live retains and reopens the same database across process recovery. Live never clears execution evidence.

Command and supervision transactions remain short. They never span Controller, Recon, Stop, or operating-system process waits.

## Deployment

Build the binary for the target operating system.

Copy the Windows binary to the Windows backtest machine.

Copy the Linux binary to the Linux VPS.

One deployment artifact supplies Server, CLI, Runner, BtBot, Sweep, and reporting commands for that platform.
