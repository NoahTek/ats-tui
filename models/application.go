package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Stage string

const (
	StageInterested Stage = "interested"
	StageFollowUp   Stage = "follow_up"
	StageCallBack   Stage = "call_back"
	StageClosed     Stage = "closed"
)

// PipelineStages defines the chronological order of Kanban columns.
var PipelineStages = []Stage{
	StageInterested,
	StageFollowUp,
	StageCallBack,
	StageClosed,
}

// StageTitle returns human-friendly column titles.
func (s Stage) Title() string {
	switch s {
	case StageInterested:
		return "Interested"
	case StageFollowUp:
		return "Follow-Up"
	case StageCallBack:
		return "Call-Back"
	case StageClosed:
		return "Closed"
	default:
		return string(s)
	}
}

// Next returns the succeeding stage, if any.
func (s Stage) Next() (Stage, bool) {
	for i, stage := range PipelineStages {
		if stage == s && i+1 < len(PipelineStages) {
			return PipelineStages[i+1], true
		}
	}
	return s, false
}

// Prev returns the preceding stage, if any.
func (s Stage) Prev() (Stage, bool) {
	for i, stage := range PipelineStages {
		if stage == s && i-1 >= 0 {
			return PipelineStages[i-1], true
		}
	}
	return s, false
}

// ApplicationCard contains candidate and application info needed for Kanban board cards.
type ApplicationCard struct {
	ApplicationID    int
	CandidateID      int
	CandidateCode    string
	CandidateName    string
	Email            string
	Phone            string
	JobCode          string
	JobTitle         string
	CurrentOrg       string
	Designation      string
	ExperienceYears  int
	ExpectedCTC      int
	NoticePeriodDays int
	Skill            string
	Products         string
	LatestComment    string
	Status           Stage
	ResumeText       string
}

// ApplicationSubmissionRow captures the 20-column client Active Pipeline / Submissions sheet (Sheet 2).
type ApplicationSubmissionRow struct {
	ApplicationID    int
	CandidateID      int
	JobCode          string
	JobTitle         string
	Role             string
	Department       string
	SubDepartment    string
	CandidateCode    string
	CandidateName    string
	Phone            string
	Email            string
	Location         string
	Qualification    string
	DOB              string
	CurrentOrg       string
	Designation      string
	Products         string
	ExperienceYears  int
	CTC              int
	ExpectedCTC      int
	NoticePeriodDays int
	Comments         string
	Status           Stage
	ResumeText       string
}

// FetchBoardApplications queries applications grouped by stage, with optional jobCode filter.
func FetchBoardApplications(ctx context.Context, pool *pgxpool.Pool, jobCodeFilter string) (map[Stage][]ApplicationCard, error) {
	var query strings.Builder
	var args []any

	query.WriteString(`
		SELECT 
			a.id, 
			c.id, 
			COALESCE(c.candidate_code, 'CAND_' || c.id),
			c.name, 
			c.email, 
			COALESCE(c.phone, ''), 
			COALESCE(j.job_code, 'JOB?'),
			COALESCE(j.job_title, 'General Role'), 
			COALESCE(a.current_org, ''), 
			COALESCE(a.designation, ''), 
			COALESCE(a.experience_years, 0), 
			COALESCE(a.expected_ctc, 0), 
			COALESCE(a.notice_period_in_days, 0),
			COALESCE(s.skill_name, 'N/A'), 
			COALESCE(string_agg(DISTINCT p.name, ', '), ''),
			COALESCE((SELECT content FROM notes WHERE application_id = a.id ORDER BY created_at DESC LIMIT 1), ''),
			a.status,
			COALESCE(c.resume_text, '')
		FROM applications a
		JOIN candidates c ON a.candidate_id = c.id
		LEFT JOIN jobs j ON a.job_id = j.id
		LEFT JOIN skill s ON a.skill_id = s.id
		LEFT JOIN application_products ap ON a.id = ap.application_id
		LEFT JOIN product p ON ap.product_id = p.id
	`)

	if strings.TrimSpace(jobCodeFilter) != "" && strings.ToUpper(jobCodeFilter) != "ALL" {
		query.WriteString(` WHERE j.job_code = $1 `)
		args = append(args, jobCodeFilter)
	}

	query.WriteString(`
		GROUP BY a.id, c.id, j.id, s.id
		ORDER BY a.created_at ASC, a.id ASC;
	`)

	rows, err := pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}
	defer rows.Close()

	stageMap := make(map[Stage][]ApplicationCard)
	for _, s := range PipelineStages {
		stageMap[s] = []ApplicationCard{}
	}

	for rows.Next() {
		var card ApplicationCard
		var statusStr string
		err := rows.Scan(
			&card.ApplicationID,
			&card.CandidateID,
			&card.CandidateCode,
			&card.CandidateName,
			&card.Email,
			&card.Phone,
			&card.JobCode,
			&card.JobTitle,
			&card.CurrentOrg,
			&card.Designation,
			&card.ExperienceYears,
			&card.ExpectedCTC,
			&card.NoticePeriodDays,
			&card.Skill,
			&card.Products,
			&card.LatestComment,
			&statusStr,
			&card.ResumeText,
		)
		if err != nil {
			return nil, fmt.Errorf("scan application card failed: %w", err)
		}
		card.Status = Stage(statusStr)
		stageMap[card.Status] = append(stageMap[card.Status], card)
	}

	return stageMap, nil
}

// FetchApplicationSubmissions returns full 20-column records matching the client Applications table sheet.
func FetchApplicationSubmissions(ctx context.Context, pool *pgxpool.Pool, jobCodeFilter string) ([]ApplicationSubmissionRow, error) {
	var query strings.Builder
	var args []any

	query.WriteString(`
		SELECT 
			a.id,
			c.id,
			COALESCE(j.job_code, 'JOB?'),
			COALESCE(j.job_title, ''),
			COALESCE(j.role, ''),
			COALESCE(j.department, ''),
			COALESCE(j.sub_department, ''),
			COALESCE(c.candidate_code, ''),
			c.name,
			COALESCE(c.phone, ''),
			c.email,
			COALESCE(l.location, 'Unassigned'),
			COALESCE(eq.qualification, 'Unassigned'),
			COALESCE(to_char(c.dob, 'YYYY-MM-DD'), ''),
			COALESCE(a.current_org, ''),
			COALESCE(a.designation, ''),
			COALESCE(string_agg(DISTINCT p.name, ', '), ''),
			COALESCE(a.experience_years, 0),
			COALESCE(a.ctc, 0),
			COALESCE(a.expected_ctc, 0),
			COALESCE(a.notice_period_in_days, 0),
			COALESCE((SELECT content FROM notes WHERE application_id = a.id ORDER BY created_at DESC LIMIT 1), ''),
			a.status,
			COALESCE(c.resume_text, '')
		FROM applications a
		JOIN candidates c ON a.candidate_id = c.id
		LEFT JOIN jobs j ON a.job_id = j.id
		LEFT JOIN location l ON c.location_id = l.id
		LEFT JOIN educational_qualification eq ON c.qualification_id = eq.id
		LEFT JOIN application_products ap ON a.id = ap.application_id
		LEFT JOIN product p ON ap.product_id = p.id
	`)

	if strings.TrimSpace(jobCodeFilter) != "" && strings.ToUpper(jobCodeFilter) != "ALL" {
		query.WriteString(` WHERE j.job_code = $1 `)
		args = append(args, jobCodeFilter)
	}

	query.WriteString(`
		GROUP BY a.id, c.id, j.id, l.id, eq.id
		ORDER BY a.created_at ASC, a.id ASC;
	`)

	rows, err := pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application submissions: %w", err)
	}
	defer rows.Close()

	var list []ApplicationSubmissionRow
	for rows.Next() {
		var r ApplicationSubmissionRow
		var statusStr string
		err := rows.Scan(
			&r.ApplicationID,
			&r.CandidateID,
			&r.JobCode,
			&r.JobTitle,
			&r.Role,
			&r.Department,
			&r.SubDepartment,
			&r.CandidateCode,
			&r.CandidateName,
			&r.Phone,
			&r.Email,
			&r.Location,
			&r.Qualification,
			&r.DOB,
			&r.CurrentOrg,
			&r.Designation,
			&r.Products,
			&r.ExperienceYears,
			&r.CTC,
			&r.ExpectedCTC,
			&r.NoticePeriodDays,
			&r.Comments,
			&statusStr,
			&r.ResumeText,
		)
		if err != nil {
			return nil, fmt.Errorf("scan application submission row failed: %w", err)
		}
		r.Status = Stage(statusStr)
		list = append(list, r)
	}

	return list, nil
}

// UpdateApplicationStage updates an application's stage in Postgres.
func UpdateApplicationStage(ctx context.Context, pool *pgxpool.Pool, appID int, newStage Stage) error {
	query := `UPDATE applications SET status = $1 WHERE id = $2;`
	_, err := pool.Exec(ctx, query, string(newStage), appID)
	if err != nil {
		return fmt.Errorf("failed to update application stage: %w", err)
	}
	return nil
}

// CreateApplication inserts a new application submission into PostgreSQL.
func CreateApplication(ctx context.Context, pool *pgxpool.Pool, sub ApplicationSubmissionRow, initialComment string) error {
	// 1. Resolve Job ID
	var jobID int
	err := pool.QueryRow(ctx, `SELECT id FROM jobs WHERE job_code = $1 LIMIT 1;`, sub.JobCode).Scan(&jobID)
	if err != nil {
		return fmt.Errorf("job code '%s' not found: %w", sub.JobCode, err)
	}

	// 2. Resolve Candidate ID
	var candidateID int
	if sub.CandidateID > 0 {
		candidateID = sub.CandidateID
	} else {
		err = pool.QueryRow(ctx, `SELECT id FROM candidates WHERE candidate_code = $1 LIMIT 1;`, sub.CandidateCode).Scan(&candidateID)
		if err != nil {
			return fmt.Errorf("candidate code '%s' not found: %w", sub.CandidateCode, err)
		}
	}

	// 3. Status validation
	status := sub.Status
	if status == "" {
		status = StageInterested
	}

	// 4. Insert application
	var appID int
	query := `
		INSERT INTO applications (
			job_id, candidate_id, current_org, designation, experience_years, ctc, expected_ctc, notice_period_in_days, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id;
	`
	err = pool.QueryRow(ctx, query,
		jobID, candidateID, sub.CurrentOrg, sub.Designation, sub.ExperienceYears, sub.CTC, sub.ExpectedCTC, sub.NoticePeriodDays, string(status),
	).Scan(&appID)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	// 5. Link products if any
	if strings.TrimSpace(sub.Products) != "" {
		for _, pName := range strings.Split(sub.Products, ",") {
			pName = strings.TrimSpace(pName)
			if pName == "" {
				continue
			}
			pID, err := FindOrCreateProduct(ctx, pool, pName)
			if err == nil && pID > 0 {
				_, _ = pool.Exec(ctx, `INSERT INTO application_products (application_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, appID, pID)
			}
		}
	}

	// 6. Insert initial comment if any
	initialComment = strings.TrimSpace(initialComment)
	if initialComment != "" {
		_, _ = pool.Exec(ctx, `INSERT INTO notes (application_id, note_type, content) VALUES ($1, 'Comment', $2);`, appID, initialComment)
	}

	return nil
}

// UpdateApplication updates an existing application submission.
func UpdateApplication(ctx context.Context, pool *pgxpool.Pool, sub ApplicationSubmissionRow) error {
	query := `
		UPDATE applications
		SET current_org = $1,
		    designation = $2,
		    experience_years = $3,
		    ctc = $4,
		    expected_ctc = $5,
		    notice_period_in_days = $6,
		    status = $7
		WHERE id = $8;
	`
	_, err := pool.Exec(ctx, query,
		sub.CurrentOrg, sub.Designation, sub.ExperienceYears, sub.CTC, sub.ExpectedCTC, sub.NoticePeriodDays, string(sub.Status), sub.ApplicationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update application ID %d: %w", sub.ApplicationID, err)
	}

	// Sync products
	if strings.TrimSpace(sub.Products) != "" {
		_, _ = pool.Exec(ctx, `DELETE FROM application_products WHERE application_id = $1;`, sub.ApplicationID)
		for _, pName := range strings.Split(sub.Products, ",") {
			pName = strings.TrimSpace(pName)
			if pName == "" {
				continue
			}
			pID, err := FindOrCreateProduct(ctx, pool, pName)
			if err == nil && pID > 0 {
				_, _ = pool.Exec(ctx, `INSERT INTO application_products (application_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, sub.ApplicationID, pID)
			}
		}
	}

	// Insert comment if new comment text is provided
	comment := strings.TrimSpace(sub.Comments)
	if comment != "" {
		var exists bool
		_ = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notes WHERE application_id = $1 AND content = $2);`, sub.ApplicationID, comment).Scan(&exists)
		if !exists {
			_, _ = pool.Exec(ctx, `INSERT INTO notes (application_id, note_type, content) VALUES ($1, 'Comment', $2);`, sub.ApplicationID, comment)
		}
	}

	return nil
}

// DeleteApplication removes an application submission by ID (cascades to notes and application_products).
func DeleteApplication(ctx context.Context, pool *pgxpool.Pool, appID int) error {
	query := `DELETE FROM applications WHERE id = $1;`
	tag, err := pool.Exec(ctx, query, appID)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("application with ID %d not found", appID)
	}
	return nil
}
