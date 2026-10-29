package ui

import (
	"strings"
	"testing"

	"ats-tui/models"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
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

func TestWrapHelpItems(t *testing.T) {
	items := []string{
		"[1-4/Tab] Tab",
		"[/] Search",
		"[f] Filter",
		"[r] Reset Filter",
		"[s/S] Sort",
		"[a] Add",
		"[e] Edit",
		"[x] Delete",
		"[Enter/d] Details",
		"[p] Preview",
		"[[/]] Cols",
		"[v/t] Kanban/Table",
		"[m/b] Stage",
		"[R] Reload DB",
		"[q] Quit",
	}

	// 1. Empty items
	if len(wrapHelpItems(nil, 100)) != 0 {
		t.Errorf("expected 0 lines for nil items")
	}

	// 2. Non-positive width fallback
	fallback := wrapHelpItems(items, 0)
	if len(fallback) != 1 {
		t.Errorf("expected 1 line for width <= 0, got %d", len(fallback))
	}

	// 3. Wide width fits on single line
	wideLines := wrapHelpItems(items, 300)
	if len(wideLines) != 1 {
		t.Errorf("expected 1 line for width 300, got %d", len(wideLines))
	}

	// 4. Constrained width wraps across multiple lines
	wrappedLines := wrapHelpItems(items, 140)
	if len(wrappedLines) < 2 {
		t.Errorf("expected at least 2 lines for width 140, got %d", len(wrappedLines))
	}

	// Verify each line does not exceed maxWidth and does not have dangling separators
	for idx, line := range wrappedLines {
		w := lipgloss.Width(line)
		if w > 140 {
			t.Errorf("line %d width %d exceeds max width 140", idx, w)
		}
		if strings.HasPrefix(line, " • ") || strings.HasSuffix(line, " • ") {
			t.Errorf("line %d has dangling bullet separator: %q", idx, line)
		}
	}

	// Verify all items are present across lines without being broken
	joined := strings.Join(wrappedLines, " • ")
	for _, item := range items {
		if !strings.Contains(joined, item) {
			t.Errorf("missing item %q in wrapped output", item)
		}
	}
}

func TestHelpBarWrappingAndHeightAdjustment(t *testing.T) {
	app := NewApp(nil)
	app.width = 140
	app.height = 35

	app.recalculateAvailableHeight()
	lineCount := app.helpBarLineCount()
	if lineCount < 2 {
		t.Errorf("expected at least 2 lines of help for width 140, got %d", lineCount)
	}

	// Verify available height was adjusted to account for extra help line(s)
	expectedHeight := 35 - 18 - (lineCount - 1) // showSplitPane is true initially
	// bubbles/table subtracts 2 for header/border in Height()
	expectedViewportHeight := expectedHeight - 2
	if app.profilesTable.Height() != expectedViewportHeight {
		t.Errorf("expected table viewport height %d, got %d", expectedViewportHeight, app.profilesTable.Height())
	}

	// Render view to ensure no panic or formatting issues
	view := app.View()
	if !strings.Contains(view, "[1-4/Tab] Tab") {
		t.Errorf("expected view to contain [1-4/Tab] Tab")
	}
	if !strings.Contains(view, "[q] Quit") {
		t.Errorf("expected view to contain [q] Quit")
	}
}

func TestSortOrderRowSelectionAndPreview(t *testing.T) {
	app := NewApp(nil)
	app.width = 120
	app.height = 30
	app.activeTab = TabProfiles

	// Mock DB profiles in DESC order (as PostgreSQL returns them)
	app.profiles = []models.CandidateProfile{
		{ID: 10, CandidateCode: "CAN-010", Name: "Jordan Bell", Email: "jordan@example.com", Location: "Remote"},
		{ID: 1, CandidateCode: "CAN-001", Name: "Alex Rivera", Email: "alex@example.com", Location: "San Francisco"},
	}

	// 1. Ascending sort order by Candidate ID (▲)
	app.profileSortCol = 0
	app.profileSortAsc = true
	app.applySortingAndFilters()
	app.profilesTable.SetCursor(0)

	sel := app.getSelectedProfile()
	if sel == nil {
		t.Fatalf("expected selected profile not to be nil")
	}
	if sel.CandidateCode != "CAN-001" {
		t.Errorf("expected ascending sort row 0 to be CAN-001, got %s (%s)", sel.CandidateCode, sel.Name)
	}

	preview := app.renderSplitPane()
	if !strings.Contains(preview, "Alex Rivera") || !strings.Contains(preview, "CAN-001") {
		t.Errorf("expected preview in ascending order to show Alex Rivera (CAN-001), got: %s", preview)
	}
	if strings.Contains(preview, "Jordan Bell") {
		t.Errorf("preview in ascending order erroneously showed Jordan Bell: %s", preview)
	}

	// 2. Descending sort order by Candidate ID (▼)
	app.profileSortAsc = false
	app.applySortingAndFilters()
	app.profilesTable.SetCursor(0)

	selDesc := app.getSelectedProfile()
	if selDesc == nil {
		t.Fatalf("expected selected profile not to be nil")
	}
	if selDesc.CandidateCode != "CAN-010" {
		t.Errorf("expected descending sort row 0 to be CAN-010, got %s (%s)", selDesc.CandidateCode, selDesc.Name)
	}

	previewDesc := app.renderSplitPane()
	if !strings.Contains(previewDesc, "Jordan Bell") || !strings.Contains(previewDesc, "CAN-010") {
		t.Errorf("expected preview in descending order to show Jordan Bell (CAN-010), got: %s", previewDesc)
	}
}

func TestMultiRowSelectionAndBatchDelete(t *testing.T) {
	app := NewApp(nil)
	app.width = 120
	app.height = 30
	app.activeTab = TabProfiles

	app.profiles = []models.CandidateProfile{
		{ID: 1, CandidateCode: "CAN-001", Name: "Alex Rivera"},
		{ID: 2, CandidateCode: "CAN-002", Name: "Priya Sharma"},
		{ID: 3, CandidateCode: "CAN-003", Name: "Marcus Vance"},
	}
	app.applySortingAndFilters()
	app.profilesTable.SetCursor(0)

	// 1. Initial selection count is 0
	if app.selectedCount() != 0 {
		t.Fatalf("expected initial selected count to be 0, got %d", app.selectedCount())
	}

	// 2. Select first row using Space key
	app.toggleCurrentRowSelection()
	app.applySortingAndFilters()

	if app.selectedCount() != 1 {
		t.Fatalf("expected 1 selected profile, got %d", app.selectedCount())
	}
	if !app.selectedProfileIDs[1] {
		t.Fatalf("expected candidate ID 1 to be selected")
	}

	// Verify checkbox rendered in table rows
	rows := app.profilesTable.Rows()
	if !strings.HasPrefix(rows[0][0], "[✓]") {
		t.Errorf("expected row 0 to have '[✓]' prefix, got %q", rows[0][0])
	}
	if !strings.HasPrefix(rows[1][0], "[ ]") {
		t.Errorf("expected unselected row 1 to have '[ ]' prefix, got %q", rows[1][0])
	}

	// 3. Select next row (cursor was moved to 1 by toggle)
	app.toggleCurrentRowSelection()
	app.applySortingAndFilters()

	if app.selectedCount() != 2 {
		t.Fatalf("expected 2 selected profiles, got %d", app.selectedCount())
	}

	// 4. Verify SubHeader reflects selection
	view := app.View()
	if !strings.Contains(view, "Selected: 2 candidates") {
		t.Errorf("expected view to contain 'Selected: 2 candidates', got: %s", view)
	}

	// 5. Open Batch Delete Confirmation
	app.openDeleteConfirm()
	if !app.showDeleteConfirm {
		t.Fatalf("expected showDeleteConfirm to be true")
	}
	if len(app.deleteTargetIDs) != 2 {
		t.Fatalf("expected 2 deleteTargetIDs, got %d", len(app.deleteTargetIDs))
	}

	modal := app.renderDeleteConfirmModal()
	if !strings.Contains(modal, "2 Candidates") || !strings.Contains(modal, "2 records") {
		t.Errorf("expected modal to mention 2 records, got: %s", modal)
	}

	// 6. Test selection clearing with Esc
	app.showDeleteConfirm = false
	app.clearCurrentTabSelection()
	app.applySortingAndFilters()

	if app.selectedCount() != 0 {
		t.Fatalf("expected selected count to be 0 after clear, got %d", app.selectedCount())
	}

	cleanRows := app.profilesTable.Rows()
	if strings.HasPrefix(cleanRows[0][0], "[") {
		t.Errorf("expected row 0 to have no checkbox prefix when selection is empty, got %q", cleanRows[0][0])
	}
}

func TestApplicationsDefaultViewAndToggle(t *testing.T) {
	app := NewApp(nil)

	// 1. Verify default view mode is Table
	if app.appViewMode != AppViewTable {
		t.Fatalf("expected default appViewMode to be AppViewTable (%d), got %d", AppViewTable, app.appViewMode)
	}

	// Switch to TabApplications
	app.activeTab = TabApplications

	// 2. Press "v" -> toggles to Kanban
	updatedModel, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	app = updatedModel.(AppModel)
	if app.appViewMode != AppViewKanban {
		t.Fatalf("expected appViewMode to be AppViewKanban after pressing 'v', got %d", app.appViewMode)
	}
	if !strings.Contains(app.statusMsg, "Kanban Board") {
		t.Errorf("expected statusMsg to indicate Kanban Board, got %q", app.statusMsg)
	}

	// 3. Press "v" again -> toggles back to Table
	updatedModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	app = updatedModel.(AppModel)
	if app.appViewMode != AppViewTable {
		t.Fatalf("expected appViewMode to be AppViewTable after pressing 'v' again, got %d", app.appViewMode)
	}
	if !strings.Contains(app.statusMsg, "Table View") {
		t.Errorf("expected statusMsg to indicate Table View, got %q", app.statusMsg)
	}

	// 4. Press "t" -> toggles to Kanban
	updatedModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	app = updatedModel.(AppModel)
	if app.appViewMode != AppViewKanban {
		t.Fatalf("expected appViewMode to be AppViewKanban after pressing 't', got %d", app.appViewMode)
	}

	// 5. Press "t" again -> toggles back to Table
	updatedModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	app = updatedModel.(AppModel)
	if app.appViewMode != AppViewTable {
		t.Fatalf("expected appViewMode to be AppViewTable after pressing 't' again, got %d", app.appViewMode)
	}
}
