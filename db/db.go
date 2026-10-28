package db

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// EnsureDatabase checks if the target database exists; if not, connects to default 'postgres' database and creates it.
func EnsureDatabase(ctx context.Context, connString string) error {
	u, err := url.Parse(connString)
	if err != nil {
		return fmt.Errorf("error parsing database URL: %w", err)
	}

	targetDB := strings.TrimPrefix(u.Path, "/")
	if targetDB == "" || targetDB == "postgres" {
		return nil // standard default database
	}

	// Try pinging the target database first
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("error parsing pool config: %w", err)
	}

	testPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err == nil {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = testPool.Ping(pingCtx)
		cancel()
		testPool.Close()
		if err == nil {
			return nil // database already exists and is reachable
		}
	}

	// If database doesn't exist (or connection failed due to missing DB), connect to 'postgres' DB
	adminURL := *u
	adminURL.Path = "/postgres"

	adminConn, err := pgx.Connect(ctx, adminURL.String())
	if err != nil {
		// If connecting to /postgres fails, return original ping error
		return fmt.Errorf("unable to connect to postgres server: %w", err)
	}
	defer adminConn.Close(ctx)

	// Check if target database exists
	var exists bool
	err = adminConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", targetDB).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		// Database name cannot be parameterized in CREATE DATABASE; sanitize targetDB name
		safeDB := strings.ReplaceAll(targetDB, "\"", "")
		_, err = adminConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", safeDB))
		if err != nil {
			var pgErr *pgconn.PgError
			if !(errorsAs(err, &pgErr) && pgErr.Code == "42P04") { // duplicate_database
				return fmt.Errorf("failed to create database %q: %w", safeDB, err)
			}
		}
	}

	return nil
}

func errorsAs(err error, target any) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if t, ok := target.(**pgconn.PgError); ok {
			*t = pgErr
			return true
		}
	}
	return false
}

// NewPool creates a verified pgx connection pool.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	// Automatically ensure database exists first
	if err := EnsureDatabase(ctx, connString); err != nil {
		// Log warning but continue, as target might already be handling permissions
		_ = err
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

// RunMigrations applies all pending migrations from the embedded filesystem.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// 1. Ensure migrations tracking table exists
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Read all embedded migration files
	entries, err := MigrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// 3. Query already applied migrations
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("failed to scan migration row: %w", err)
		}
		applied[v] = true
	}

	// 4. Run pending migrations in individual transactions
	for _, file := range files {
		if applied[file] {
			continue
		}

		content, err := MigrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migration %s failed: %w", file, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}
	}

	return nil
}
