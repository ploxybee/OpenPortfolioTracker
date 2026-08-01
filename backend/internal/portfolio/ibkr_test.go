package portfolio

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func testResponse(status int, payload string) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}
}

func TestIBKRProviderFetchesFirstAccountAndAllPositionPages(t *testing.T) {
	requested := []string{}
	provider := &IBKRProvider{baseURL: "https://gateway.test/v1/api", client: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		requested = append(requested, r.URL.Path)
		switch r.URL.Path {
		case "/v1/api/portfolio/accounts":
			return testResponse(http.StatusOK, `[{"id":"FIRST"},{"id":"SECOND"}]`), nil
		case "/v1/api/portfolio/FIRST/positions/0":
			page := make([]ibkrPosition, ibkrMaxPositionsPageSize)
			for i := range page {
				page[i] = ibkrPosition{Ticker: "A", CompanyName: "A", Position: 1, MktPrice: 1, MktValue: 1, Currency: "USD", AssetClass: "STK", ListingExchange: "NASDAQ"}
			}
			payload, err := json.Marshal(page)
			return testResponse(http.StatusOK, string(payload)), err
		case "/v1/api/portfolio/FIRST/positions/1":
			return testResponse(http.StatusOK, `[{"ticker":"B","companyName":"B","position":2,"mktPrice":2,"mktValue":4,"currency":"SGD","assetClass":"STK","listingExchange":"SG"}]`), nil
		default:
			return testResponse(http.StatusNotFound, ""), nil
		}
	})}, fxRates: func(context.Context, []string, string) (map[string]float64, error) {
		return map[string]float64{"USD": 1.35, "SGD": 1}, nil
	}}

	snapshot, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if math.Abs(snapshot.TotalValue-139) > 0.000001 || len(snapshot.Holdings) != 101 || snapshot.Holdings[0].Account != "IBKR FIRST" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if strings.Join(requested, ",") != "/v1/api/portfolio/accounts,/v1/api/portfolio/FIRST/positions/0,/v1/api/portfolio/FIRST/positions/1" {
		t.Fatalf("unexpected request order: %#v", requested)
	}
}

func TestIBKRProviderNormalizesCurrenciesAndSkipsZeroPositions(t *testing.T) {
	provider := &IBKRProvider{fxRates: func(_ context.Context, currencies []string, base string) (map[string]float64, error) {
		if strings.Join(currencies, ",") != "SGD,USD" || base != "SGD" {
			t.Fatalf("unexpected FX request: %v, %s", currencies, base)
		}
		return map[string]float64{"USD": 1.35, "SGD": 1}, nil
	}}
	snapshot, err := provider.buildSnapshot(context.Background(), "PRIMARY", []ibkrPosition{
		{Ticker: "AAPL", CompanyName: "Apple", Position: 2, MktPrice: 100, MktValue: 200, AvgCost: 90, UnrealizedPnL: 20, Currency: "USD", AssetClass: "STK", ListingExchange: "NASDAQ"},
		{Ticker: "D05", CompanyName: "DBS", Position: 1, MktPrice: 30, MktValue: 30, Currency: "SGD", AssetClass: "STK", ListingExchange: "SG"},
		{Ticker: "ZERO", MktValue: 0, Currency: "USD"},
	})
	if err != nil {
		t.Fatalf("buildSnapshot: %v", err)
	}
	if snapshot.TotalValue != 300 || len(snapshot.Holdings) != 2 || snapshot.Holdings[0].Symbol != "AAPL" || snapshot.Holdings[0].Country != "United States" || snapshot.Holdings[0].MarketValue != 270 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestIBKRProviderDoesNotExposeAccountIDWhenGatewayIsUnauthenticated(t *testing.T) {
	provider := &IBKRProvider{baseURL: "https://gateway.test/v1/api", client: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return testResponse(http.StatusUnauthorized, ""), nil
	})}}
	_, err := provider.positions(context.Background(), "SECRET-ACCOUNT")
	if err == nil || strings.Contains(err.Error(), "SECRET-ACCOUNT") || !strings.Contains(err.Error(), "not authenticated") {
		t.Fatalf("unexpected error: %v", err)
	}
}
