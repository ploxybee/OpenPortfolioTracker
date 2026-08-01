package config

import "testing"

func TestLoadAppliesDefaults(t *testing.T) {
	config, err := load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Port != "8080" || config.Tiger.BaseCurrency != "USD" || config.Portfolio.Targets["Equities"] != 1 || config.Portfolio.RebalanceTolerance != 5 {
		t.Fatalf("unexpected defaults: %#v", config)
	}
}

func TestLoadReadsPortfolioConfiguration(t *testing.T) {
	values := map[string]string{
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
