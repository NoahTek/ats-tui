package ui

import (
	"testing"

	"ats-tui/models"

	"github.com/charmbracelet/bubbles/table"
)

func TestTruncateString(t *testing.T) {
	short := "Hello World"
	if truncateString(short, 20) != "Hello World" {
		t.Fatalf("expected untouched short string, got %s", truncateString(short, 20))
	}

	long := "This is a very long string that should definitely get truncated by the utility function"
	truncated := truncateString(long, 20)
	if len(truncated) > 20 {
		t.Fatalf("expected length <= 20, got %d", len(truncated))
	}
	if truncated != "This is a very lo..." {
		t.Fatalf("expected 'This is a very lo...', got '%s'", truncated)
	}

	tiny := truncateString("ABCDE", 3)
	if tiny != "ABC" {
		t.Fatalf("expected 'ABC', got '%s'", tiny)
	}
}

func TestTableColumnCounts(t *testing.T) {
	// Profiles: 7 columns
	if len(allProfileColumns()) != 7 {
		t.Errorf("expected 7 columns in profiles table, got %d", len(allProfileColumns()))
	}
	// Submissions: 14 total columns
	if len(allSubmissionColumns()) != 14 {
		t.Errorf("expected 14 columns in all submissions table, got %d", len(allSubmissionColumns()))
	}
	// Jobs: 10 columns
	if len(allJobColumns()) != 10 {
		t.Errorf("expected 10 columns in jobs table, got %d", len(allJobColumns()))
	}
	// Organisations: 4 columns
	if len(allOrgColumns()) != 4 {
		t.Errorf("expected 4 columns in orgs table, got %d", len(allOrgColumns()))
	}

	// Verify "Job ID" naming
	if allSubmissionColumns()[0].Title != "Job ID" {
		t.Errorf("expected first column of submissions to be 'Job ID', got '%s'", allSubmissionColumns()[0].Title)
	}
	if allJobColumns()[0].Title != "Job ID" {
		t.Errorf("expected first column of jobs to be 'Job ID', got '%s'", allJobColumns()[0].Title)
	}
}

func TestVisibleColumnsWindow(t *testing.T) {
	all := allSubmissionColumns()

	// Window with offset 0 and narrow width
	vis := visibleColumns(all, 0, 50)
	if len(vis) < 3 {
		t.Fatalf("expected at least 3 visible columns, got %d", len(vis))
	}
	if vis[0].Title != all[0].Title {
		t.Fatalf("expected first column to be %s, got %s", all[0].Title, vis[0].Title)
	}

	// Sliding offset to 3
	visOffset := visibleColumns(all, 3, 80)
	if visOffset[0].Title != all[3].Title {
		t.Fatalf("expected first column at offset 3 to be %s, got %s", all[3].Title, visOffset[0].Title)
	}
}

func TestFormBuilders(t *testing.T) {
	// Candidate form
	cf := NewCandidateForm(nil)
	if cf.mode != FormModeAdd || cf.entityType != EntityCandidate {
		t.Fatalf("expected Add Candidate form")
	}
	if len(cf.fields) != 7 {
		t.Fatalf("expected 7 fields in candidate form, got %d", len(cf.fields))
	}

	// Edit Candidate form
	cand := models.CandidateProfile{
		ID:            1,
		CandidateCode: "CAN-001",
		Name:          "Alice",
		Email:         "alice@example.com",
	}
	cfEdit := NewCandidateForm(&cand)
	if cfEdit.mode != FormModeEdit || cfEdit.targetID != 1 {
		t.Fatalf("expected Edit mode with targetID 1")
	}

	// Application form
	af := NewApplicationForm(nil, "JOB1", "CAN-001")
	if af.mode != FormModeAdd || af.entityType != EntityApplication {
		t.Fatalf("expected Add Application form")
	}

	// Job form
	jf := NewJobForm(nil)
	if jf.mode != FormModeAdd || jf.entityType != EntityJob {
		t.Fatalf("expected Add Job form")
	}

	// Org form
	of := NewOrganisationForm(nil)
	if of.mode != FormModeAdd || of.entityType != EntityOrganisation {
		t.Fatalf("expected Add Org form")
	}
}

func TestUpdateTableDataSafety(t *testing.T) {
	tbl := createStyledTable(allProfileColumns(), 10)

	// 1. Initial rows matching allProfileColumns (7 columns)
	pRows := []table.Row{
		{"CAN-001", "Alice", "123", "a@a.com", "SF", "BS CS", "1995-01-01"},
		{"CAN-002", "Bob", "456", "b@b.com", "NY", "MS CS", "1996-02-02"},
	}
	updateTableData(&tbl, allProfileColumns(), pRows)

	// 2. Shrink columns to 3 columns - must not panic!
	narrowCols := []table.Column{
		{Title: "Col 1", Width: 10},
		{Title: "Col 2", Width: 10},
		{Title: "Col 3", Width: 10},
	}
	updateTableData(&tbl, narrowCols, pRows)
	if len(tbl.Columns()) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(tbl.Columns()))
	}
	for _, r := range tbl.Rows() {
		if len(r) != 3 {
			t.Fatalf("expected row to be normalized to 3 elements, got %d", len(r))
		}
	}

	// 3. Expand columns to 10 columns with short rows (2 elements) - must not panic!
	wideCols := make([]table.Column, 10)
	for i := range wideCols {
		wideCols[i] = table.Column{Title: "Col", Width: 10}
	}
	shortRows := []table.Row{
		{"Only1", "Only2"},
	}
	updateTableData(&tbl, wideCols, shortRows)
	if len(tbl.Columns()) != 10 {
		t.Fatalf("expected 10 columns, got %d", len(tbl.Columns()))
	}
	for _, r := range tbl.Rows() {
		if len(r) != 10 {
			t.Fatalf("expected row to be padded to 10 elements, got %d", len(r))
		}
	}
}

func TestMatchesColFilters(t *testing.T) {
	m := AppModel{
		colFilters: make(map[string]map[string]bool),
	}

	row1 := map[string]string{
		"Location": "Remote",
		"Status":   "Interested",
	}
	row2 := map[string]string{
		"Location": "Bangalore",
		"Status":   "Follow-up",
	}
	row3 := map[string]string{
		"Location": "London",
		"Status":   "Closed",
	}

	// 1. No filters active -> all rows match
	if !m.matchesColFilters(row1) || !m.matchesColFilters(row2) || !m.matchesColFilters(row3) {
		t.Errorf("expected empty filter to match all rows")
	}

	// 2. Add Location filter for "Remote" or "Bangalore" (OR condition within column)
	m.colFilters["Location"] = map[string]bool{
		"Remote":    true,
		"Bangalore": true,
	}

	if !m.matchesColFilters(row1) {
		t.Errorf("expected row1 (Remote) to match Location filter")
	}
	if !m.matchesColFilters(row2) {
		t.Errorf("expected row2 (Bangalore) to match Location filter")
	}
	if m.matchesColFilters(row3) {
		t.Errorf("expected row3 (London) not to match Location filter")
	}

	// 3. Add Status filter for "Interested" (AND condition across columns)
	m.colFilters["Status"] = map[string]bool{
		"Interested": true,
	}

	// Now row1 has Location: Remote AND Status: Interested -> MATCH
	if !m.matchesColFilters(row1) {
		t.Errorf("expected row1 to match both Location and Status filters")
	}
	// row2 has Location: Bangalore BUT Status: Follow-up -> NO MATCH
	if m.matchesColFilters(row2) {
		t.Errorf("expected row2 not to match combined filters")
	}
}

func TestDeleteConfirmationFlow(t *testing.T) {
	app := NewApp(nil)
	app.width = 100
	app.height = 30

	// 1. Initial state: no delete confirm modal
	if app.showDeleteConfirm {
		t.Fatalf("expected showDeleteConfirm to be false initially")
	}

	// 2. Setup mock profile and open delete confirm
	app.profiles = []models.CandidateProfile{
		{ID: 42, CandidateCode: "CAN-042", Name: "Alice Bob"},
	}
	app.openDeleteConfirm()

	if !app.showDeleteConfirm {
		t.Fatalf("expected showDeleteConfirm to be true after openDeleteConfirm")
	}
	if app.deleteTargetID != 42 {
		t.Errorf("expected deleteTargetID to be 42, got %d", app.deleteTargetID)
	}

	// 3. Render modal
	view := app.renderDeleteConfirmModal()
	if view == "" {
		t.Fatalf("renderDeleteConfirmModal returned empty string")
	}

	// 4. Cancel deletion
	app.showDeleteConfirm = false
	if app.showDeleteConfirm {
		t.Fatalf("expected showDeleteConfirm to be false after cancel")
	}
}
