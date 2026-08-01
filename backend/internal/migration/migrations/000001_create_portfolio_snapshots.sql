CREATE TABLE portfolio_snapshots (
    id BIGSERIAL PRIMARY KEY,
    portfolio_key TEXT NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL,
    snapshot JSONB NOT NULL
);

CREATE INDEX portfolio_snapshots_portfolio_key_fetched_at_idx
    ON portfolio_snapshots (portfolio_key, fetched_at DESC);
