package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FindOrCreateLocation finds an existing location ID or creates a new entry on-the-fly.
func FindOrCreateLocation(ctx context.Context, pool *pgxpool.Pool, name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Remote"
	}

	var id int
	query := `
		INSERT INTO location (location)
		VALUES ($1)
		ON CONFLICT (location) DO UPDATE SET location = EXCLUDED.location
		RETURNING id;
	`
	err := pool.QueryRow(ctx, query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to find or create location '%s': %w", name, err)
	}
	return id, nil
}

// FindOrCreateQualification finds an existing qualification ID or creates a new entry on-the-fly.
func FindOrCreateQualification(ctx context.Context, pool *pgxpool.Pool, name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "General / Other"
	}

	var id int
	query := `
		INSERT INTO educational_qualification (qualification)
		VALUES ($1)
		ON CONFLICT (qualification) DO UPDATE SET qualification = EXCLUDED.qualification
		RETURNING id;
	`
	err := pool.QueryRow(ctx, query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to find or create qualification '%s': %w", name, err)
	}
	return id, nil
}

// FindOrCreateProduct finds an existing product ID or creates a new entry.
func FindOrCreateProduct(ctx context.Context, pool *pgxpool.Pool, name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, nil
	}

	var id int
	query := `
		INSERT INTO product (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id;
	`
	err := pool.QueryRow(ctx, query, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to find or create product '%s': %w", name, err)
	}
	return id, nil
}

// FindOrCreateIndustryHierarchy finds or creates parent industry and child subcategory.
func FindOrCreateIndustryHierarchy(ctx context.Context, pool *pgxpool.Pool, indName, subCatName string) (int, error) {
	indName = strings.TrimSpace(indName)
	if indName == "" {
		indName = "General Technology"
	}
	subCatName = strings.TrimSpace(subCatName)
	if subCatName == "" {
		subCatName = "General"
	}

	// 1. Find or create top-level parent industry
	var parentID int
	err := pool.QueryRow(ctx, `SELECT id FROM industry WHERE name = $1 AND parent_id IS NULL LIMIT 1;`, indName).Scan(&parentID)
	if err != nil {
		// insert parent
		err = pool.QueryRow(ctx, `INSERT INTO industry (name, parent_id) VALUES ($1, NULL) RETURNING id;`, indName).Scan(&parentID)
		if err != nil {
			return 0, fmt.Errorf("failed to create parent industry '%s': %w", indName, err)
		}
	}

	// 2. Find or create sub-category linked to parent
	var subCatID int
	err = pool.QueryRow(ctx, `SELECT id FROM industry WHERE name = $1 AND parent_id = $2 LIMIT 1;`, subCatName, parentID).Scan(&subCatID)
	if err != nil {
		err = pool.QueryRow(ctx, `INSERT INTO industry (name, parent_id) VALUES ($1, $2) RETURNING id;`, subCatName, parentID).Scan(&subCatID)
		if err != nil {
			return 0, fmt.Errorf("failed to create subcategory '%s': %w", subCatName, err)
		}
	}

	return subCatID, nil
}
