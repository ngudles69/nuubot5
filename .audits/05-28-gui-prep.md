# GUI Preparation Audit

Date: 2026-07-28

Scope: Prepare Nuubot5 data and Server boundaries for one combined control and analysis GUI.

This file contains all currently identified changes.

No implementation is included.

## Assessment

Nuubot3's generated report is the correct visual starting point.

Reuse its layout, cards, ECharts presentation, theme, and report density.

Do not copy its generated HTML or embedded Python JavaScript wholesale.

The storage and publication design is disciplined.

The browser renderer is monolithic and should be separated into small native JavaScript modules.

Nuubot5 needs main-worktree data changes before the Server can reproduce the report correctly.

## Accepted Product Decisions

- The GUI combines control and analysis.
- The GUI remains simple, themeable, and configurable.
- ECharts remains the chart engine.
- HTMX handles pages, forms, tables, controls, and polling.
- Native JavaScript owns ECharts state and layer loading.
- The initial chart loads only base data.
- Optional layers load only when visible.
- Chart focus has two levels: whole Bot and one BotCycle.
- Trade-level focus is not required.
- Layers and individual layer elements can be shown or hidden.
- Signal decisions are always stored.
- Additional Signaler output storage is configurable.
- Historical indicators are never recalculated inside HTTP requests.
- BotManager and SweepManager read durable data, not live object pointers.
- Admin pages remain lower priority.

## Current Confirmed State

### BotCycle

`internal/botcycle.Result` already contains:

```go
CycleNumber int
StartMS     uint64
EndMS       uint64
DurationMS  uint64
```

BotCycle already calculates and returns these values.

`internal/resultpublisher` currently discards them from `botcycle_result`.

The current table stores only:

```sql
cycle_number INTEGER PRIMARY KEY
```

This is a publication gap, not a BotCycle calculation gap.

### Signaler

Signaler already calculates one immutable Package per admitted signal bar.

Each Package contains:

```text
symbol
timestamp_ms
action
regime
risk_score
custom fields
```

Macross custom fields currently include:

```text
bar_start_ms
signal_price
fast_ma
slow_ma
regime_ma
```

RSI custom fields currently include:

```text
bar_start_ms
signal_price
rsi
volume_ratio
oversold
overbought
```

Controller currently retains only timestamp and action in its terminal Signal decisions.

`signal_decision` currently stores:

```sql
sequence
timestamp_ms
action
```

The remaining Package data is lost during result publication.

### Server

`cmd/nuubot-server/main.go` currently prints `Under Construction.`.

The approved Server design already owns:

```text
WebServer
API
BotManager
SweepManager
process supervision
graceful shutdown
```

Managers own child launch, PID tracking, cancellation, exit handling, and restart policy.

No separate ProcessManager should be introduced before shared process code is proven necessary.

## Required Main-Worktree Changes

These changes belong in the main `nuubot5` worktree.

They must land before cycle-focused GUI work.

### 1. Publish BotCycle Boundaries

Extend `botcycle_result`:

```sql
CREATE TABLE botcycle_result (
    cycle_number INTEGER PRIMARY KEY,
    start_ms INTEGER NOT NULL,
    end_ms INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL
);
```

Publish the existing `botcycle.Result` values without recalculation.

Required proof:

- Published boundaries equal the terminal BotCycle result.
- `duration_ms` equals the existing result value.
- Observer-only and no-trade cycles retain valid boundaries.
- Existing Executor foreign keys still resolve.
- Result publication remains atomic.

Likely owners:

```text
internal/resultpublisher/resultpublisher.go
internal/resultpublisher/resultpublisher_test.go
wiki/design/packages/resultpublisher.md
```

Do not add boundary calculations to ResultPublisher.

BotCycle remains the canonical lifecycle owner.

### 2. Preserve Canonical Signal Decisions

Keep `signal_decision` mandatory.

It remains the minimum durable record of Controller-observed Signaler actions.

Chart-storage settings must never disable it.

Required proof:

- Every newly observed Package still creates one ordered Signal decision.
- Disabling chart output does not remove Signal decisions.
- Controller behavior remains independent from GUI settings.

### 3. Add Configurable Signaler Chart Storage

Add one validated storage mode:

```text
none
events
series
```

Mode semantics:

```text
none    Store no additional Signaler chart payload.
events  Store full Packages only when the effective action changes.
series  Store every admitted full Package.
```

`signal_decision` remains stored in every mode.

The setting controls publication only.

It must not change Signaler calculation or Controller decisions.

Use one application-runtime setting per execution model unless a proven per-Bot requirement appears.

Suggested configuration shape:

```toml
[live]
signaler_storage = "events"

[backtest]
signaler_storage = "series"
```

Exact naming must follow the owning configuration design review.

Reject unknown modes during configuration loading.

### 4. Publish Renderer-Neutral Signaler Samples

Add one durable table for optional full Package payloads.

Minimal proposed shape:

```sql
CREATE TABLE signaler_sample (
    sequence INTEGER PRIMARY KEY,
    timestamp_ms INTEGER NOT NULL,
    payload_json TEXT NOT NULL
);
```

`payload_json` contains the Package's existing flat JSON representation.

Do not create indicator-specific SQL columns.

Different Signaler implementations own different custom fields.

Do not store ECharts series or renderer configuration.

Required proof:

- `none` publishes zero samples.
- `events` publishes only action transitions.
- `series` publishes every Controller-observed Package.
- Payload JSON preserves standard and custom Package fields.
- Samples remain chronological and uniquely sequenced.
- Invalid storage modes fail configuration loading.
- Existing trading results remain identical across storage modes.

Likely owners:

```text
internal/config/config.go
internal/config/config_test.go
internal/controller/controller.go
internal/controller/controller_test.go
internal/resultpublisher/resultpublisher.go
internal/resultpublisher/resultpublisher_test.go
internal/signaler/signaler.go
wiki/design/packages/signaler.md
wiki/design/packages/resultpublisher.md
```

The implementation must avoid a second Signaler calculation path.

The published payload must come from the Package already read by Controller.

### 5. Retain Base Chart Evidence

The base chart requires:

```text
OHLCV
balance
equity
volume
```

OHLCV remains sourced from admitted market data.

Balance and equity remain sourced from stored telemetry or equivalent published observations.

Do not copy the complete historical market dataset into each Bot database.

The Server must query bounded market ranges by symbol, timeframe, and focus window.

### 6. Main-Worktree Completion Gate

Main is ready for Server consumption when:

```text
BotCycle boundaries are published.
Signal decisions remain mandatory.
Signaler chart storage is configurable.
Stored Package payloads are renderer-neutral.
Storage modes have behavioral proof.
Owning design pages match implementation.
```

## Required Server-Worktree Changes

These changes belong in `nuubot5-server`.

### 1. Server Composition

Keep `cmd/nuubot-server/main.go` minimal.

It should construct configuration, invoke the Server package, and return its exit status.

Recommended ownership:

```text
cmd/nuubot-server/main.go

internal/server/
    server.go
    api.go
    pages.go

internal/botmanager/
internal/sweepmanager/
```

`server.go` owns HTTP lifecycle, dependencies, graceful shutdown, and route registration.

`api.go` owns transport parsing, validation, manager calls, and HTTP error mapping.

`pages.go` owns HTML routes and template rendering.

API handlers contain no Bot, Sweep, process, trading, or analysis policy.

### 2. Web Assets

Use one root Go package:

```text
web/
    embed.go
    templates/
        layout.html
        bots.html
        bot.html
        sweeps.html
        sweep.html
    assets/
        css/
            app.css
        js/
            app.js
            chart/
                chart.js
                theme.js
                layers/
                    base.js
                    signaler.js
                    cycles.js
                    trades.js
                    orders.js
                    fills.js
        vendor/
            echarts.min.js
            htmx.min.js
```

`web/embed.go` embeds templates and assets using Go's standard `embed` package.

Vendor ECharts and HTMX for offline single-binary deployment.

Preserve both upstream licenses.

Do not add Node, npm, Vite, React, Vue, or a frontend build step.

### 3. HTMX Boundary

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

HTMX must not own:

```text
ECharts instance lifecycle
series composition
chart zoom
layer visibility state
chart data caching
```

Chart modules must survive or recreate themselves after an HTMX fragment replacement.

Use one explicit initialization hook after chart-containing fragments settle.

### 4. Chart Composer

Create one chart controller.

It owns:

```text
ECharts instance
base data
active focus
visible layers
loaded layer cache
theme
timeframe
data zoom
resize handling
```

It must not contain layer-specific styling or data interpretation.

Each layer is one independent native JavaScript module.

Minimal layer contract:

```js
export const layer = {
  id: "cycles",
  label: "Cycles",
  elements: ["range", "start", "stop", "label"],
  build(data, context) {
    return { series: [], legend: [], extents: null };
  }
};
```

Do not add a framework or generalized plugin registry.

Import the known modules directly.

### 5. Chart Layers

#### Base

Contains:

```text
price
balance
equity
volume
```

Base loads with the page.

Base remains visible while optional layers change.

#### Signaler

Contains available stored Package fields.

Examples:

```text
fast_ma
slow_ma
regime_ma
rsi
volume_ratio
action markers
regime
```

The layer must inspect the returned field metadata.

Missing, unrecorded series display as unavailable.

The Server must not silently recalculate them.

#### BotCycles

Contains:

```text
cycle range
cycle start
cycle stop
cycle label
```

The range uses stored `start_ms` and `end_ms`.

#### Trades

Contains trade open and close markers.

#### Orders

Contains submitted, updated, cancelled, rejected, and completed markers when stored evidence supports them.

#### Fills

Contains entry and exit fill markers.

Layer names describe evidence types, not trading conclusions.

### 6. Visibility

Every optional layer has one group control.

Every layer element has one individual control.

Enabling an unloaded layer fetches it once for the current focus.

Disabling a layer hides its series without another request.

Changing focus invalidates optional-layer cache entries.

Use stable series identifiers:

```text
base:price
base:balance
base:equity
base:volume
signaler:fast_ma
cycles:range
trades:open
orders:submitted
fills:entry
```

Use ECharts legend selection for element visibility.

Custom controls may dispatch native ECharts legend actions.

### 7. Focus

Support exactly two focus levels:

```text
whole_bot
botcycle
```

Whole-Bot focus uses the complete Bot execution range at the selected timeframe.

BotCycle focus uses the stored cycle start and end boundaries.

Selecting a cycle updates:

```text
focus state
URL
summary fragment
tables
base chart range
enabled optional layers
```

Do not add trade focus.

The focus URL must be shareable and reloadable.

Suggested query:

```text
/bots/42?cycle=7
```

Absence of `cycle` means whole-Bot focus.

### 8. Lazy Data Loading

Initial Bot-page response contains:

```text
summary cards
BotCycle index
chart shell
base chart bootstrap request
```

It does not contain all transactions or Signaler series.

Suggested read endpoints:

```text
GET /api/bots/{botID}
GET /api/bots/{botID}/cycles
GET /api/bots/{botID}/chart/base
GET /api/bots/{botID}/chart/signaler
GET /api/bots/{botID}/chart/cycles
GET /api/bots/{botID}/chart/trades
GET /api/bots/{botID}/chart/orders
GET /api/bots/{botID}/chart/fills
```

Chart requests accept:

```text
cycle
timeframe
from_ms
to_ms
```

The Server validates Bot ownership, cycle existence, bounds, and timeframe.

Responses contain renderer-neutral facts.

They must not contain ECharts option objects.

Avoid one endpoint per indicator.

The Signaler response can return all stored fields for the requested range.

### 9. Theme

Use CSS custom properties and one document theme attribute.

Example:

```html
<html data-theme="dark">
```

ECharts theme values come from the same semantic color definitions.

Initial themes:

```text
dark
light
```

Store browser-local theme and layer preferences in `localStorage`.

Do not add Server-side preference storage yet.

### 10. BotManager

BotManager owns Bot use cases:

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
```

Delete remains restricted to configured Bots.

Terminal execution evidence is not deleted through configured-Bot deletion.

Views read the central database and the Bot result database.

Running-process health is stored durably by supervision.

Page reads must not call live Bot objects.

Browser polling is sufficient initially.

Do not add WebSockets or Server-Sent Events before polling proves inadequate.

### 11. SweepManager

SweepManager owns Sweep use cases:

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
```

Sweep pages reuse the same Bot report and chart components.

Additional Sweep views can include:

```text
run comparison
parameter heatmap
distribution analysis
Monte Carlo results
```

Heatmaps and Monte Carlo consume published artifacts.

HTTP handlers must not perform long analytical calculations.

Headless jobs calculate and publish those artifacts first.

### 12. Control API

Page controls and CLI should invoke the same manager-backed API.

Suggested mutation shape:

```text
POST /api/bots
POST /api/bots/{botID}/clone
POST /api/bots/{botID}/start
POST /api/bots/{botID}/pause
POST /api/bots/{botID}/stop
DELETE /api/bots/{botID}

POST /api/sweeps
POST /api/sweeps/{sweepID}/clone
POST /api/sweeps/{sweepID}/start
POST /api/sweeps/{sweepID}/pause
POST /api/sweeps/{sweepID}/stop
DELETE /api/sweeps/{sweepID}
```

The API translates transport inputs into manager commands.

Managers validate lifecycle transitions and deletion policy.

### 13. Admin

Do not create an Admin package during initial Server implementation.

Add Admin only after concrete requirements exist.

Likely future concerns:

```text
server health
process generations
storage usage
logs
configuration visibility
```

## Data Contracts

### Base Chart Response

```json
{
  "focus": {
    "kind": "whole_bot",
    "bot_id": 42,
    "cycle_number": null,
    "from_ms": 0,
    "to_ms": 0
  },
  "timeframe": "15m",
  "price": [],
  "balance": [],
  "equity": [],
  "volume": []
}
```

### Optional Layer Response

```json
{
  "layer": "signaler",
  "focus": {
    "kind": "botcycle",
    "bot_id": 42,
    "cycle_number": 7,
    "from_ms": 0,
    "to_ms": 0
  },
  "elements": {},
  "unavailable": []
}
```

Exact point encoding should be selected during implementation.

Prefer compact arrays when measured payload size justifies them.

Do not invent compression before measurement.

## Required Implementation Order

```text
1. Main: publish existing BotCycle boundaries.
2. Main: add Signaler storage policy and renderer-neutral samples.
3. Main: prove trading results remain unchanged.
4. Server: define bounded read contracts against published data.
5. Server: implement lifecycle, thin API, and embedded web assets.
6. Server: reproduce the Nuubot3 report with the base chart.
7. Server: add independent lazy-loaded chart layers.
8. Server: add whole-Bot and BotCycle focus.
9. Server: add Bot controls through BotManager.
10. Server: add Sweep controls and analysis pages.
11. Add Admin only after approval.
```

Main and Server work must remain in their matching worktrees.

Commits should remain independently reviewable.

Main data contracts should land before Server code depends on them.

## Verification

### Main

```text
BotCycle boundary tests
Signaler storage-mode tests
ResultPublisher database tests
configuration validation tests
cross-mode trading-result comparison
go test -tags noasm ./...
go vet -tags noasm ./...
git diff --check
```

### Server

```text
manager lifecycle tests
thin API handler tests
configured-only deletion tests
bounded chart query tests
whole-Bot focus tests
BotCycle focus tests
embedded asset tests
HTMX fragment tests
manual dark and light rendering
manual layer lazy-load verification
go test -tags noasm ./...
go vet -tags noasm ./...
git diff --check
```

Browser verification must confirm:

- Initial page does not fetch optional layers.
- Enabling one layer fetches only that layer.
- Disabling and reenabling uses the current-focus cache.
- Selecting a cycle bounds every active dataset.
- Returning to whole Bot restores the complete range.
- Missing Signaler series are reported as unavailable.
- Theme changes preserve chart readability.
- Controls report manager errors without losing page state.

## Explicit Non-Goals

- No trade-level chart focus.
- No indicator recalculation in HTTP handlers.
- No direct access to live Bot object state.
- No ECharts objects in persisted data or API contracts.
- No frontend framework or build pipeline.
- No WebSocket requirement.
- No Server-side UI preference storage.
- No Admin implementation.
- No generic chart plugin framework.
- No indicator-specific database schema.
- No duplication of market history per Bot database.

## Completion Definition

Preparation is complete when the main worktree publishes sufficient canonical evidence and the Server consumes it through bounded, renderer-neutral APIs.

The resulting GUI must preserve Nuubot3's useful visual format without preserving its monolithic renderer implementation.
