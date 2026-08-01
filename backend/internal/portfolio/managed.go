package portfolio

import (
	"context"

	"github.com/ploxybee/OpenPortfolioTracker/config"
)

// ManagedProvider adds user-configured allocation targets and contribution-only
// guidance to any read-only broker provider. It does not create orders.
type ManagedProvider struct {
	provider     Provider
	targets      map[string]float64
	tolerance    float64
	contribution float64
}

// NewManagedProvider adds allocation guidance to a broker provider.
func NewManagedProvider(provider Provider, config config.PortfolioConfig) *ManagedProvider {
	return &ManagedProvider{provider: provider, targets: config.Targets, tolerance: config.RebalanceTolerance, contribution: config.MonthlyContribution}
}

// Snapshot enriches a broker snapshot with targets and contribution guidance.
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
