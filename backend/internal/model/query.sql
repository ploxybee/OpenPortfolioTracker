-- name: CreatePortfolioSnapshot :one
INSERT INTO portfolio_snapshots (portfolio_key, fetched_at, snapshot)
VALUES ($1, $2, $3)
RETURNING id, portfolio_key, fetched_at, snapshot;

-- name: GetLatestPortfolioSnapshot :one
SELECT id, portfolio_key, fetched_at, snapshot
FROM portfolio_snapshots
WHERE portfolio_key = $1
ORDER BY fetched_at DESC
LIMIT 1;

-- name: CreatePortfolioHoldingSnapshot :exec
INSERT INTO portfolio_holding_snapshots (
    portfolio_snapshot_id, broker, account, symbol, market, asset_class, country,
    native_currency, quantity, average_cost_native, latest_price_native,
    market_value_native, unrealized_pnl_native, average_cost_sgd,
    latest_price_sgd, market_value_sgd, unrealized_pnl_sgd, weight
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
);

-- name: ListPortfolioValueHistory :many
SELECT snapshots.fetched_at, SUM(holdings.market_value_sgd)::DOUBLE PRECISION AS total_value_sgd
FROM portfolio_snapshots AS snapshots
JOIN portfolio_holding_snapshots AS holdings ON holdings.portfolio_snapshot_id = snapshots.id
WHERE snapshots.portfolio_key = $1
GROUP BY snapshots.id, snapshots.fetched_at
ORDER BY snapshots.fetched_at;

-- name: ListTickerHistory :many
SELECT snapshots.fetched_at, holdings.quantity, holdings.native_currency,
       holdings.latest_price_native, holdings.market_value_native,
       holdings.latest_price_sgd, holdings.market_value_sgd, holdings.weight
FROM portfolio_snapshots AS snapshots
JOIN portfolio_holding_snapshots AS holdings ON holdings.portfolio_snapshot_id = snapshots.id
WHERE snapshots.portfolio_key = $1
  AND holdings.broker = $2
  AND holdings.account = $3
  AND holdings.symbol = $4
  AND holdings.market = $5
ORDER BY snapshots.fetched_at;
