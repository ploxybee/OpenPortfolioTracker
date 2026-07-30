package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/ploxybee/OpenPortfolioTracker/internal/api"
	"github.com/ploxybee/OpenPortfolioTracker/internal/portfolio"
)

func main() {
	// A missing .env is normal in deployed environments and activates demo mode locally.
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	provider := portfolio.NewTigerProviderFromEnv()
	server := api.NewServer(provider)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("portfolio API listening on http://localhost:%s", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
