# Architecture

## Goal

Open Portfolio Tracker is a read-only personal analytics application. It makes holdings and allocation easy to understand; it does not place, amend, or cancel orders.

## Components

```text
Next.js dashboard (localhost:3000)
        │ GET /api/v1/portfolio / POST /api/v1/portfolio/refresh
        ▼
Go API (localhost:8080)
        │ hourly cache lookup
        ▼
PostgreSQL ◄── normalized snapshots (JSONB history)
        │ cache miss or forced refresh
        ▼
Portfolio Provider interface
        ▼
Combined provider ──► Tiger OpenAPI adapter ──► Tiger Brokers
        │
        └────────────► IBKR Gateway REST adapter ──► Client Portal Gateway (loopback) ──► IBKR
```

- `frontend/` renders the dashboard only. It never receives broker credentials.
- `backend/cmd/api` starts the HTTP service.
- `backend/config` loads and validates environment configuration before the service starts.
- `backend/internal/api` owns HTTP concerns: routing, CORS, errors, and JSON.
- `backend/internal/portfolio` owns the broker-neutral data model, allocation calculations, and provider adapters.
- `backend/internal/model` owns the sqlc schema, named PostgreSQL queries, and generated query package. `backend/internal/migration` owns embedded, versioned database migrations.

## Contract

The dashboard calls `GET /api/v1/portfolio`. A snapshot includes holdings plus pre-computed allocations by country and market. The API first reads the newest saved snapshot for its single default portfolio key. It returns one saved within an hour; otherwise it reads the configured provider and appends the resulting normalized snapshot to PostgreSQL. `POST /api/v1/portfolio/refresh` forces that provider read. On a broker error with a previously saved snapshot, the API returns that snapshot with `cacheStatus: "stale"`; it never mislabels it as current data. Every portfolio is reported in SGD, including single-currency portfolios. The Tiger adapter obtains Frankfurter reference rates as needed and retains native holding values alongside their SGD equivalents.

`DATABASE_URL` is required at API startup, and the process fails fast when PostgreSQL is unreachable. Migrations are intentionally explicit: operators run `make db-migrate` before deployment. The `portfolio_snapshots` table is append-only, retaining the complete normalized JSONB response. Its `portfolio_holding_snapshots` child table records one fixed-column row per broker/account/ticker/market for each successful live refresh, which supports future history charts without parsing JSONB. Demo snapshots and cache reads are not recorded.

All current portfolio and history values use SGD. Each holding-history row also retains native-currency prices, market value, and P&L for audit/detail views, but the application does not retain the FX rate that produced its SGD values. Existing non-SGD snapshot rows are intentionally not backfilled because their original FX rate is unavailable.

## Tiger integration

The Tiger provider uses the official Go SDK and calls the positions endpoint with `SecType=STK`. It enters demo mode if `TIGER_ID`, `TIGER_PRIVATE_KEY`, or `TIGER_ACCOUNT` is absent. Demo mode is intentional: local UI development requires no brokerage connection.

## IBKR integration

The IBKR adapter is enabled only by `IBKR_ENABLED=true`. It connects over validated HTTPS to the Client Portal Gateway on the same host, using a configured CA file; it never disables TLS verification or exposes Gateway traffic to the public network. Individual IBKR Pro users authenticate with a browser and 2FA on that same host at least daily. The adapter calls `/portfolio/accounts` before selecting the first account and paging through `/portfolio/{accountId}/positions/{pageId}`. It calls no trading, cash, or margin endpoints.

When one or both broker integrations are configured, the combined provider concurrently fetches each live snapshot and recomputes holding weights and allocations only after every snapshot is normalized to SGD. A configured broker failure fails the entire live refresh, preventing partial totals. If neither provider is configured, Tiger's isolated demo snapshot is used for local UI development.

## Security boundaries

- Keep `backend/.env` local and ignored.
- Do not add brokerage credentials to `NEXT_PUBLIC_*` variables.
- Keep the application read-only until an explicit product decision changes that boundary.
- Avoid logging request signing material or account identifiers.
- Keep the IBKR Gateway loopback-only and verify its certificate with `IBKR_GATEWAY_CA_FILE`.
