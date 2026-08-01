package portfolio

import (
	"math"
	"sort"
)

// targetAllocations compares current holdings with their target allocations.
func targetAllocations(holdings []Holding, targets map[string]float64) []Allocation {
	values := map[string]float64{}
	total := 0.0
	for _, holding := range holdings {
		values[holding.AssetClass] += holding.MarketValue
		total += holding.MarketValue
	}
	for assetClass := range targets {
		if _, ok := values[assetClass]; !ok {
			values[assetClass] = 0
		}
	}
	out := make([]Allocation, 0, len(values))
	for assetClass, value := range values {
		percentage := 0.0
		if total > 0 {
			percentage = value / total * 100
		}
		target := targets[assetClass] * 100
		out = append(out, Allocation{Label: assetClass, Value: value, Percentage: percentage, TargetPercentage: target, DriftPercentage: percentage - target})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value > out[j].Value })
	return out
}

// contributionPlan directs new funds to the largest target shortfalls first.
func contributionPlan(totalValue float64, allocations []Allocation, contribution float64) []ContributionSuggestion {
	if contribution <= 0 {
		return []ContributionSuggestion{}
	}
	futureValue := totalValue + contribution
	remaining := contribution
	plan := []ContributionSuggestion{}
	sort.Slice(allocations, func(i, j int) bool {
		return futureValue*allocations[i].TargetPercentage/100-allocations[i].Value > futureValue*allocations[j].TargetPercentage/100-allocations[j].Value
	})
	for _, allocation := range allocations {
		shortfall := futureValue*allocation.TargetPercentage/100 - allocation.Value
		if shortfall <= 0 || remaining <= 0 {
			continue
		}
		amount := math.Min(remaining, shortfall)
		plan = append(plan, ContributionSuggestion{AssetClass: allocation.Label, Action: "Buy", Amount: math.Round(amount*100) / 100})
		remaining -= amount
	}
	return plan
}
