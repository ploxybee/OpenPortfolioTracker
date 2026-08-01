package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ploxybee/OpenPortfolioTracker/config"
	"github.com/ploxybee/OpenPortfolioTracker/internal/migration"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
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
	if err := migration.Apply(ctx, pool); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Print("database migrations are up to date")
}
