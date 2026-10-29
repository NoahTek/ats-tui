package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seed.sql
var SeedSQL string

// IsDatabaseEmpty returns true if there are zero candidates in the database.
func IsDatabaseEmpty(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var count int
	err := pool.QueryRow(ctx, "SELECT count(*) FROM candidates").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// SeedSampleData executes the embedded seed.sql script to load sample data.
func SeedSampleData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, SeedSQL)
	if err != nil {
		return fmt.Errorf("failed to seed sample data: %w", err)
	}
	return nil
}
