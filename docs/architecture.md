# Architecture

## Goal

Open Portfolio Tracker is a read-only personal analytics application. It makes holdings and allocation easy to understand; it does not place, amend, or cancel orders.

## Components

```text
Next.js dashboard (localhost:3000)
        │ GET /api/v1/portfolio
        ▼
Go API (localhost:8080)
        │ Portfolio Provider interface
        ▼
Tiger OpenAPI adapter ──► Tiger Brokers
        │
        └── future: IBKR adapter / CSV import
```

- `frontend/` renders the dashboard only. It never receives broker credentials.
- `backend/cmd/api` starts the HTTP service.
- `backend/config` loads and validates environment configuration before the service starts.
- `backend/internal/api` owns HTTP concerns: routing, CORS, errors, and JSON.
- `backend/internal/portfolio` owns the broker-neutral data model, allocation calculations, and provider adapters.

## Contract

The dashboard calls `GET /api/v1/portfolio`. A snapshot includes holdings plus pre-computed allocations by country and market. Single-currency portfolios calculate allocation directly. For mixed-currency portfolios, the Tiger adapter uses Frankfurter's no-key daily reference-rate API to convert values to `PORTFOLIO_BASE_CURRENCY` (USD by default) before calculating totals and weights. It retains each position's `nativeCurrency`; per-share prices remain native broker values and are not currently shown in the dashboard.

## Tiger integration

The Tiger provider uses the official Go SDK and calls the positions endpoint with `SecType=STK`. It enters demo mode if `TIGER_ID`, `TIGER_PRIVATE_KEY`, or `TIGER_ACCOUNT` is absent. Demo mode is intentional: local UI development requires no brokerage connection.

## Adding IBKR

Implement `portfolio.Provider`, normalize IBKR positions into `Holding`, and add a composition provider when both sources should be combined. Keep broker details out of HTTP handlers and frontend components. Add fixtures and unit tests covering symbol normalization, market-to-country mapping, and allocations.

## Security boundaries

- Keep `backend/.env` local and ignored.
- Do not add brokerage credentials to `NEXT_PUBLIC_*` variables.
- Keep the application read-only until an explicit product decision changes that boundary.
- Avoid logging request signing material or account identifiers.
