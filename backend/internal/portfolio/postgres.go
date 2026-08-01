package portfolio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	generated "github.com/ploxybee/OpenPortfolioTracker/internal/model/generated"
)

// PostgresSnapshotStore persists normalized snapshots using sqlc-generated queries.
type PostgresSnapshotStore struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPostgresSnapshotStore creates a snapshot store backed by the supplied pool.
func NewPostgresSnapshotStore(pool *pgxpool.Pool) *PostgresSnapshotStore {
	return &PostgresSnapshotStore{pool: pool, queries: generated.New(pool)}
}

// LatestSnapshot returns the newest snapshot for a portfolio key.
func (s *PostgresSnapshotStore) LatestSnapshot(ctx context.Context, portfolioKey string) (StoredSnapshot, error) {
	row, err := s.queries.GetLatestPortfolioSnapshot(ctx, portfolioKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return StoredSnapshot{}, ErrSnapshotNotFound
	}
	if err != nil {
		return StoredSnapshot{}, fmt.Errorf("get latest portfolio snapshot: %w", err)
	}
	var snapshot Snapshot
	if err := json.Unmarshal(row.Snapshot, &snapshot); err != nil {
		return StoredSnapshot{}, fmt.Errorf("decode portfolio snapshot: %w", err)
	}
	return StoredSnapshot{Snapshot: snapshot, FetchedAt: row.FetchedAt.Time}, nil
}

// SaveSnapshot appends a normalized snapshot to the portfolio history.
func (s *PostgresSnapshotStore) SaveSnapshot(ctx context.Context, portfolioKey string, snapshot Snapshot, fetchedAt time.Time) error {
	if snapshot.Mode == "demo" {
		return nil
	}
	if snapshot.Currency != "SGD" {
		return fmt.Errorf("portfolio snapshot currency must be SGD, got %q", snapshot.Currency)
	}
	snapshot.CacheStatus = ""
	snapshot.CachedAt = time.Time{}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode portfolio snapshot: %w", err)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin portfolio snapshot transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)
	stored, err := queries.CreatePortfolioSnapshot(ctx, generated.CreatePortfolioSnapshotParams{
		PortfolioKey: portfolioKey,
		FetchedAt:    pgtype.Timestamptz{Time: fetchedAt, Valid: true},
		Snapshot:     payload,
	})
	if err != nil {
		return fmt.Errorf("save portfolio snapshot: %w", err)
	}
	for _, holding := range snapshot.Holdings {
		if holding.Currency != "SGD" {
			return fmt.Errorf("holding %s currency must be SGD, got %q", holding.Symbol, holding.Currency)
		}
		if err := queries.CreatePortfolioHoldingSnapshot(ctx, generated.CreatePortfolioHoldingSnapshotParams{
			PortfolioSnapshotID: stored.ID,
			Broker:              snapshot.Broker,
			Account:             holding.Account,
			Symbol:              holding.Symbol,
			Market:              holding.Market,
			AssetClass:          holding.AssetClass,
			Country:             holding.Country,
			NativeCurrency:      holding.NativeCurrency,
			Quantity:            holding.Quantity,
			AverageCostNative:   holding.AverageCost,
			LatestPriceNative:   holding.LatestPrice,
			MarketValueNative:   holding.NativeMarketValue,
			UnrealizedPnlNative: holding.NativeUnrealizedPnL,
			AverageCostSgd:      holding.AverageCostSGD,
			LatestPriceSgd:      holding.LatestPriceSGD,
			MarketValueSgd:      holding.MarketValue,
			UnrealizedPnlSgd:    holding.UnrealizedPnL,
			Weight:              holding.Weight,
		}); err != nil {
			return fmt.Errorf("save holding snapshot for %s: %w", holding.Symbol, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit portfolio snapshot transaction: %w", err)
	}
	return nil
}
