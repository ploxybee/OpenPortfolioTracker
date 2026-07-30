package portfolio

import (
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
