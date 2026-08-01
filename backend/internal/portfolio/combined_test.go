package portfolio

import (
	"context"
	"errors"
	"testing"
)

func TestCombinedProviderRecalculatesWeightsAndAllocations(t *testing.T) {
	first := &fakeProvider{snapshot: Snapshot{Broker: "Tiger Brokers", Mode: "live", Currency: "SGD", Holdings: []Holding{{Symbol: "A", Country: "United States", Market: "US", MarketValue: 75}}}}
	second := &fakeProvider{snapshot: Snapshot{Broker: "Interactive Brokers", Mode: "live", Currency: "SGD", Holdings: []Holding{{Symbol: "B", Country: "Singapore", Market: "SG", MarketValue: 25}}}}
	snapshot, err := NewCombinedProvider("SGD", first, second).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if snapshot.Broker != "Interactive Brokers + Tiger Brokers" || snapshot.TotalValue != 100 || snapshot.Holdings[0].Weight != 75 || snapshot.CountryAllocation[0].Percentage != 75 {
		t.Fatalf("unexpected combined snapshot: %#v", snapshot)
	}
}

func TestCombinedProviderFailsWhenAnyLiveBrokerFails(t *testing.T) {
	good := &fakeProvider{snapshot: Snapshot{Broker: "Tiger Brokers", Mode: "live", Currency: "SGD"}}
	bad := &fakeProvider{err: errors.New("gateway unavailable")}
	if _, err := NewCombinedProvider("SGD", good, bad).Snapshot(context.Background()); err == nil {
		t.Fatal("expected combined provider error")
	}
}
