package db

import (
	"context"
	"os"
	"testing"
)

func TestMigrations(t *testing.T) {
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5434/ats_test"
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, connStr)
	if err != nil {
		t.Skipf("Skipping migration test, unable to connect to %s: %v", connStr, err)
		return
	}
	defer pool.Close()

	if err := RunMigrations(ctx, pool); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify tables exist
	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query information_schema: %v", err)
	}

	if count < 8 {
		t.Errorf("Expected at least 8 tables created, got %d", count)
	}

	// Run migrations again to verify idempotency
	if err := RunMigrations(ctx, pool); err != nil {
		t.Fatalf("Idempotency failed on second RunMigrations: %v", err)
	}
}
