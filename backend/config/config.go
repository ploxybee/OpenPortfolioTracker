// Package config loads and validates application configuration.
package config

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	Tiger     TigerConfig
	Portfolio PortfolioConfig
}

type TigerConfig struct {
	ID           string
	PrivateKey   string
	Account      string
	License      string
	BaseCurrency string
}

type PortfolioConfig struct {
	Targets             map[string]float64
	RebalanceTolerance  float64
	MonthlyContribution float64
}

// Load reads backend/.env when present, then validates the environment.
func Load() (Config, error) {
	// A missing .env is normal in deployed environments and activates demo mode locally.
	_ = godotenv.Load()
	return load(os.Getenv)
}

// load builds validated configuration from an environment lookup function.
func load(getenv func(string) string) (Config, error) {
	targets, err := parseTargets(getenv("PORTFOLIO_TARGETS"))
	if err != nil {
		return Config{}, err
	}
	tolerance, err := optionalFloat(getenv("PORTFOLIO_REBALANCE_TOLERANCE"), 5)
	if err != nil || tolerance < 0 {
		return Config{}, fmt.Errorf("PORTFOLIO_REBALANCE_TOLERANCE must be a non-negative percentage")
	}
	contribution, err := optionalFloat(getenv("PORTFOLIO_MONTHLY_CONTRIBUTION"), 0)
	if err != nil || contribution < 0 {
		return Config{}, fmt.Errorf("PORTFOLIO_MONTHLY_CONTRIBUTION must be a non-negative amount")
	}

	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	baseCurrency := getenv("PORTFOLIO_BASE_CURRENCY")
	if baseCurrency == "" {
		baseCurrency = "USD"
	}

	return Config{
		Port: port,
		Tiger: TigerConfig{
			ID:           getenv("TIGER_ID"),
			PrivateKey:   getenv("TIGER_PRIVATE_KEY"),
			Account:      getenv("TIGER_ACCOUNT"),
			License:      getenv("TIGER_LICENSE"),
			BaseCurrency: baseCurrency,
		},
		Portfolio: PortfolioConfig{
			Targets:             targets,
			RebalanceTolerance:  tolerance,
			MonthlyContribution: contribution,
		},
	}, nil
}

// optionalFloat parses a value or returns its default when unset.
func optionalFloat(value string, fallback float64) (float64, error) {
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(value, 64)
}

// parseTargets converts comma-separated percentage targets into decimal weights.
func parseTargets(value string) (map[string]float64, error) {
	if strings.TrimSpace(value) == "" {
		return map[string]float64{"Equities": 1}, nil
	}
	targets := map[string]float64{}
	for _, pair := range strings.Split(value, ",") {
		key, raw, ok := strings.Cut(pair, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("invalid PORTFOLIO_TARGETS entry %q", pair)
		}
		percentage, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil || percentage < 0 {
			return nil, fmt.Errorf("invalid target percentage for %s", key)
		}
		targets[strings.TrimSpace(key)] = percentage / 100
	}
	sum := 0.0
	for _, target := range targets {
		sum += target
	}
	if math.Abs(sum-1) > 0.0001 {
		return nil, fmt.Errorf("PORTFOLIO_TARGETS must total 100%%; received %.2f%%", sum*100)
	}
	return targets, nil
}
