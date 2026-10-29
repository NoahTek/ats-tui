package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// JobPosting represents the Jobs (Open Positions) sheet (Sheet 3).
type JobPosting struct {
	ID              int
	JobCode         string
	ClientName      string
	JobTitle        string
	Role            string
	Department      string
	SubDepartment   string
	Location        string
	RequisitionDate string
	NumPositions    int
	NumShared       int
}

// FetchJobPostings queries all open job requisitions and computes live count of profiles shared.
func FetchJobPostings(ctx context.Context, pool *pgxpool.Pool) ([]JobPosting, error) {
	query := `
		SELECT 
			j.id,
			COALESCE(j.job_code, 'JOB' || j.id),
			COALESCE(c.org_name, 'Unassigned'),
			j.job_title,
			COALESCE(j.role, ''),
			COALESCE(j.department, ''),
			COALESCE(j.sub_department, ''),
			COALESCE(l.location, 'Unassigned'),
			COALESCE(to_char(j.requisition_date, 'YYYY-MM-DD'), ''),
			COALESCE(j.num_positions, 1),
			COUNT(a.id) AS num_shared
		FROM jobs j
		LEFT JOIN clients c ON j.client_id = c.id
		LEFT JOIN location l ON j.location_id = l.id
		LEFT JOIN applications a ON j.id = a.job_id
		GROUP BY j.id, c.org_name, l.location
		ORDER BY j.created_at DESC;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job postings: %w", err)
	}
	defer rows.Close()

	var jobs []JobPosting
	for rows.Next() {
		var jp JobPosting
		err := rows.Scan(
			&jp.ID,
			&jp.JobCode,
			&jp.ClientName,
			&jp.JobTitle,
			&jp.Role,
			&jp.Department,
			&jp.SubDepartment,
			&jp.Location,
			&jp.RequisitionDate,
			&jp.NumPositions,
			&jp.NumShared,
		)
		if err != nil {
			return nil, fmt.Errorf("scan job posting failed: %w", err)
		}
		jobs = append(jobs, jp)
	}

	return jobs, nil
}

// CreateJob inserts a new job requisition into PostgreSQL.
func CreateJob(ctx context.Context, pool *pgxpool.Pool, j JobPosting) error {
	locID, err := FindOrCreateLocation(ctx, pool, j.Location)
	if err != nil {
		return err
	}

	var clientID *int
	var cid int
	err = pool.QueryRow(ctx, `SELECT id FROM clients WHERE org_code = $1 OR org_name = $1 LIMIT 1;`, j.ClientName).Scan(&cid)
	if err == nil && cid > 0 {
		clientID = &cid
	}

	numPos := j.NumPositions
	if numPos <= 0 {
		numPos = 1
	}

	query := `
		INSERT INTO jobs (job_code, client_id, job_title, role, department, sub_department, location_id, num_positions, requisition_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, '')::date);
	`
	_, err = pool.Exec(ctx, query,
		j.JobCode, clientID, j.JobTitle, j.Role, j.Department, j.SubDepartment, locID, numPos, j.RequisitionDate,
	)
	if err != nil {
		return fmt.Errorf("failed to create job posting: %w", err)
	}
	return nil
}

// UpdateJob updates an existing job requisition.
func UpdateJob(ctx context.Context, pool *pgxpool.Pool, j JobPosting) error {
	locID, err := FindOrCreateLocation(ctx, pool, j.Location)
	if err != nil {
		return err
	}

	var clientID *int
	var cid int
	err = pool.QueryRow(ctx, `SELECT id FROM clients WHERE org_code = $1 OR org_name = $1 LIMIT 1;`, j.ClientName).Scan(&cid)
	if err == nil && cid > 0 {
		clientID = &cid
	}

	numPos := j.NumPositions
	if numPos <= 0 {
		numPos = 1
	}

	query := `
		UPDATE jobs
		SET job_title = $1,
		    role = $2,
		    department = $3,
		    sub_department = $4,
		    location_id = $5,
		    num_positions = $6,
		    client_id = $7
		WHERE id = $8 OR job_code = $9;
	`
	_, err = pool.Exec(ctx, query,
		j.JobTitle, j.Role, j.Department, j.SubDepartment, locID, numPos, clientID, j.ID, j.JobCode,
	)
	if err != nil {
		return fmt.Errorf("failed to update job posting: %w", err)
	}
	return nil
}

// DeleteJob removes a job requisition by ID.
func DeleteJob(ctx context.Context, pool *pgxpool.Pool, id int) error {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE job_id = $1;`, id).Scan(&count)
	if err == nil && count > 0 {
		return fmt.Errorf("cannot delete job: %d application(s) are linked to it. Please delete or reassign them first", count)
	}

	query := `DELETE FROM jobs WHERE id = $1;`
	tag, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job posting: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job posting with ID %d not found", id)
	}
	return nil
}

// DeleteJobs removes multiple job requisitions by their IDs.
func DeleteJobs(ctx context.Context, pool *pgxpool.Pool, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE job_id = ANY($1);`, ids).Scan(&count)
	if err == nil && count > 0 {
		return fmt.Errorf("cannot delete jobs: %d application(s) are linked to them. Please delete or reassign them first", count)
	}

	query := `DELETE FROM jobs WHERE id = ANY($1);`
	tag, err := pool.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("failed to delete job postings: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no job postings found to delete")
	}
	return nil
}
