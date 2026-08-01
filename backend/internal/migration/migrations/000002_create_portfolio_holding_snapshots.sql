CREATE TABLE portfolio_holding_snapshots (
    id BIGSERIAL PRIMARY KEY,
    portfolio_snapshot_id BIGINT NOT NULL REFERENCES portfolio_snapshots(id) ON DELETE CASCADE,
    broker TEXT NOT NULL,
    account TEXT NOT NULL,
    symbol TEXT NOT NULL,
    market TEXT NOT NULL,
    asset_class TEXT NOT NULL,
    country TEXT NOT NULL,
    native_currency TEXT NOT NULL,
    quantity NUMERIC NOT NULL,
    average_cost_native NUMERIC NOT NULL,
    latest_price_native NUMERIC NOT NULL,
    market_value_native NUMERIC NOT NULL,
    unrealized_pnl_native NUMERIC NOT NULL,
    average_cost_sgd NUMERIC NOT NULL,
    latest_price_sgd NUMERIC NOT NULL,
    market_value_sgd NUMERIC NOT NULL,
    unrealized_pnl_sgd NUMERIC NOT NULL,
    weight NUMERIC NOT NULL,
    UNIQUE (portfolio_snapshot_id, broker, account, symbol, market)
);

CREATE INDEX portfolio_holding_snapshots_broker_ticker_idx
    ON portfolio_holding_snapshots (broker, account, symbol, market, portfolio_snapshot_id);
