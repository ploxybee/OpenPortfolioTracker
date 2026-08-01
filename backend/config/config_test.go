package config

import "testing"

func TestLoadAppliesDefaults(t *testing.T) {
	config, err := load(func(name string) string {
		if name == "DATABASE_URL" {
			return "postgres://portfolio:password@localhost:5432/portfolio"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Port != "8080" || config.Tiger.BaseCurrency != "SGD" || config.Portfolio.Targets["Equities"] != 1 || config.Portfolio.RebalanceTolerance != 5 {
		t.Fatalf("unexpected defaults: %#v", config)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	if _, err := load(func(string) string { return "" }); err == nil {
		t.Fatal("expected DATABASE_URL validation error")
	}
}

func TestLoadReadsPortfolioConfiguration(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":                   "postgres://portfolio:password@localhost:5432/portfolio",
		"PORT":                           "9090",
		"PORTFOLIO_BASE_CURRENCY":        "SGD",
		"PORTFOLIO_TARGETS":              "Equities=70,Bonds=30",
		"PORTFOLIO_REBALANCE_TOLERANCE":  "3",
		"PORTFOLIO_MONTHLY_CONTRIBUTION": "1000",
	}
	config, err := load(func(name string) string { return values[name] })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Port != "9090" || config.Tiger.BaseCurrency != "SGD" || config.Portfolio.Targets["Equities"] != 0.7 || config.Portfolio.MonthlyContribution != 1000 || config.Portfolio.RebalanceTolerance != 3 {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestLoadRejectsNonSGDBaseCurrency(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":            "postgres://portfolio:password@localhost:5432/portfolio",
		"PORTFOLIO_BASE_CURRENCY": "USD",
	}
	if _, err := load(func(name string) string { return values[name] }); err == nil {
		t.Fatal("expected non-SGD base currency validation error")
	}
}

func TestLoadConfiguresIBKRLoopbackGateway(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":         "postgres://portfolio:password@localhost:5432/portfolio",
		"IBKR_ENABLED":         "true",
		"IBKR_GATEWAY_CA_FILE": "/etc/ibkr/ca.pem",
	}
	config, err := load(func(name string) string { return values[name] })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !config.IBKR.Enabled || config.IBKR.GatewayURL != "https://localhost:5000/v1/api" || config.IBKR.CAFile != "/etc/ibkr/ca.pem" {
		t.Fatalf("unexpected IBKR config: %#v", config.IBKR)
	}
}

func TestLoadRejectsNonLoopbackIBKRGateway(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":         "postgres://portfolio:password@localhost:5432/portfolio",
		"IBKR_ENABLED":         "true",
		"IBKR_GATEWAY_URL":     "https://gateway.example.com/v1/api",
		"IBKR_GATEWAY_CA_FILE": "/etc/ibkr/ca.pem",
	}
	if _, err := load(func(name string) string { return values[name] }); err == nil {
		t.Fatal("expected non-loopback gateway rejection")
	}
}
