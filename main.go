package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ats-tui/db"
	"ats-tui/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPort5434URI = "postgres://postgres:postgres@localhost:5434/ats"
	defaultPort5432URI = "postgres://postgres:postgres@localhost:5432/ats"
)

func main() {
	var (
		flagURI   string
		flagSetup bool
	)

	flag.StringVar(&flagURI, "db", "", "PostgreSQL connection string (e.g., postgres://user:pass@host:port/dbname)")
	flag.BoolVar(&flagSetup, "setup", false, "Open In-App Setup Wizard to configure database settings")
	flag.Parse()

	// 1. Try to load saved config from ~/.config/ats-tui/config.json
	savedCfg, _ := db.LoadConfig()

	// If explicit --setup requested, launch Setup Wizard directly
	if flagSetup {
		root := ui.NewRootModel(nil, savedCfg, true)
		p := tea.NewProgram(root, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running ATS TUI setup: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 2. Determine connection candidate list in priority order
	var candidateURIs []string

	if flagURI != "" {
		candidateURIs = []string{flagURI}
	} else if env := os.Getenv("DATABASE_URL"); env != "" {
		candidateURIs = []string{env}
	} else if env := os.Getenv("ATS_DB_URI"); env != "" {
		candidateURIs = []string{env}
	} else if savedCfg != nil {
		candidateURIs = []string{savedCfg.ConnectionURI()}
	} else {
		candidateURIs = []string{defaultPort5434URI, defaultPort5432URI}
	}

	// 3. Attempt quiet connection to candidate URIs
	var pool *pgxpool.Pool
	for _, uri := range candidateURIs {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		p, err := db.NewPool(ctx, uri)
		cancel()
		if err == nil {
			pool = p
			break
		}
	}

	// 4. If connection succeeded, ensure migrations and initial bootstrap
	if pool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.RunMigrations(ctx, pool); err == nil {
			// If database is completely empty and user hasn't explicitly disabled demo seeding
			if savedCfg == nil || savedCfg.SeedDemo {
				if empty, err := db.IsDatabaseEmpty(ctx, pool); err == nil && empty {
					_ = db.SeedSampleData(ctx, pool)
				}
			}

			// Launch main application directly
			root := ui.NewRootModel(pool, savedCfg, false)
			p := tea.NewProgram(root, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running ATS TUI: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// If migrations failed, close pool and fall through to Setup Wizard
		pool.Close()
		pool = nil
	}

	// 5. If no connection succeeded, seamlessly launch the In-App Setup Wizard!
	root := ui.NewRootModel(nil, savedCfg, true)
	p := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ATS TUI: %v\n", err)
		os.Exit(1)
	}
}
