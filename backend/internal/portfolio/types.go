package portfolio

import (
	"context"
	"time"
)

type Holding struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	Market          string  `json:"market"`
	Country         string  `json:"country"`
	Currency        string  `json:"currency"`
	NativeCurrency  string  `json:"nativeCurrency"`
	Quantity        float64 `json:"quantity"`
	AverageCost     float64 `json:"averageCost"`
	LatestPrice     float64 `json:"latestPrice"`
	MarketValue     float64 `json:"marketValue"`
	UnrealizedPnL   float64 `json:"unrealizedPnl"`
	UnrealizedPnLPc float64 `json:"unrealizedPnlPercent"`
	Weight          float64 `json:"weight"`
}

type Allocation struct {
	Label      string  `json:"label"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
}

type Snapshot struct {
	Broker            string       `json:"broker"`
	Mode              string       `json:"mode"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	TotalValue        float64      `json:"totalValue"`
	Currency          string       `json:"currency"`
	Holdings          []Holding    `json:"holdings"`
	CountryAllocation []Allocation `json:"countryAllocation"`
	MarketAllocation  []Allocation `json:"marketAllocation"`
}

type Provider interface {
	Snapshot(ctx context.Context) (Snapshot, error)
}
