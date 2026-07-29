# Server Design

Date: 2026-07-28

Status: Current approved design. Implementation pending.

Purpose: Define the Nuubot5 Server shell, managers, WebServer, chart surface, and embedded frontend.

## Product Boundary

Server is one application shell.

It owns:

```text
configuration
datastore
WebServer
BotManager
SweepManager
process supervision
graceful shutdown
```

Server does not own:

```text
Controller
Signaler calculations
Risk
BotCycle execution
Executor policy
Account
Ledger
Venue behavior
backtest execution
live execution
```

Runner, BtBot, and BtSweep remain independently executable.

## Process Shape

```text
nuubot-server
`-- Server
    |-- Datastore
    |-- BotManager
    |-- SweepManager
    `-- WebServer
        |-- Bot pages and API
        |-- Sweep pages and API
        |-- Charts
        `-- Admin
```

Server launches and supervises child processes.

Managers own child lifecycle policy.

Server never constructs Controller or trading objects.

## File Structure

```text
cmd/
`-- nuubot-server/
    `-- main.go

internal/
`-- server/
    |-- execute.go
    |-- server.go
    |
    |-- botmanager/
    |   |-- manager.go
    |   |-- commands.go
    |   |-- queries.go
    |   `-- process.go
    |
    |-- sweepmanager/
    |   |-- manager.go
    |   |-- commands.go
    |   |-- queries.go
    |   `-- process.go
    |
    `-- webserver/
        |-- webserver.go
        |-- routes.go
        |-- middleware.go
        |-- pages.go
        |-- responses.go
        |-- bots.go
        |-- sweeps.go
        |-- charts.go
        `-- admin.go

web/
|-- embed.go
|-- templates/
|   |-- layout.html
|   |-- bots.html
|   |-- bot.html
|   |-- sweeps.html
|   |-- sweep.html
|   `-- admin.html
|
`-- assets/
    |-- css/
    |   `-- app.css
    |
    |-- js/
    |   |-- app.js
    |   `-- chart/
    |       |-- chart.js
    |       |-- theme.js
    |       `-- layers/
    |           |-- base.js
    |           |-- signaler.js
    |           |-- indicators.js
    |           |-- cycles.js
    |           |-- trades.js
    |           |-- orders.js
    |           `-- fills.js
    |
    |-- images/
    |   |-- logo.svg
    |   |-- favicon.svg
    |   `-- icons/
    |
    `-- vendor/
        |-- echarts.min.js
        `-- htmx.min.js
```

Do not create speculative packages.

Start with the listed files.

Split a file only after its independent ownership or testing boundary is proven.

## Command Entrypoint

`cmd/nuubot-server/main.go` is the operating-system adapter.

It owns:

```text
argument parsing
signal context
calling server.Execute
terminal error reporting
exit status
```

It does not own:

```text
HTTP routes
templates
database queries
manager construction details
chart data
business policy
```

Target shape:

```go
func main() {
    ctx, stop := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,
    )
    defer stop()

    options, err := parseArguments(os.Args[1:])
    if err == nil {
        err = server.Execute(ctx, options)
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

Exact arguments remain an implementation decision.

## Server Shell

`internal/server/execute.go` owns the complete application lifecycle.

Expected flow:

```text
validate options
load configuration
open central datastore
construct BotManager
construct SweepManager
construct WebServer
start supervision
serve HTTP
wait for cancellation or service failure
stop accepting requests
stop Server-owned background work
close datastore
return terminal error
```

`internal/server/server.go` owns assembled dependencies and lifecycle state.

It contains no HTTP handler or trading logic.

## BotManager

BotManager is the responder for Bot requests.

It owns:

```text
create
view
clone
start
pause
stop
delete configured Bot
status
health
results
child process policy
```

BotManager reads durable central and per-Bot databases.

It never reads live object pointers.

Delete is allowed only for configured Bots.

Terminal execution evidence is not deleted through configured-Bot deletion.

Process health is stored durably.

Web pages read stored health.

HTTP handlers do not synchronously probe a child process.

## SweepManager

SweepManager is the responder for Sweep requests.

It owns:

```text
create
view
clone
start
pause
stop
delete configured Sweep
status
health
results
child process policy
```

SweepManager reads durable central and result data.

It never expands permutations inside HTTP handlers.

BtSweep owns expansion, bounded workers, cancellation, and result aggregation.

Future heatmaps and Monte Carlo pages read published artifacts.

Long analytics never execute inside an HTTP request.

## WebServer

WebServer owns the HTTP boundary.

Use Go standard packages first:

```text
net/http
html/template
encoding/json
embed
```

It owns:

```text
route registration
request parsing
boundary validation
HTML rendering
HTMX fragments
JSON responses
static assets
HTTP error mapping
graceful HTTP shutdown
```

It does not own Bot or Sweep lifecycle policy.

Mutating handlers call BotManager or SweepManager.

API routes remain thin.

## WebServer Files

### `webserver.go`

Owns:

```text
http.Server
dependencies
start
shutdown
```

### `routes.go`

Registers every page, API, asset, and Admin route.

Use one standard `http.ServeMux`.

Do not add an external router without a proven missing capability.

### `middleware.go`

Contains only shared HTTP boundary behavior.

Potential first requirements:

```text
request logging
panic recovery
security headers
request size limits
```

Do not add middleware chains before requirements exist.

### `pages.go`

Owns shared layout rendering and HTMX full-page or fragment selection.

### `responses.go`

Owns consistent JSON, fragment, and HTTP error responses.

### `bots.go`

Owns Bot page handlers and thin Bot API handlers.

### `sweeps.go`

Owns Sweep page handlers and thin Sweep API handlers.

### `charts.go`

Owns all chart read orchestration initially.

Keep it in one file first.

Use private functions for separate data pulls.

Do not create a chart Go package before the file becomes difficult to navigate or test.

### `admin.go`

Admin is deliberately off the main development path.

It contains stable operational pages and endpoints.

Initial concerns may include:

```text
Server health
process status
storage status
configuration display
log links
```

Do not create an AdminManager without real policy.

Once Admin works, leave it stable.

## Charts

The Nuubot3 report is the minimum visual and analytical baseline.

This is not a future feature list.

### Minimum Summary Surface

```text
Info
Performance
PnL
Drawdown
Wins
Loss
Long
Short
```

### Minimum Chart Surface

```text
candles including warmup
balance
equity
volume
indicators
Signaler events
BotCycle ranges
Trade entry and exit
Orders
Fills
```

### Minimum Interaction

```text
whole-Bot focus
one-BotCycle focus
timeframe switching
chart zoom
layer visibility
element visibility
lazy optional-layer loading
dark theme
light theme
```

Trade-level focus is not required.

## `charts.go`

`charts.go` manages every chart data pull.

Initial private functions:

```go
resolveChartRange(...)
loadCandles(...)
loadAccountSeries(...)
loadIndicators(...)
loadSignals(...)
loadCycles(...)
loadTrades(...)
loadOrders(...)
loadFills(...)
```

These functions return renderer-neutral Go values.

They do not return ECharts options.

`charts.go` does not calculate trading results or indicators.

It reads published evidence and admitted market data.

### Data Sources

```text
Candles          admitted OHLCV source
Volume           admitted OHLCV source
Balance          stored account telemetry
Equity           stored account telemetry
Indicators       stored Signaler chart samples
Signaler events  stored Signal decisions or chart samples
BotCycles        stored BotCycle results
Trades           stored Trade evidence
Orders           stored Order evidence
Fills            stored Fill evidence
```

Chart data is database-first.

Chart requests never inspect running Bot objects.

## Chart Range

Chart range has two focus modes:

```text
whole_bot
botcycle
```

### Whole Bot

Whole-Bot display begins at the admitted warmup start.

It ends at the Bot execution end.

```text
warmup start -------- Bot start ---------------- Bot end
      candles             candles
      indicators          indicators
                          cycles and transactions
```

Warmup candles are visible.

Stored indicator values cover the displayed warmup range when series storage is enabled.

### One BotCycle

BotCycle focus uses stored cycle start and end boundaries.

It displays the selected cycle's bounded chart evidence.

The indicators shown inside that range were calculated using their required historical warmup.

The Server does not recalculate them.

## Chart Storage Dependency

Main-worktree publication must provide:

```text
BotCycle start_ms
BotCycle end_ms
BotCycle duration_ms
mandatory Signal decisions
optional renderer-neutral Signaler chart samples
warmup Signaler series when configured
```

Signaler chart storage modes remain:

```text
none
events
series
```

Mode behavior:

```text
none    Keep mandatory Signal decisions only.
events  Store event Packages.
series  Store complete chart series, including warmup.
```

Execution Packages and chart samples are distinct outputs.

```text
Controller Packages
`-- Executable Bot range only

Signaler chart samples
`-- Warmup and executable Bot range
```

Both outputs must come from the existing Signaler calculation.

Do not run a second indicator calculation.

Do not calculate historical indicators inside Server requests.

Missing series display as not recorded.

## Lazy Loading

Initial Bot page loads:

```text
summary cards
BotCycle index
chart shell
base chart
```

Base chart contains:

```text
candles
balance
equity
volume
```

Optional layers fetch only when visible:

```text
indicators
Signaler events
BotCycles
Trades
Orders
Fills
```

Enabling an unloaded layer fetches it once for the current focus.

Disabling a layer hides cached browser data.

Changing focus invalidates layer cache entries for the old focus.

Do not add a Server cache initially.

## Chart HTTP Surface

Initial read routes:

```text
GET /api/bots/{botID}/chart/base
GET /api/bots/{botID}/chart/indicators
GET /api/bots/{botID}/chart/signals
GET /api/bots/{botID}/chart/cycles
GET /api/bots/{botID}/chart/trades
GET /api/bots/{botID}/chart/orders
GET /api/bots/{botID}/chart/fills
```

Requests may accept:

```text
cycle
timeframe
from_ms
to_ms
```

The route validates:

```text
Bot existence
cycle existence
timeframe
requested bounds
result ownership
```

Responses contain renderer-neutral facts.

Avoid one route per indicator.

## Frontend Ownership

HTMX owns:

```text
page navigation
forms
Bot controls
Sweep controls
tables
detail panels
cycle selection
status polling
fragment replacement
```

Native JavaScript owns:

```text
ECharts lifecycle
series composition
chart zoom
layer visibility
element visibility
chart data cache
theme synchronization
```

HTMX does not manage ECharts series.

One explicit initialization hook recreates charts after relevant fragment replacement.

## Frontend Chart Modules

Backend Go remains in one `charts.go` file initially.

Frontend chart layers remain independent modules.

This modularity is required by layer and element toggles.

```text
chart.js
|-- base.js
|-- signaler.js
|-- indicators.js
|-- cycles.js
|-- trades.js
|-- orders.js
`-- fills.js
```

`chart.js` owns:

```text
ECharts instance
active focus
timeframe
visible layers
loaded-layer cache
data zoom
resize
theme
```

Each layer module owns only its data-to-series conversion and styles.

Layer modules do not fetch data.

Layer modules do not create ECharts instances.

Stable series identifiers include:

```text
base:candles
base:balance
base:equity
base:volume
signaler:start
signaler:stop
indicators:fast_ma
cycles:range
trades:entry
trades:exit
orders:submitted
fills:entry
fills:exit
```

Use native ECharts legend selection for element visibility.

## Embedded Web Assets

The root `web` package embeds the frontend.

Target:

```go
//go:embed templates assets
var Files embed.FS
```

Embedded assets include:

```text
templates
CSS
JavaScript
images
SVG
icons
fonts when required
ECharts
HTMX
licenses
```

Deployment remains one executable.

Do not add Node, npm, Vite, React, or Vue.

Prefer SVG and CSS before raster images.

## Theme

Use CSS custom properties and one document theme attribute.

```html
<html data-theme="dark">
```

Initial themes:

```text
dark
light
```

ECharts reads the same semantic colors.

Browser-local preferences use `localStorage`.

Do not add Server-side preference storage initially.

## Control API

Page controls and CLI use the same manager-backed operations.

Initial Bot routes:

```text
POST   /api/bots
GET    /api/bots/{botID}
POST   /api/bots/{botID}/clone
POST   /api/bots/{botID}/start
POST   /api/bots/{botID}/pause
POST   /api/bots/{botID}/stop
DELETE /api/bots/{botID}
```

Initial Sweep routes:

```text
POST   /api/sweeps
GET    /api/sweeps/{sweepID}
POST   /api/sweeps/{sweepID}/clone
POST   /api/sweeps/{sweepID}/start
POST   /api/sweeps/{sweepID}/pause
POST   /api/sweeps/{sweepID}/stop
DELETE /api/sweeps/{sweepID}
```

HTTP handlers parse inputs and map errors.

Managers validate lifecycle transitions and policy.

## Initial Runtime Updates

Browser polling is sufficient.

Do not add WebSockets or Server-Sent Events initially.

Running state comes from durable manager health records.

Polling fragments must not reload the chart unnecessarily.

## Implementation Order

```text
1. Main publishes BotCycle boundaries.
2. Main publishes configured Signaler chart evidence, including warmup series.
3. Main proves trading results remain unchanged.
4. Server shell and lifecycle.
5. BotManager and SweepManager read paths.
6. Embedded WebServer and static page shell.
7. Nuubot3 summary cards and base chart.
8. Lazy chart layers in charts.go.
9. Whole-Bot and BotCycle focus.
10. Bot controls.
11. Sweep controls.
12. Stable Admin page.
```

Keep main and Server commits independently reviewable.

## Required Proof

### Server Lifecycle

```text
invalid options fail before serving
startup failure closes opened resources
context cancellation stops HTTP cleanly
terminal service errors reach main
datastore closes after dependent services
```

### Managers

```text
configured-only deletion
valid lifecycle transitions
invalid lifecycle rejection
durable status and health reads
child generation tracking
no live object dependency
```

### WebServer

```text
page routes
API routes
HTMX fragment responses
JSON error mapping
embedded templates
embedded CSS
embedded JavaScript
embedded images
graceful HTTP shutdown
```

### Charts

```text
warmup candles included in whole-Bot range
warmup indicators included when recorded
one-BotCycle bounds use stored boundaries
initial page fetches base only
one enabled layer causes one layer request
hidden layers do not fetch
focus changes invalidate old layer cache
missing indicators report not recorded
no Server-side indicator recalculation
dark and light rendering remain readable
```

Canonical Go verification:

```text
go test -tags noasm ./...
go vet -tags noasm ./...
git diff --check
```

## Deferred

These are outside the minimum release:

```text
trade-level chart focus
Server chart cache
WebSockets
Server-Sent Events
Server-side UI preferences
frontend framework
frontend build pipeline
generic chart plugin system
expanded Admin subsystem
heatmaps
Monte Carlo
```

The minimum report and chart surface is not deferred.

## Completion Definition

Server is complete when it controls Bots and Sweeps and reproduces the Nuubot3 minimum report surface from durable evidence.

It must remain one clean application shell with a thin HTTP boundary and independently executable trading processes.
