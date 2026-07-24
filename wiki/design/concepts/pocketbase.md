# PocketBase

Status: Approved — unimplemented.
Covers: `cmd/nuubot-server/main.go`
Purpose: Provide the Server-owned web, API, authentication, realtime, and SQLite application framework.

## Official Sources

- [PocketBase documentation](https://pocketbase.io/docs/)
- [PocketBase Go extension documentation](https://pocketbase.io/docs/go-overview/)
- [PocketBase database documentation](https://pocketbase.io/docs/go-database/)

## Scope

PocketBase is embedded in `nuubot-server`.

PocketBase replaces the planned PostgreSQL persistence direction.

PocketBase provides:

- the HTTP web server;
- custom and collection APIs;
- authentication and authorization rules;
- the administration dashboard;
- realtime subscriptions;
- static file serving;
- migrations;
- SQLite connection management; and
- physical write serialization.

## Ownership

Server owns one PocketBase application.

That PocketBase application owns the writable SQLite database.

Runners and Bots call Server-owned store operations.

Runners and Bots MUST NOT open the PocketBase database directly.

The existing BtRunner Sweep database remains separate, read-only, and immutable.

## Write Flow

```text
Bot or Runner
  -> Server-owned store operation
  -> PocketBase transaction
  -> serialized SQLite write
```

PocketBase queues concurrent writes through its single write connection.

SQLite WAL permits concurrent reads while one write transaction runs.

Nuubot MUST NOT add another generic write queue or database mutex.

## Nuubot Responsibilities

PocketBase manages physical database concurrency.

Nuubot still owns:

- valid lifecycle transitions;
- conditional updates;
- Bot generations;
- idempotent commands;
- unique CLOIDs and venue event identities;
- domain transactions; and
- domain constraints and indexes.

Critical multi-record changes MUST use one short PocketBase transaction.

External calls and slow work MUST remain outside database transactions.

Market ticks MUST remain runtime events instead of database writes.

## API Boundary

PocketBase collection APIs MAY expose safe configuration and observation data.

Trading mutations MUST pass through custom Server routes and manager operations.

The PocketBase administration dashboard manages records and application settings.

The administration dashboard is an infrastructure administration tool.

Nuubot owns its trading interface, operational dashboards, analytics, and
reports.

The Nuubot interface uses PocketBase authentication, APIs, and realtime
subscriptions.

PocketBase MAY serve the Nuubot web assets.

## Responsibility Split

| User and Nuubot provide | PocketBase provides |
|---|---|
| Trading interface design and workflows. | HTTP web server and static asset serving. |
| Operational dashboards, charts, analytics, and reports. | API routing and realtime subscriptions. |
| Trading behavior and manager operations. | Custom route and hook framework. |
| Collection schemas, fields, indexes, and migrations. | Migration runner and collection management. |
| Domain DML and transaction boundaries. | SQLite connections, transactions, WAL, and serialized writes. |
| Valid state transitions and generation checks. | Durable execution of accepted writes. |
| Unique CLOID and venue event requirements. | Unique indexes and constraint enforcement. |
| Roles, permissions, and access policy. | Authentication, tokens, and API rule enforcement. |
| Trading and reporting user experience. | Infrastructure administration dashboard. |

PocketBase provides mechanisms. Nuubot defines trading meaning and policy.

## Deployment

One `nuubot-server` process owns one PocketBase application and one writable SQLite database.

Windows and the intended Ubuntu 24 deployment use the same pure-Go SQLite boundary.

Multiple PocketBase processes MUST NOT share one writable database.

## Required Proof

- Five concurrent Bots complete short writes without lock failures.
- Concurrent reads continue during writes.
- Duplicate CLOIDs and venue events fail.
- Stale Bot generations cannot mutate current state.
- Failed multi-record operations roll back completely.
- Runners have no direct writable database path.
- HTTP, API, authentication, administration, and realtime start and stop with Server.
