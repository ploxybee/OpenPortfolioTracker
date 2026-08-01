package migration

import "testing"

func TestMigrationFilesAreVersionOrdered(t *testing.T) {
	entries, err := migrationFiles()
	if err != nil {
		t.Fatalf("list migration files: %v", err)
	}
	if len(entries) != 2 || entries[0].Name() != "000001_create_portfolio_snapshots.sql" || entries[1].Name() != "000002_create_portfolio_holding_snapshots.sql" {
		t.Fatalf("unexpected migration files: %#v", entries)
	}
}
