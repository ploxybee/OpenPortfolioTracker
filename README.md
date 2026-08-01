# Open Portfolio Tracker

A private, self-hosted dashboard for understanding stock allocation across Tiger Brokers and Interactive Brokers. It makes concentration visible by holding, market, and country.

## Quick start

Requirements: Go 1.23+, Node.js 20+, npm, PostgreSQL, and sqlc.

```bash
cp backend/.env.example backend/.env
createdb portfolio
make db-migrate
make dev
```

Open [http://localhost:3000](http://localhost:3000). With no credentials, the dashboard intentionally shows sample data so the interface can be explored safely.

## Database and cache

PostgreSQL stores a history of normalized portfolio snapshots. Set `DATABASE_URL` in `backend/.env` to the connection string for your existing database, for example:

```dotenv
DATABASE_URL=postgres://portfolio:password@localhost:5432/portfolio?sslmode=disable
```

Run `make db-migrate` before starting the API and after pulling a change that adds a migration. The API will not start unless PostgreSQL is reachable. It serves a saved snapshot for up to one hour; after that, it fetches the broker again and appends the successful result to the history. If a scheduled refresh cannot reach the broker, the dashboard shows the most recently saved data with a warning.

Use `make sqlc` after changing the SQL schema or named queries in `backend/internal/model`; generated Go code is committed with the application.

## Connect Tiger Brokers

Create API credentials and an RSA private key in Tiger Open Platform, then place the values in `backend/.env`:

```dotenv
TIGER_ID=your_developer_id
TIGER_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
TIGER_ACCOUNT=your_account_number
TIGER_LICENSE=TBNZ
PORTFOLIO_BASE_CURRENCY=SGD
PORTFOLIO_TARGETS=Equities=70,Bonds=20,Cash=10
PORTFOLIO_REBALANCE_TOLERANCE=5
PORTFOLIO_MONTHLY_CONTRIBUTION=1000
```

Then run `make dev`. The backend loads `backend/.env` automatically and only reads positions; it contains no order-placement routes. All portfolio values use SGD, including single-currency portfolios, so stored historical values are directly comparable. The backend obtains reference conversion rates from Frankfurter; no additional API key is required.

`PORTFOLIO_TARGETS` must total 100. `PORTFOLIO_MONTHLY_CONTRIBUTION` is in the base currency and is used only to create contribution suggestions; the service never places orders.

Tiger's credentials must never be placed in the frontend or committed. See [Tiger's Go SDK documentation](https://pkg.go.dev/github.com/tigerfintech/openapi-go-sdk) for credential setup and account permissions.

## Connect Interactive Brokers

Individual IBKR Pro accounts use the REST API through IBKR's Client Portal Gateway. Install Java and the Gateway on the same Hetzner VPS as this application, bind the Gateway to loopback only, and configure a Gateway TLS certificate signed by a CA available to the API process. Do not expose port 5000 publicly and do not disable certificate validation.

```dotenv
IBKR_ENABLED=true
IBKR_GATEWAY_URL=https://localhost:5000/v1/api
IBKR_GATEWAY_CA_FILE=/etc/open-portfolio-tracker/ibkr-gateway-ca.pem
```

Start the Gateway, then open its loopback login page through a private SSH-tunneled remote desktop/browser session on the VPS and complete your normal IBKR 2FA login. IBKR requires this browser login on the same machine as the Gateway and requires reauthentication at least daily. The API first reads the available accounts and uses the first one returned, then reads all position pages. It requests no cash, margin, order, or trading endpoints.

When both Tiger and IBKR are configured, the dashboard combines their securities positions only after converting each position to SGD. If either configured broker cannot be read, the refresh fails rather than reporting a partial combined total. See [IBKR's Gateway quick start](https://www.interactivebrokers.com/docs/web-api/api/web-api/quick-start) and [portfolio endpoint documentation](https://ibkrcampus.com/campus/ibkr-api-page/webapi-doc/).

## Commands

| Command | Purpose |
| --- | --- |
| `make dev` | Start frontend and API together |
| `make api` | Start Go API at `http://localhost:8080` |
| `make web` | Start Next.js at `http://localhost:3000` |
| `make db-migrate` | Apply pending PostgreSQL migrations |
| `make sqlc` | Generate PostgreSQL query code from sqlc definitions |
| `make test` | Run backend tests and frontend linting |
| `make format` | Format code |

## API

`GET /api/v1/portfolio` returns a normalized portfolio snapshot, using a saved snapshot when it is less than an hour old. `POST /api/v1/portfolio/refresh` bypasses that cache, fetches every configured broker, and saves a new historical snapshot. Each successful live refresh stores one fixed-column ticker row per broker/account/symbol/market, as well as the full normalized snapshot JSONB. Historical monetary values are recorded in SGD alongside the broker's native values; FX rates themselves are not stored. Responses include `cacheStatus` (`cached`, `refreshed`, or `stale`) and `cachedAt` so clients can show when saved data is being used. The API remains deliberately broker-neutral so Tiger, IBKR, and CSV adapters can implement the same interface.

Read [the architecture](docs/architecture.md) before adding a broker or feature, and [the contribution guide](AGENTS.md) before working in the repository.
