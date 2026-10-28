package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"ats-tui/db"
	"ats-tui/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const defaultDBURI = "postgres://postgres:postgres@localhost:5434/ats"

func resolveDBURI() string {
	var flagURI string
	flag.StringVar(&flagURI, "db", "", "PostgreSQL connection string (e.g., postgres://user:pass@host:port/dbname)")
	flag.Parse()

	if flagURI != "" {
		return flagURI
	}

	if env := os.Getenv("DATABASE_URL"); env != "" {
		return env
	}
	if env := os.Getenv("ATS_DB_URI"); env != "" {
		return env
	}

	return defaultDBURI
}

func main() {
	ctx := context.Background()
	connStr := resolveDBURI()

	pool, err := db.NewPool(ctx, connStr)
	if err != nil {
		fmt.Printf("\n❌ Database connection failed:\n  %v\n\n", err)
		fmt.Printf("Please ensure PostgreSQL is running and supply connection details via:\n")
		fmt.Printf("  • CLI flag:        ./ats-tui -db \"postgres://user:password@localhost:5432/ats\"\n")
		fmt.Printf("  • Environment var: export DATABASE_URL=\"postgres://user:password@localhost:5432/ats\"\n\n")
		os.Exit(1)
	}
	defer pool.Close()

	// Automatically apply embedded database migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		fmt.Printf("\n❌ Database migration failed:\n  %v\n\n", err)
		os.Exit(1)
	}

	app := ui.NewApp(pool)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ATS TUI: %v\n", err)
		os.Exit(1)
	}
}
