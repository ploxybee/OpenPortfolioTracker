package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ploxybee/OpenPortfolioTracker/internal/portfolio"
)

type testProvider struct {
	snapshot      portfolio.Snapshot
	snapshotCalls int
	refreshCalls  int
}

func (p *testProvider) Snapshot(context.Context) (portfolio.Snapshot, error) {
	p.snapshotCalls++
	return p.snapshot, nil
}

func (p *testProvider) Refresh(context.Context) (portfolio.Snapshot, error) {
	p.refreshCalls++
	p.snapshot.CacheStatus = "refreshed"
	return p.snapshot, nil
}

func TestPortfolioRoutesUseCacheAndForceRefresh(t *testing.T) {
	provider := &testProvider{snapshot: portfolio.Snapshot{Broker: "Test Broker", CacheStatus: "cached"}}
	server := NewServer(provider)

	get := httptest.NewRecorder()
	server.Routes().ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/v1/portfolio", nil))
	if get.Code != http.StatusOK || provider.snapshotCalls != 1 || provider.refreshCalls != 0 {
		t.Fatalf("unexpected GET result: status=%d snapshot=%d refresh=%d", get.Code, provider.snapshotCalls, provider.refreshCalls)
	}

	refresh := httptest.NewRecorder()
	server.Routes().ServeHTTP(refresh, httptest.NewRequest(http.MethodPost, "/api/v1/portfolio/refresh", nil))
	if refresh.Code != http.StatusOK || provider.refreshCalls != 1 {
		t.Fatalf("unexpected refresh result: status=%d refresh=%d", refresh.Code, provider.refreshCalls)
	}
	var response portfolio.Snapshot
	if err := json.NewDecoder(refresh.Body).Decode(&response); err != nil || response.CacheStatus != "refreshed" {
		t.Fatalf("unexpected refresh body: %#v, err=%v", response, err)
	}
}

func TestCorsAllowsRefresh(t *testing.T) {
	server := NewServer(&testProvider{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/portfolio/refresh", nil)
	server.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected CORS response: %#v", response.Result())
	}
}
