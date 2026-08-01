package portfolio

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// CombinedProvider merges live broker snapshots already normalized into the
// configured base currency. It never silently drops a configured broker.
type CombinedProvider struct {
	providers []Provider
	currency  string
}

func NewCombinedProvider(currency string, providers ...Provider) *CombinedProvider {
	return &CombinedProvider{providers: providers, currency: currencyCode(currency)}
}

func (p *CombinedProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	if len(p.providers) == 0 {
		return Snapshot{}, fmt.Errorf("no live portfolio providers configured")
	}
	type result struct {
		snapshot Snapshot
		err      error
	}
	results := make(chan result, len(p.providers))
	for _, provider := range p.providers {
		go func(provider Provider) {
			snapshot, err := provider.Snapshot(ctx)
			results <- result{snapshot: snapshot, err: err}
		}(provider)
	}

	holdings := []Holding{}
	brokers := make([]string, 0, len(p.providers))
	for range p.providers {
		result := <-results
		if result.err != nil {
			return Snapshot{}, result.err
		}
		if result.snapshot.Mode != "live" {
			return Snapshot{}, fmt.Errorf("configured broker returned a non-live snapshot")
		}
		if currencyCode(result.snapshot.Currency) != p.currency {
			return Snapshot{}, fmt.Errorf("broker returned an unexpected reporting currency")
		}
		holdings = append(holdings, result.snapshot.Holdings...)
		brokers = append(brokers, result.snapshot.Broker)
	}
	sort.Strings(brokers)
	return snapshotFromHoldings(strings.Join(brokers, " + "), "live", p.currency, holdings), nil
}

func snapshotFromHoldings(broker, mode, currency string, holdings []Holding) Snapshot {
	total := 0.0
	for i := range holdings {
		holdings[i].Currency = currency
		total += holdings[i].MarketValue
	}
	for i := range holdings {
		if total != 0 {
			holdings[i].Weight = holdings[i].MarketValue / total * 100
		}
	}
	sort.Slice(holdings, func(i, j int) bool { return holdings[i].MarketValue > holdings[j].MarketValue })
	return Snapshot{Broker: broker, Mode: mode, UpdatedAt: time.Now().UTC(), TotalValue: total, Currency: currency, Holdings: holdings,
		CountryAllocation: allocations(holdings, func(h Holding) string { return h.Country }),
		MarketAllocation:  allocations(holdings, func(h Holding) string { return h.Market })}
}
