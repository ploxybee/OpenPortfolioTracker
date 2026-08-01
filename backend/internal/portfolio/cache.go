package portfolio

import (
	"context"
	"errors"
	"time"
)

const defaultPortfolioKey = "default"

var ErrSnapshotNotFound = errors.New("portfolio snapshot not found")

// StoredSnapshot is a saved normalized portfolio snapshot.
type StoredSnapshot struct {
	Snapshot  Snapshot
	FetchedAt time.Time
}

// SnapshotStore persists normalized snapshots without depending on a broker SDK.
type SnapshotStore interface {
	LatestSnapshot(context.Context, string) (StoredSnapshot, error)
	SaveSnapshot(context.Context, string, Snapshot, time.Time) error
}

// CachingProvider returns recent snapshots from storage and refreshes stale data.
type CachingProvider struct {
	provider Provider
	store    SnapshotStore
	ttl      time.Duration
	now      func() time.Time
}

// NewCachingProvider wraps a live provider with an hourly persistent cache.
func NewCachingProvider(provider Provider, store SnapshotStore) *CachingProvider {
	return &CachingProvider{provider: provider, store: store, ttl: time.Hour, now: time.Now}
}

// Snapshot returns a fresh cached snapshot or retrieves and stores a new one.
func (p *CachingProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	stored, err := p.store.LatestSnapshot(ctx, defaultPortfolioKey)
	if err == nil {
		if p.now().UTC().Sub(stored.FetchedAt) < p.ttl {
			return withCacheStatus(stored, "cached"), nil
		}
		return p.refresh(ctx, &stored)
	}
	if !errors.Is(err, ErrSnapshotNotFound) {
		return Snapshot{}, err
	}
	return p.refresh(ctx, nil)
}

// Refresh bypasses the cache, saving a newly retrieved snapshot when successful.
func (p *CachingProvider) Refresh(ctx context.Context) (Snapshot, error) {
	stored, err := p.store.LatestSnapshot(ctx, defaultPortfolioKey)
	if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
		return Snapshot{}, err
	}
	if errors.Is(err, ErrSnapshotNotFound) {
		return p.refresh(ctx, nil)
	}
	return p.refresh(ctx, &stored)
}

func (p *CachingProvider) refresh(ctx context.Context, stale *StoredSnapshot) (Snapshot, error) {
	snapshot, err := p.provider.Snapshot(ctx)
	if err != nil {
		if stale != nil {
			return withCacheStatus(*stale, "stale"), nil
		}
		return Snapshot{}, err
	}
	if snapshot.Mode == "demo" {
		snapshot.CacheStatus = "refreshed"
		return snapshot, nil
	}
	fetchedAt := p.now().UTC()
	if err := p.store.SaveSnapshot(ctx, defaultPortfolioKey, snapshot, fetchedAt); err != nil {
		return Snapshot{}, err
	}
	return withCacheStatus(StoredSnapshot{Snapshot: snapshot, FetchedAt: fetchedAt}, "refreshed"), nil
}

func withCacheStatus(stored StoredSnapshot, status string) Snapshot {
	snapshot := stored.Snapshot
	snapshot.CacheStatus = status
	snapshot.CachedAt = stored.FetchedAt
	return snapshot
}
