package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ploxybee/OpenPortfolioTracker/config"
	"github.com/ploxybee/OpenPortfolioTracker/internal/api"
	"github.com/ploxybee/OpenPortfolioTracker/internal/portfolio"
)

// main loads configuration and starts the portfolio API.
func main() {
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("invalid portfolio configuration: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, appConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping PostgreSQL: %v", err)
	}

	tigerProvider := portfolio.NewTigerProvider(appConfig.Tiger)
	liveProviders := make([]portfolio.Provider, 0, 2)
	if tigerProvider.Configured() {
		liveProviders = append(liveProviders, tigerProvider)
	}
	if appConfig.IBKR.Enabled {
		ibkrProvider, err := portfolio.NewIBKRProvider(appConfig.IBKR)
		if err != nil {
			log.Fatalf("configure IBKR Gateway client: %v", err)
		}
		liveProviders = append(liveProviders, ibkrProvider)
	}
	var brokerProvider portfolio.Provider
	if len(liveProviders) == 0 {
		brokerProvider = tigerProvider // intentional local development demo
	} else {
		brokerProvider = portfolio.NewCombinedProvider("SGD", liveProviders...)
	}
	liveProvider := portfolio.NewManagedProvider(brokerProvider, appConfig.Portfolio)
	provider := portfolio.NewCachingProvider(liveProvider, portfolio.NewPostgresSnapshotStore(pool))
	server := api.NewServer(provider)
	httpServer := &http.Server{
		Addr:              ":" + appConfig.Port,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("portfolio API listening on http://localhost:%s", appConfig.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
