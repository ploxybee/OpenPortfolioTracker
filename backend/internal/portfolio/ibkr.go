package portfolio

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	appconfig "github.com/ploxybee/OpenPortfolioTracker/config"
)

const ibkrMaxPositionsPageSize = 100

// IBKRProvider reads positions through an authenticated local Client Portal
// Gateway. It intentionally contains no order-related endpoint calls.
type IBKRProvider struct {
	baseURL string
	client  *http.Client
	fxRates func(context.Context, []string, string) (map[string]float64, error)
}

type ibkrAccount struct {
	ID string `json:"id"`
}

type ibkrPosition struct {
	ContractDesc    string  `json:"contractDesc"`
	CompanyName     string  `json:"companyName"`
	Ticker          string  `json:"ticker"`
	Position        float64 `json:"position"`
	MktPrice        float64 `json:"mktPrice"`
	MktValue        float64 `json:"mktValue"`
	AvgCost         float64 `json:"avgCost"`
	UnrealizedPnL   float64 `json:"unrealizedPnl"`
	UnrealizedPnLP  float64 `json:"unrealizedPnlPercent"`
	Currency        string  `json:"currency"`
	AssetClass      string  `json:"assetClass"`
	ListingExchange string  `json:"listingExchange"`
}

// NewIBKRProvider creates an HTTPS client that trusts only the configured
// Gateway CA. Certificate verification is never disabled.
func NewIBKRProvider(config appconfig.IBKRConfig) (*IBKRProvider, error) {
	pem, err := os.ReadFile(config.CAFile)
	if err != nil {
		return nil, fmt.Errorf("read IBKR Gateway CA certificate: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("IBKR Gateway CA file contains no certificates")
	}
	parsedURL, err := url.Parse(config.GatewayURL)
	if err != nil {
		return nil, fmt.Errorf("parse IBKR Gateway URL: %w", err)
	}
	return &IBKRProvider{
		baseURL: strings.TrimRight(parsedURL.String(), "/"),
		client: &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
		}},
		fxRates: frankfurterFXRates,
	}, nil
}

// Snapshot retrieves every page of the first account returned by IBKR, then
// normalizes all values into the configured reporting currency.
func (p *IBKRProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	accounts, err := p.accounts(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	if len(accounts) == 0 || strings.TrimSpace(accounts[0].ID) == "" {
		return Snapshot{}, fmt.Errorf("IBKR returned no portfolio accounts")
	}
	positions, err := p.positions(ctx, accounts[0].ID)
	if err != nil {
		return Snapshot{}, err
	}
	return p.buildSnapshot(ctx, accounts[0].ID, positions)
}

func (p *IBKRProvider) accounts(ctx context.Context) ([]ibkrAccount, error) {
	var accounts []ibkrAccount
	if err := p.getJSON(ctx, "/portfolio/accounts", &accounts); err != nil {
		return nil, fmt.Errorf("fetch IBKR portfolio accounts: %w", err)
	}
	return accounts, nil
}

func (p *IBKRProvider) positions(ctx context.Context, accountID string) ([]ibkrPosition, error) {
	positions := []ibkrPosition{}
	for pageID := 0; ; pageID++ {
		var page []ibkrPosition
		endpoint := "/portfolio/" + url.PathEscape(accountID) + "/positions/" + fmt.Sprint(pageID)
		if err := p.getJSON(ctx, endpoint, &page); err != nil {
			return nil, fmt.Errorf("fetch IBKR positions: %w", err)
		}
		positions = append(positions, page...)
		if len(page) < ibkrMaxPositionsPageSize {
			return positions, nil
		}
	}
}

func (p *IBKRProvider) getJSON(ctx context.Context, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path.Clean("/"+endpoint), nil)
	if err != nil {
		return err
	}
	response, err := p.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			return fmt.Errorf("IBKR Gateway session is not authenticated")
		}
		return fmt.Errorf("IBKR Gateway returned %s", response.Status)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("decode IBKR Gateway response: %w", err)
	}
	return nil
}

func (p *IBKRProvider) buildSnapshot(ctx context.Context, accountID string, positions []ibkrPosition) (Snapshot, error) {
	currencies := make([]string, 0, len(positions))
	for _, position := range positions {
		if position.MktValue != 0 {
			currencies = append(currencies, currencyCode(position.Currency))
		}
	}
	currencies = uniqueCurrencies(currencies)
	rates, err := p.fxRates(ctx, currencies, "SGD")
	if err != nil {
		return Snapshot{}, fmt.Errorf("fetch currency conversion rates: %w", err)
	}
	holdings := make([]Holding, 0, len(positions))
	for _, position := range positions {
		if position.MktValue == 0 {
			continue
		}
		currency := currencyCode(position.Currency)
		rate, ok := rates[currency]
		if !ok || rate <= 0 {
			return Snapshot{}, fmt.Errorf("missing FX rate for an IBKR position currency")
		}
		market := strings.TrimSpace(position.ListingExchange)
		symbol := firstNonBlank(position.Ticker, position.ContractDesc)
		holdings = append(holdings, Holding{
			Symbol: symbol, Name: firstNonBlank(position.CompanyName, position.ContractDesc, symbol), Account: "IBKR " + accountID,
			AssetClass: ibkrAssetClass(position.AssetClass), Market: market, Country: countryForMarket(market), Currency: "SGD", NativeCurrency: currency,
			Quantity: position.Position, AverageCost: position.AvgCost, LatestPrice: position.MktPrice, NativeMarketValue: position.MktValue,
			NativeUnrealizedPnL: position.UnrealizedPnL, AverageCostSGD: position.AvgCost * rate, LatestPriceSGD: position.MktPrice * rate,
			MarketValue: position.MktValue * rate, UnrealizedPnL: position.UnrealizedPnL * rate, UnrealizedPnLPc: position.UnrealizedPnLP,
		})
	}
	return snapshotFromHoldings("Interactive Brokers", "live", "SGD", holdings), nil
}

func uniqueCurrencies(currencies []string) []string {
	set := map[string]struct{}{}
	for _, currency := range currencies {
		if currency != "" {
			set[currency] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for currency := range set {
		out = append(out, currency)
	}
	sort.Strings(out)
	return out
}

func ibkrAssetClass(assetClass string) string {
	switch strings.ToUpper(strings.TrimSpace(assetClass)) {
	case "STK", "ETF":
		return "Equities"
	case "BOND":
		return "Bonds"
	default:
		return firstNonBlank(strings.TrimSpace(assetClass), "Securities")
	}
}
