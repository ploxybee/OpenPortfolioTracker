// Package migration applies the application's embedded PostgreSQL schema migrations.
package migration

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var files embed.FS

// Apply runs every unapplied migration in filename order.
func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}
	entries, err := migrationFiles()
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := applyFile(ctx, pool, entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

func migrationFiles() ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(files, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, nil
}

func applyFile(ctx context.Context, pool *pgxpool.Pool, name string) error {
	var applied bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&applied); err != nil {
		return fmt.Errorf("check migration %s: %w", name, err)
	}
	if applied {
		return nil
	}
	sql, err := files.ReadFile("migrations/" + name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}
