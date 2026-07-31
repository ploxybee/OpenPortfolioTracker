package portfolio

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

// ManagedProvider adds user-configured allocation targets and contribution-only
// guidance to any read-only broker provider. It does not create orders.
type ManagedProvider struct {
	provider     Provider
	targets      map[string]float64
	tolerance    float64
	contribution float64
}

func NewManagedProviderFromEnv(provider Provider) (*ManagedProvider, error) {
	targets, err := parseTargets(os.Getenv("PORTFOLIO_TARGETS"))
	if err != nil {
		return nil, err
	}
	tolerance, err := optionalFloatEnv("PORTFOLIO_REBALANCE_TOLERANCE", 5)
	if err != nil || tolerance < 0 {
		return nil, fmt.Errorf("PORTFOLIO_REBALANCE_TOLERANCE must be a non-negative percentage")
	}
	contribution, err := optionalFloatEnv("PORTFOLIO_MONTHLY_CONTRIBUTION", 0)
	if err != nil || contribution < 0 {
		return nil, fmt.Errorf("PORTFOLIO_MONTHLY_CONTRIBUTION must be a non-negative amount")
	}
	return &ManagedProvider{provider: provider, targets: targets, tolerance: tolerance, contribution: contribution}, nil
}

func optionalFloatEnv(name string, fallback float64) (float64, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(value, 64)
}

func (p *ManagedProvider) Snapshot(ctx context.Context) (Snapshot, error) {
	snapshot, err := p.provider.Snapshot(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Targets = p.targets
	snapshot.RebalanceTolerance = p.tolerance
	snapshot.MonthlyContribution = p.contribution
	snapshot.AssetAllocation = targetAllocations(snapshot.Holdings, p.targets)
	snapshot.ContributionPlan = contributionPlan(snapshot.TotalValue, snapshot.AssetAllocation, p.contribution)
	return snapshot, nil
}
