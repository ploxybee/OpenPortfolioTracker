package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProvider struct {
	snapshot Snapshot
	err      error
	calls    int
}

func (p *fakeProvider) Snapshot(context.Context) (Snapshot, error) {
	p.calls++
	return p.snapshot, p.err
}

type fakeSnapshotStore struct {
	stored  *StoredSnapshot
	getErr  error
	saveErr error
	saves   []StoredSnapshot
}

func (s *fakeSnapshotStore) LatestSnapshot(context.Context, string) (StoredSnapshot, error) {
	if s.getErr != nil {
		return StoredSnapshot{}, s.getErr
	}
	if s.stored == nil {
		return StoredSnapshot{}, ErrSnapshotNotFound
	}
	return *s.stored, nil
}

func (s *fakeSnapshotStore) SaveSnapshot(_ context.Context, _ string, snapshot Snapshot, fetchedAt time.Time) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saves = append(s.saves, StoredSnapshot{Snapshot: snapshot, FetchedAt: fetchedAt})
	return nil
}

func TestCachingProviderReturnsRecentSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	provider := &fakeProvider{snapshot: Snapshot{Broker: "live"}}
	store := &fakeSnapshotStore{stored: &StoredSnapshot{Snapshot: Snapshot{Broker: "saved"}, FetchedAt: now.Add(-59 * time.Minute)}}
	cache := NewCachingProvider(provider, store)
	cache.now = func() time.Time { return now }

	snapshot, err := cache.Snapshot(context.Background())
	if err != nil || snapshot.Broker != "saved" || snapshot.CacheStatus != "cached" || provider.calls != 0 {
		t.Fatalf("unexpected cached result: %#v, err=%v, calls=%d", snapshot, err, provider.calls)
	}
}

func TestCachingProviderRefreshesExpiredSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	provider := &fakeProvider{snapshot: Snapshot{Broker: "live"}}
	store := &fakeSnapshotStore{stored: &StoredSnapshot{Snapshot: Snapshot{Broker: "saved"}, FetchedAt: now.Add(-time.Hour)}}
	cache := NewCachingProvider(provider, store)
	cache.now = func() time.Time { return now }

	snapshot, err := cache.Snapshot(context.Background())
	if err != nil || snapshot.Broker != "live" || snapshot.CacheStatus != "refreshed" || len(store.saves) != 1 || provider.calls != 1 {
		t.Fatalf("unexpected refresh result: %#v, err=%v", snapshot, err)
	}
}

func TestCachingProviderForceRefreshBypassesRecentSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	provider := &fakeProvider{snapshot: Snapshot{Broker: "live"}}
	store := &fakeSnapshotStore{stored: &StoredSnapshot{Snapshot: Snapshot{Broker: "saved"}, FetchedAt: now.Add(-time.Minute)}}
	cache := NewCachingProvider(provider, store)
	cache.now = func() time.Time { return now }

	snapshot, err := cache.Refresh(context.Background())
	if err != nil || snapshot.Broker != "live" || snapshot.CacheStatus != "refreshed" || len(store.saves) != 1 || provider.calls != 1 {
		t.Fatalf("unexpected forced refresh result: %#v, err=%v", snapshot, err)
	}
}

func TestCachingProviderReturnsStaleSnapshotWhenRefreshFails(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	provider := &fakeProvider{err: errors.New("broker unavailable")}
	store := &fakeSnapshotStore{stored: &StoredSnapshot{Snapshot: Snapshot{Broker: "saved"}, FetchedAt: now.Add(-2 * time.Hour)}}
	cache := NewCachingProvider(provider, store)
	cache.now = func() time.Time { return now }

	snapshot, err := cache.Refresh(context.Background())
	if err != nil || snapshot.Broker != "saved" || snapshot.CacheStatus != "stale" || provider.calls != 1 {
		t.Fatalf("unexpected stale result: %#v, err=%v", snapshot, err)
	}
}

func TestCachingProviderDoesNotStoreDemoSnapshots(t *testing.T) {
	provider := &fakeProvider{snapshot: Snapshot{Mode: "demo"}}
	store := &fakeSnapshotStore{}
	cache := NewCachingProvider(provider, store)

	snapshot, err := cache.Snapshot(context.Background())
	if err != nil || snapshot.CacheStatus != "refreshed" || len(store.saves) != 0 {
		t.Fatalf("unexpected demo result: %#v, err=%v, saves=%d", snapshot, err, len(store.saves))
	}
}
