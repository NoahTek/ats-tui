package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Organisation represents the Organisations (Clients) sheet (Sheet 4).
type Organisation struct {
	ID          int
	OrgName     string
	OrgCode     string
	Industry    string
	SubCategory string
}

// FetchOrganisations queries client organisations joined with their Industry and Sub-Category hierarchy.
func FetchOrganisations(ctx context.Context, pool *pgxpool.Pool) ([]Organisation, error) {
	query := `
		SELECT 
			c.id,
			c.org_name,
			c.org_code,
			COALESCE(parent.name, ind.name, 'Unassigned') AS industry,
			COALESCE(CASE WHEN parent.name IS NOT NULL THEN ind.name ELSE 'General' END, 'General') AS sub_category
		FROM clients c
		LEFT JOIN industry ind ON c.industry_id = ind.id
		LEFT JOIN industry parent ON ind.parent_id = parent.id
		ORDER BY c.org_name ASC;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organisations: %w", err)
	}
	defer rows.Close()

	var list []Organisation
	for rows.Next() {
		var o Organisation
		err := rows.Scan(
			&o.ID,
			&o.OrgName,
			&o.OrgCode,
			&o.Industry,
			&o.SubCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("scan organisation failed: %w", err)
		}
		list = append(list, o)
	}

	return list, nil
}

// CreateOrganisation inserts a new client organisation into PostgreSQL.
func CreateOrganisation(ctx context.Context, pool *pgxpool.Pool, o Organisation) error {
	indID, err := FindOrCreateIndustryHierarchy(ctx, pool, o.Industry, o.SubCategory)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO clients (org_name, org_code, industry_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_code) DO UPDATE 
		SET org_name = EXCLUDED.org_name, industry_id = EXCLUDED.industry_id;
	`
	_, err = pool.Exec(ctx, query, o.OrgName, o.OrgCode, indID)
	if err != nil {
		return fmt.Errorf("failed to create organisation: %w", err)
	}
	return nil
}

// UpdateOrganisation updates an existing client organisation.
func UpdateOrganisation(ctx context.Context, pool *pgxpool.Pool, o Organisation) error {
	indID, err := FindOrCreateIndustryHierarchy(ctx, pool, o.Industry, o.SubCategory)
	if err != nil {
		return err
	}

	query := `
		UPDATE clients
		SET org_name = $1,
		    industry_id = $2
		WHERE id = $3 OR org_code = $4;
	`
	_, err = pool.Exec(ctx, query, o.OrgName, indID, o.ID, o.OrgCode)
	if err != nil {
		return fmt.Errorf("failed to update organisation %s: %w", o.OrgCode, err)
	}
	return nil
}

// DeleteOrganisation removes a client organisation by ID.
func DeleteOrganisation(ctx context.Context, pool *pgxpool.Pool, id int) error {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE client_id = $1;`, id).Scan(&count)
	if err == nil && count > 0 {
		return fmt.Errorf("cannot delete organisation: %d job(s) are assigned to it. Please reassign or delete the jobs first", count)
	}

	query := `DELETE FROM clients WHERE id = $1;`
	tag, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organisation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("organisation with ID %d not found", id)
	}
	return nil
}
