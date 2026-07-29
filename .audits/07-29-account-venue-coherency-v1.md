# Account, Venue, Simulator Coherency V1

Date: 2026-07-29
Reviewer: Independent read-only agent
Reviewer status: FAIL
Final status: PASS

## Objective

Audit the Account to concrete Venue to Simulator hardcut.

Account must never access Simulator directly.

Simulator must never call Account.

Venue must own simnet Simulator lifecycle and protocol calls.

## Finding 1

Severity: Major

Status: REJECTED.

Reachability: Any warmed simnet Venue receiving a non-crossing IOC Order.

Evidence:

- `account.go` maps Market Orders to IOC.
- `simulator.go` makes every IOC cross.
- `matchAdded` therefore fills non-crossing IOC Orders.

Reviewer concern: Simulator fills IOC Orders without validating market
crossing.

Rejection:

- Current backtest BBO data is sampled once per second.
- It cannot prove intra-second Exchange ticks or exact crossing.
- Executor reads available BBO and owns IOC price selection.
- Simulator intentionally trusts submitted IOC pricing.
- Executable IOC fills at submitted price with adverse slippage.
- Crossing validation would invent precision unavailable from the input.

No production change is required.

## Proof Checked

- Account initializes one concrete Venue.
- Venue initializes and owns Simulator.
- Account operations delegate only through Venue.
- Simulator owns MarketData subscription and Exchange truth.
- Account owns dirty state and clean-sweep discovery.
- Account has no Simulator import or reference.
- Simulator has no Account callback.
- No private Venue interface or compatibility path remains.
- Venue and Simulator build with `-tags noasm`.
- Scoped formatting and whitespace checks pass.

## Proof Missing

- Account type-check remains blocked by unfinished Ledger migration files.

## Bloat Verdict

PASS.

Concrete Venue delegation is the required ownership boundary.

Simulator persistence remains isolated and unused by Venue.

No material Account, Venue, or Simulator boundary finding remains.
