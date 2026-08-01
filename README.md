# Open Portfolio Tracker

A private, self-hosted dashboard for understanding stock allocation across brokers. The first release imports stock positions from Tiger Brokers and makes concentration visible by holding, market, and country. IBKR is planned next.

## Quick start

Requirements: Go 1.23+, Node.js 20+, and npm.

```bash
cp backend/.env.example backend/.env
make dev
```

Open [http://localhost:3000](http://localhost:3000). With no credentials, the dashboard intentionally shows sample data so the interface can be explored safely.

## Connect Tiger Brokers

Create API credentials and an RSA private key in Tiger Open Platform, then place the values in `backend/.env`:

```dotenv
TIGER_ID=your_developer_id
TIGER_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
TIGER_ACCOUNT=your_account_number
TIGER_LICENSE=TBNZ
PORTFOLIO_BASE_CURRENCY=USD
PORTFOLIO_TARGETS=Equities=70,Bonds=20,Cash=10
PORTFOLIO_REBALANCE_TOLERANCE=5
PORTFOLIO_MONTHLY_CONTRIBUTION=1000
```

Then run `make dev`. The backend loads `backend/.env` automatically and only reads positions; it contains no order-placement routes. A portfolio containing one currency works directly. For mixed-currency portfolios, set `PORTFOLIO_BASE_CURRENCY` to your preferred reporting currency (for example `SGD`). The backend obtains daily reference conversion rates from Frankfurter; no additional API key is required.

`PORTFOLIO_TARGETS` must total 100. `PORTFOLIO_MONTHLY_CONTRIBUTION` is in the base currency and is used only to create contribution suggestions; the service never places orders.

Tiger's credentials must never be placed in the frontend or committed. See [Tiger's Go SDK documentation](https://pkg.go.dev/github.com/tigerfintech/openapi-go-sdk) for credential setup and account permissions.

## Commands

| Command | Purpose |
| --- | --- |
| `make dev` | Start frontend and API together |
| `make api` | Start Go API at `http://localhost:8080` |
| `make web` | Start Next.js at `http://localhost:3000` |
| `make test` | Run backend tests and frontend linting |
| `make format` | Format code |

## API

`GET /api/v1/portfolio` returns a normalized portfolio snapshot. It is deliberately broker-neutral so Tiger, IBKR, and CSV adapters can implement the same interface.

Read [the architecture](docs/architecture.md) before adding a broker or feature, and [the contribution guide](AGENTS.md) before working in the repository.
