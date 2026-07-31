package portfolio

import "testing"

func TestBuildTargetAllocationsAndContributionPlan(t *testing.T) {
	holdings := []Holding{
		{Symbol: "WORLD", AssetClass: "Equities", MarketValue: 7000},
		{Symbol: "BOND", AssetClass: "Bonds", MarketValue: 2000},
		{Symbol: "CASH", AssetClass: "Cash", MarketValue: 1000},
	}
	targets := map[string]float64{"Equities": 0.6, "Bonds": 0.3, "Cash": 0.1}

	allocations := targetAllocations(holdings, targets)
	if allocations[0].Label != "Equities" || allocations[0].Percentage != 70 || allocations[0].TargetPercentage != 60 || allocations[0].DriftPercentage != 10 {
		t.Fatalf("unexpected target allocation: %#v", allocations[0])
	}

	plan := contributionPlan(10_000, allocations, 1000)
	if len(plan) != 1 || plan[0].AssetClass != "Bonds" || plan[0].Amount != 1000 {
		t.Fatalf("unexpected contribution plan: %#v", plan)
	}
}

func TestParseTargetsRejectsInvalidTotal(t *testing.T) {
	if _, err := parseTargets("Equities=80,Bonds=10"); err == nil {
		t.Fatal("expected invalid targets to be rejected")
	}
}
