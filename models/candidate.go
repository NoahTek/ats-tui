package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Candidate struct {
	ID       int
	Name     string
	Email    string
	Phone    string
	Location string
}

// CandidateProfile represents the Master Candidate Pool (Profiles View - Sheet 1).
type CandidateProfile struct {
	ID            int
	CandidateCode string
	Name          string
	Phone         string
	Email         string
	Location      string
	Qualification string
	DOB           string
	ResumeText    string
}

// FetchCandidateProfiles queries candidates joined with their resolved Location and Qualification names.
func FetchCandidateProfiles(ctx context.Context, pool *pgxpool.Pool, searchQuery string) ([]CandidateProfile, error) {
	var query strings.Builder
	var args []any

	query.WriteString(`
		SELECT 
			c.id,
			COALESCE(c.candidate_code, 'CAN-' || LPAD(c.id::text, 3, '0')) AS candidate_code,
			c.name,
			COALESCE(c.phone, '') AS phone,
			c.email,
			COALESCE(l.location, 'Unassigned') AS location,
			COALESCE(eq.qualification, 'Unassigned') AS qualification,
			COALESCE(to_char(c.dob, 'YYYY-MM-DD'), '') AS dob,
			COALESCE(c.resume_text, '') AS resume_text
		FROM candidates c
		LEFT JOIN location l ON c.location_id = l.id
		LEFT JOIN educational_qualification eq ON c.qualification_id = eq.id
	`)

	searchQuery = strings.TrimSpace(searchQuery)
	if searchQuery != "" {
		query.WriteString(`
			WHERE c.search_vector @@ plainto_tsquery('english', $1)
			   OR c.name ILIKE '%' || $1 || '%'
			   OR c.email ILIKE '%' || $1 || '%'
			   OR c.candidate_code ILIKE '%' || $1 || '%'
			ORDER BY ts_rank(c.search_vector, plainto_tsquery('english', $1)) DESC, c.id DESC;
		`)
		args = append(args, searchQuery)
	} else {
		query.WriteString(` ORDER BY c.id DESC; `)
	}

	rows, err := pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candidate profiles: %w", err)
	}
	defer rows.Close()

	var profiles []CandidateProfile
	for rows.Next() {
		var p CandidateProfile
		err := rows.Scan(
			&p.ID,
			&p.CandidateCode,
			&p.Name,
			&p.Phone,
			&p.Email,
			&p.Location,
			&p.Qualification,
			&p.DOB,
			&p.ResumeText,
		)
		if err != nil {
			return nil, fmt.Errorf("scan candidate profile failed: %w", err)
		}
		profiles = append(profiles, p)
	}

	return profiles, nil
}

// CreateCandidate inserts a new candidate into PostgreSQL, automatically assigning a CAN-xxx code.
func CreateCandidate(ctx context.Context, pool *pgxpool.Pool, p CandidateProfile) (string, error) {
	locID, err := FindOrCreateLocation(ctx, pool, p.Location)
	if err != nil {
		return "", err
	}

	qualID, err := FindOrCreateQualification(ctx, pool, p.Qualification)
	if err != nil {
		return "", err
	}

	var code string
	if strings.TrimSpace(p.CandidateCode) != "" {
		query := `
			INSERT INTO candidates (candidate_code, name, email, phone, location_id, qualification_id, dob, resume_text)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::date, $8)
			RETURNING candidate_code;
		`
		err = pool.QueryRow(ctx, query, p.CandidateCode, p.Name, p.Email, p.Phone, locID, qualID, p.DOB, p.ResumeText).Scan(&code)
	} else {
		query := `
			INSERT INTO candidates (name, email, phone, location_id, qualification_id, dob, resume_text)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::date, $7)
			RETURNING candidate_code;
		`
		err = pool.QueryRow(ctx, query, p.Name, p.Email, p.Phone, locID, qualID, p.DOB, p.ResumeText).Scan(&code)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create candidate: %w", err)
	}
	return code, nil
}

// UpdateCandidate updates an existing candidate profile by ID or CandidateCode.
func UpdateCandidate(ctx context.Context, pool *pgxpool.Pool, p CandidateProfile) error {
	locID, err := FindOrCreateLocation(ctx, pool, p.Location)
	if err != nil {
		return err
	}

	qualID, err := FindOrCreateQualification(ctx, pool, p.Qualification)
	if err != nil {
		return err
	}

	query := `
		UPDATE candidates
		SET name = $1,
		    email = $2,
		    phone = $3,
		    location_id = $4,
		    qualification_id = $5,
		    dob = NULLIF($6, '')::date,
		    resume_text = $7
		WHERE id = $8 OR candidate_code = $9;
	`
	_, err = pool.Exec(ctx, query, p.Name, p.Email, p.Phone, locID, qualID, p.DOB, p.ResumeText, p.ID, p.CandidateCode)
	if err != nil {
		return fmt.Errorf("failed to update candidate %s: %w", p.CandidateCode, err)
	}
	return nil
}

// DeleteCandidate removes a candidate by ID (cascades to applications, notes, and application_products).
func DeleteCandidate(ctx context.Context, pool *pgxpool.Pool, id int) error {
	query := `DELETE FROM candidates WHERE id = $1;`
	tag, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete candidate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("candidate with ID %d not found", id)
	}
	return nil
}

// DeleteCandidates removes multiple candidates by their IDs (cascades to applications, notes, and application_products).
func DeleteCandidates(ctx context.Context, pool *pgxpool.Pool, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	query := `DELETE FROM candidates WHERE id = ANY($1);`
	tag, err := pool.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("failed to delete candidates: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no candidates found to delete")
	}
	return nil
}

// FetchProfiles is an alias for FetchCandidateProfiles.
func FetchProfiles(ctx context.Context, pool *pgxpool.Pool, searchQuery string) ([]CandidateProfile, error) {
	return FetchCandidateProfiles(ctx, pool, searchQuery)
}
