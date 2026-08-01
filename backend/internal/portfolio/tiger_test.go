package portfolio

import (
	"context"
	"testing"

	"github.com/tigerfintech/openapi-go-sdk/model"
)

func TestBuildSnapshotCalculatesWeightsAndAllocations(t *testing.T) {
	s := buildSnapshot([]model.Position{{Symbol: "A", Name: "A", Market: "US", Currency: "USD", PositionQty: 1, MarketValue: 75}, {Symbol: "B", Name: "B", Market: "SG", Currency: "SGD", PositionQty: 1, MarketValue: 50}}, "live", "USD", map[string]float64{"USD": 1, "SGD": 0.5})
	if s.TotalValue != 100 || len(s.Holdings) != 2 || s.Holdings[0].Weight != 75 {
		t.Fatalf("unexpected snapshot: %#v", s)
	}
	if s.CountryAllocation[0].Label != "United States" || s.CountryAllocation[0].Percentage != 75 {
		t.Fatalf("unexpected allocation: %#v", s.CountryAllocation)
	}
}

func TestTigerProviderConvertsSingleCurrencyPortfolioToSGD(t *testing.T) {
	provider := &TigerProvider{baseCurrency: "SGD", fxRates: func(_ context.Context, currencies []string, base string) (map[string]float64, error) {
		if len(currencies) != 1 || currencies[0] != "USD" || base != "SGD" {
			t.Fatalf("unexpected FX request: currencies=%#v base=%s", currencies, base)
		}
		return map[string]float64{"USD": 1.35, "SGD": 1}, nil
	}}
	snapshot, err := provider.buildLiveSnapshot(context.Background(), []model.Position{{Symbol: "A", Currency: "USD", PositionQty: 2, AverageCost: 10, LatestPrice: 15, MarketValue: 30, UnrealizedPnl: 10}})
	if err != nil || snapshot.Currency != "SGD" || snapshot.TotalValue != 40.5 || snapshot.Holdings[0].LatestPriceSGD != 20.25 {
		t.Fatalf("unexpected SGD snapshot: %#v, err=%v", snapshot, err)
	}
}

func TestHoldingCurrenciesNormalizesBrokerValues(t *testing.T) {
	positions := []model.Position{{Currency: " usd ", MarketValue: 1}, {Currency: "USD", MarketValue: 2}, {Currency: "HKD", MarketValue: 0}}
	currencies := holdingCurrencies(positions)
	if len(currencies) != 1 || currencies[0] != "USD" {
		t.Fatalf("unexpected currencies: %#v", currencies)
	}
}

func TestCurrencyCodeNormalizesValues(t *testing.T) {
	if got := currencyCode(" sgd "); got != "SGD" {
		t.Fatalf("unexpected currency code: %s", got)
	}
}
