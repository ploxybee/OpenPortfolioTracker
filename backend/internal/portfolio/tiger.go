package portfolio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tigerfintech/openapi-go-sdk/config"
	"github.com/tigerfintech/openapi-go-sdk/model"
	"github.com/tigerfintech/openapi-go-sdk/trade"
)

type TigerProvider struct {
	configured                                          bool
	tigerID, privateKey, account, license, baseCurrency string
}

func NewTigerProviderFromEnv() *TigerProvider {
	p := &TigerProvider{tigerID: os.Getenv("TIGER_ID"), privateKey: os.Getenv("TIGER_PRIVATE_KEY"), account: os.Getenv("TIGER_ACCOUNT"), license: os.Getenv("TIGER_LICENSE"), baseCurrency: os.Getenv("PORTFOLIO_BASE_CURRENCY")}
	if p.baseCurrency == "" {
		p.baseCurrency = "USD"
	}
	p.configured = p.tigerID != "" && p.privateKey != "" && p.account != ""
	return p
}

func (p *TigerProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	if !p.configured {
		return demoSnapshot(), nil
	}
	cfg, err := config.NewClientConfig(config.WithTigerID(p.tigerID), config.WithPrivateKey(strings.ReplaceAll(p.privateKey, `\n`, "\n")), config.WithAccount(p.account), config.WithLicense(p.license), config.WithLanguage("en_US"))
	if err != nil {
		return Snapshot{}, fmt.Errorf("configure Tiger client: %w", err)
	}
	tradeClient := trade.NewTradeClientFromConfig(cfg)
	positions, err := tradeClient.Positions(model.PositionsRequest{SecType: "STK", Currency: "ALL"})
	if err != nil {
		return Snapshot{}, fmt.Errorf("fetch Tiger positions: %w", err)
	}
	// A single-currency portfolio has no FX requirement. This avoids an
	// unnecessary assets request for standard accounts, where that response may
	// omit a currency row when there is no cash balance in that currency.
	if currencies := holdingCurrencies(positions); len(currencies) == 1 {
		currency := currencies[0]
		return buildSnapshot(positions, "live", currency, map[string]float64{currency: 1}), nil
	}
	rates, err := frankfurterFXRates(ctx, holdingCurrencies(positions), p.baseCurrency)
	if err != nil {
		return Snapshot{}, fmt.Errorf("fetch currency conversion rates: %w", err)
	}
	baseCurrency := currencyCode(p.baseCurrency)
	if err := validateFXRates(positions, rates); err != nil {
		return Snapshot{}, err
	}
	return buildSnapshot(positions, "live", baseCurrency, rates), nil
}

// tigerFXRates uses assets rather than aggregate_assets because aggregate_assets is
// unavailable for standard Tiger accounts. The assets response includes the account
// exchange rate for each currency when market_value is requested.
const frankfurterAPI = "https://api.frankfurter.dev/v2/rate"

// frankfurterFXRates fetches direct pairs (for example USD/SGD) from a
// no-key, central-bank-rate API. Tiger's account and quote endpoints are not
// consistently available across its account types for currency conversion.
func frankfurterFXRates(ctx context.Context, currencies []string, baseCurrency string) (map[string]float64, error) {
	baseCurrency = currencyCode(baseCurrency)
	rates := map[string]float64{baseCurrency: 1}
	client := &http.Client{Timeout: 8 * time.Second}
	for _, currency := range currencies {
		currency = currencyCode(currency)
		if currency == baseCurrency {
			continue
		}
		rate, err := frankfurterRate(ctx, client, currency, baseCurrency)
		if err != nil {
			return nil, err
		}
		rates[currency] = rate
	}
	return rates, nil
}

func frankfurterRate(ctx context.Context, client *http.Client, from, to string) (float64, error) {
	endpoint := frankfurterAPI + "/" + url.PathEscape(from) + "/" + url.PathEscape(to)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("FX provider returned %s for %s/%s", response.Status, from, to)
	}
	var payload struct {
		Base  string  `json:"base"`
		Quote string  `json:"quote"`
		Rate  float64 `json:"rate"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return 0, err
	}
	if currencyCode(payload.Base) != from || currencyCode(payload.Quote) != to || payload.Rate <= 0 {
		return 0, fmt.Errorf("FX provider returned an invalid rate for %s/%s", from, to)
	}
	return payload.Rate, nil
}

func buildSnapshot(positions []model.Position, mode, baseCurrency string, rates map[string]float64) Snapshot {
	holdings := make([]Holding, 0, len(positions))
	total := 0.0
	for _, p := range positions {
		if p.MarketValue == 0 {
			continue
		}
		rate := rates[currencyCode(p.Currency)]
		marketValue := p.MarketValue * rate
		total += marketValue
		holdings = append(holdings, Holding{Symbol: p.Symbol, Name: firstNonBlank(p.Name, p.Symbol), Market: p.Market, Country: countryForMarket(p.Market), Currency: baseCurrency, NativeCurrency: p.Currency, Quantity: p.PositionQty, AverageCost: p.AverageCost, LatestPrice: p.LatestPrice, MarketValue: marketValue, UnrealizedPnL: p.UnrealizedPnl * rate, UnrealizedPnLPc: p.UnrealizedPnlPercent})
	}
	for i := range holdings {
		if total != 0 {
			holdings[i].Weight = holdings[i].MarketValue / total * 100
		}
	}
	sort.Slice(holdings, func(i, j int) bool { return holdings[i].MarketValue > holdings[j].MarketValue })
	return Snapshot{Broker: "Tiger Brokers", Mode: mode, UpdatedAt: time.Now().UTC(), TotalValue: total, Currency: baseCurrency, Holdings: holdings, CountryAllocation: allocations(holdings, func(h Holding) string { return h.Country }), MarketAllocation: allocations(holdings, func(h Holding) string { return h.Market })}
}

func validateFXRates(positions []model.Position, rates map[string]float64) error {
	for _, position := range positions {
		if position.MarketValue == 0 {
			continue
		}
		if _, ok := rates[currencyCode(position.Currency)]; !ok {
			return fmt.Errorf("Tiger did not return an FX rate for %s; cannot calculate accurate portfolio weights", position.Currency)
		}
	}
	return nil
}

func holdingCurrencies(positions []model.Position) []string {
	set := map[string]struct{}{}
	for _, position := range positions {
		if position.MarketValue != 0 {
			set[currencyCode(position.Currency)] = struct{}{}
		}
	}
	currencies := make([]string, 0, len(set))
	for currency := range set {
		if currency != "" {
			currencies = append(currencies, currency)
		}
	}
	sort.Strings(currencies)
	return currencies
}

func currencyCode(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func allocations(holdings []Holding, key func(Holding) string) []Allocation {
	totals := map[string]float64{}
	total := 0.0
	for _, h := range holdings {
		totals[key(h)] += h.MarketValue
		total += h.MarketValue
	}
	out := make([]Allocation, 0, len(totals))
	for label, value := range totals {
		percentage := 0.0
		if total != 0 {
			percentage = value / total * 100
		}
		out = append(out, Allocation{Label: label, Value: value, Percentage: percentage})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value > out[j].Value })
	return out
}
func firstNonBlank(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return "Unknown"
}
func countryForMarket(market string) string {
	switch strings.ToUpper(market) {
	case "US":
		return "United States"
	case "SG":
		return "Singapore"
	case "HK":
		return "Hong Kong"
	case "AU":
		return "Australia"
	case "CN":
		return "China"
	case "UK", "LSE":
		return "United Kingdom"
	default:
		return "Other"
	}
}
func demoSnapshot() Snapshot {
	positions := []model.Position{{Symbol: "VOO", Name: "Vanguard S&P 500 ETF", Market: "US", Currency: "USD", PositionQty: 12, AverageCost: 475.10, LatestPrice: 529.42, MarketValue: 6353.04, UnrealizedPnl: 651.84, UnrealizedPnlPercent: 11.43}, {Symbol: "AAPL", Name: "Apple Inc.", Market: "US", Currency: "USD", PositionQty: 18, AverageCost: 174.20, LatestPrice: 214.40, MarketValue: 3859.20, UnrealizedPnl: 723.60, UnrealizedPnlPercent: 23.08}, {Symbol: "D05", Name: "DBS Group", Market: "SG", Currency: "USD", PositionQty: 100, AverageCost: 34.80, LatestPrice: 38.50, MarketValue: 2850, UnrealizedPnl: 274, UnrealizedPnlPercent: 10.63}, {Symbol: "0700", Name: "Tencent Holdings", Market: "HK", Currency: "USD", PositionQty: 40, AverageCost: 310, LatestPrice: 375, MarketValue: 1923, UnrealizedPnl: 333, UnrealizedPnlPercent: 20.97}}
	return buildSnapshot(positions, "demo", "USD", map[string]float64{"USD": 1})
}
