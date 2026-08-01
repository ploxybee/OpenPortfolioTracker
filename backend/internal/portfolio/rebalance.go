package portfolio

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

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
