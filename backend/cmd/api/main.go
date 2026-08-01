package main

import (
	"log"
	"net/http"
	"time"

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
	provider := portfolio.NewManagedProvider(portfolio.NewTigerProvider(appConfig.Tiger), appConfig.Portfolio)
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
