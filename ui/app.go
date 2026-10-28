package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"ats-tui/kanban"
	"ats-tui/models"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Active Tab Constants
const (
	TabProfiles      = 0
	TabApplications  = 1
	TabJobs          = 2
	TabOrganisations = 3
)

// Applications View Mode
const (
	AppViewKanban = 0
	AppViewTable  = 1
)

// Messages
type ProfilesLoadedMsg struct {
	Profiles []models.CandidateProfile
}

type SubmissionsLoadedMsg struct {
	Submissions []models.ApplicationSubmissionRow
}

type JobsLoadedMsg struct {
	Jobs []models.JobPosting
}

type OrgsLoadedMsg struct {
	Orgs []models.Organisation
}

type NotesLoadedMsg struct {
	AppID int
	Notes []models.NoteRecord
}

type CandidateSavedMsg struct {
	Code   string
	IsEdit bool
}

type ApplicationSavedMsg struct {
	CandidateCode string
	JobID         string
	IsEdit        bool
}

type JobSavedMsg struct {
	JobID  string
	IsEdit bool
}

type OrgSavedMsg struct {
	OrgCode string
	IsEdit  bool
}

type CandidateDeletedMsg struct {
	Desc string
}

type ApplicationDeletedMsg struct {
	Desc string
}

type JobDeletedMsg struct {
	Desc string
}

type OrgDeletedMsg struct {
	Desc string
}

type AppErrMsg struct {
	Err error
}

type AppModel struct {
	pool *pgxpool.Pool

	// Navigation & View State
	activeTab   int
	appViewMode int // Kanban vs Table in Applications
	width       int
	height      int

	// Master Data Slices
	profiles    []models.CandidateProfile
	submissions []models.ApplicationSubmissionRow
	jobs        []models.JobPosting
	orgs        []models.Organisation

	// Tables
	profilesTable    table.Model
	submissionsTable table.Model
	jobsTable        table.Model
	orgsTable        table.Model

	// Universal Horizontal Column Slicing across all 4 tabs
	tabColOffsets [4]int

	// Kanban
	kanbanModel kanban.Model

	// Global Search
	searchInput textinput.Model
	isSearching bool
	searchQuery string

	// Column Filter Modal State
	isColFilterModal     bool
	colFilterColIdx      int
	colFilterValIdx      int
	colFilterSearching   bool
	colFilterSearchInput textinput.Model
	colFilters           map[string]map[string]bool // [column][value] -> checked

	// Sorting State
	profileSortCol int
	profileSortAsc bool
	appSortCol     int
	appSortAsc     bool
	jobSortCol     int
	jobSortAsc     bool
	orgSortCol     int
	orgSortAsc     bool

	// Deletion Confirmation Modal
	showDeleteConfirm bool
	deleteTargetType  EntityType
	deleteTargetID    int
	deleteTargetDesc  string

	// CRUD Form Modal
	showFormModal bool
	formModal     FormModel

	// Dual Inspection
	showSplitPane  bool
	showModal      bool
	detailViewport viewport.Model
	modalTitle     string
	activeAppNotes []models.NoteRecord

	// Status
	statusMsg string
	isError   bool
}

func NewApp(pool *pgxpool.Pool) AppModel {
	ti := textinput.New()
	ti.Placeholder = "Type search term... [Enter to apply, Esc to cancel]"
	ti.CharLimit = 100
	ti.Width = 60

	filterTi := textinput.New()
	filterTi.Placeholder = "Type to search values... [Enter/Esc to exit]"
	filterTi.CharLimit = 50
	filterTi.Width = 36

	vp := viewport.New(80, 20)

	pTable := createStyledTable(allProfileColumns(), 10)
	sTable := createStyledTable(allSubmissionColumns(), 10)
	jTable := createStyledTable(allJobColumns(), 10)
	oTable := createStyledTable(allOrgColumns(), 10)

	kModel := kanban.NewModel(pool)

	return AppModel{
		pool:                 pool,
		activeTab:            TabProfiles,
		appViewMode:          AppViewKanban,
		profilesTable:        pTable,
		submissionsTable:     sTable,
		jobsTable:            jTable,
		orgsTable:            oTable,
		kanbanModel:          kModel,
		searchInput:          ti,
		colFilterSearchInput: filterTi,
		colFilters:           make(map[string]map[string]bool),
		detailViewport:       vp,
		showSplitPane:        true,
		profileSortAsc:       true,
		appSortAsc:           true,
		jobSortAsc:           true,
		orgSortAsc:           true,
		statusMsg:            "Welcome to ATS TUI. Use [1-4] or [Tab] to navigate.",
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.kanbanModel.Init(),
		fetchProfilesCmd(m.pool, ""),
		fetchSubmissionsCmd(m.pool, ""),
		fetchJobsCmd(m.pool),
		fetchOrgsCmd(m.pool),
	)
}

// Full Column Definitions across all 4 tabs
func allProfileColumns() []table.Column {
	return []table.Column{
		{Title: "Candidate ID", Width: 14},
		{Title: "Candidate Name", Width: 20},
		{Title: "Contact Number", Width: 18},
		{Title: "Email ID", Width: 26},
		{Title: "Location", Width: 18},
		{Title: "Educational Qualification", Width: 28},
		{Title: "Date of Birth", Width: 14},
	}
}

func allSubmissionColumns() []table.Column {
	return []table.Column{
		{Title: "Job ID", Width: 10},
		{Title: "Candidate ID", Width: 14},
		{Title: "Candidate Name", Width: 18},
		{Title: "Designation", Width: 20},
		{Title: "Exp", Width: 6},
		{Title: "CTC", Width: 10},
		{Title: "Exp CTC", Width: 10},
		{Title: "Status", Width: 12},
		{Title: "Current Org", Width: 20},
		{Title: "Products", Width: 26},
		{Title: "Notice", Width: 8},
		{Title: "Location", Width: 18},
		{Title: "Qualification", Width: 24},
		{Title: "Latest Comments", Width: 32},
	}
}

func allJobColumns() []table.Column {
	return []table.Column{
		{Title: "Job ID", Width: 10},
		{Title: "Client", Width: 24},
		{Title: "Job Title", Width: 28},
		{Title: "Role", Width: 22},
		{Title: "Department", Width: 16},
		{Title: "Sub-department", Width: 20},
		{Title: "Location", Width: 16},
		{Title: "Req Date", Width: 12},
		{Title: "Positions", Width: 10},
		{Title: "Profiles Shared", Width: 16},
	}
}

func allOrgColumns() []table.Column {
	return []table.Column{
		{Title: "Organisation Name", Width: 28},
		{Title: "Org. Code", Width: 12},
		{Title: "Industry", Width: 26},
		{Title: "Sub Category", Width: 26},
	}
}

func (m AppModel) getAllColumnsForTab(tab int) []table.Column {
	switch tab {
	case TabProfiles:
		return allProfileColumns()
	case TabApplications:
		return allSubmissionColumns()
	case TabJobs:
		return allJobColumns()
	case TabOrganisations:
		return allOrgColumns()
	default:
		return nil
	}
}

func (m AppModel) getAllColumnsForActiveTab() []table.Column {
	return m.getAllColumnsForTab(m.activeTab)
}

// visibleColumns computes the sliding column window that fits within availableWidth
func visibleColumns(all []table.Column, offset int, availableWidth int) []table.Column {
	if len(all) == 0 {
		return nil
	}
	if offset < 0 {
		offset = 0
	}
	if len(all) >= 3 && offset > len(all)-3 {
		offset = len(all) - 3
	} else if offset >= len(all) {
		offset = 0
	}

	var visible []table.Column
	curWidth := 0
	for i := offset; i < len(all); i++ {
		w := all[i].Width + 2 // include cell padding
		if curWidth+w > availableWidth && len(visible) >= 2 {
			break
		}
		visible = append(visible, all[i])
		curWidth += w
	}

	if len(visible) == 0 {
		limit := 3
		if len(all) < limit {
			limit = len(all)
		}
		return all[:limit]
	}
	return visible
}

// createStyledTable sets strict header and cell alignment to eliminate cumulative horizontal drift
func createStyledTable(cols []table.Column, height int) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	s := table.DefaultStyles()
	// Fix cumulative drift: remove left, right, top borders on Header so it doesn't add +2 chars per column
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#5A56E0")).
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("#FFFFFF"))
	// Ensure identical cell horizontal padding
	s.Cell = s.Cell.Padding(0, 1)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#5A56E0")).
		Bold(true)
	t.SetStyles(s)
	return t
}

// updateTableData safely synchronizes table columns and rows, avoiding index-out-of-range panics in bubbles/table
func updateTableData(t *table.Model, cols []table.Column, rows []table.Row) {
	// 1. Temporarily clear rows so SetColumns doesn't render mismatched existing rows in UpdateViewport()
	t.SetRows(nil)
	// 2. Set new columns
	t.SetColumns(cols)
	// 3. Strictly normalize every row to match len(cols)
	targetLen := len(cols)
	safeRows := make([]table.Row, len(rows))
	for i, r := range rows {
		if len(r) == targetLen {
			safeRows[i] = r
		} else if len(r) > targetLen {
			safeRows[i] = r[:targetLen]
		} else {
			padded := make(table.Row, targetLen)
			copy(padded, r)
			for j := len(r); j < targetLen; j++ {
				padded[j] = ""
			}
			safeRows[i] = padded
		}
	}
	// 4. Set normalized rows
	t.SetRows(safeRows)
}

// Database Commands
func fetchProfilesCmd(pool *pgxpool.Pool, search string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		profiles, err := models.FetchProfiles(ctx, pool, search)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return ProfilesLoadedMsg{Profiles: profiles}
	}
}

func fetchSubmissionsCmd(pool *pgxpool.Pool, jobCode string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		submissions, err := models.FetchApplicationSubmissions(ctx, pool, jobCode)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return SubmissionsLoadedMsg{Submissions: submissions}
	}
}

func fetchJobsCmd(pool *pgxpool.Pool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		jobs, err := models.FetchJobPostings(ctx, pool)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return JobsLoadedMsg{Jobs: jobs}
	}
}

func fetchOrgsCmd(pool *pgxpool.Pool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		orgs, err := models.FetchOrganisations(ctx, pool)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return OrgsLoadedMsg{Orgs: orgs}
	}
}

func fetchNotesCmd(pool *pgxpool.Pool, appID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		notes, err := models.FetchApplicationNotes(ctx, pool, appID)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return NotesLoadedMsg{AppID: appID, Notes: notes}
	}
}

func saveCandidateCmd(pool *pgxpool.Pool, p models.CandidateProfile, isEdit bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if isEdit {
			err := models.UpdateCandidate(ctx, pool, p)
			if err != nil {
				return AppErrMsg{Err: err}
			}
			return CandidateSavedMsg{Code: p.CandidateCode, IsEdit: true}
		}

		code, err := models.CreateCandidate(ctx, pool, p)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return CandidateSavedMsg{Code: code, IsEdit: false}
	}
}

func saveApplicationCmd(pool *pgxpool.Pool, sub models.ApplicationSubmissionRow, isEdit bool, initialComment string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if isEdit {
			err := models.UpdateApplication(ctx, pool, sub)
			if err != nil {
				return AppErrMsg{Err: err}
			}
			return ApplicationSavedMsg{CandidateCode: sub.CandidateCode, JobID: sub.JobCode, IsEdit: true}
		}

		err := models.CreateApplication(ctx, pool, sub, initialComment)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return ApplicationSavedMsg{CandidateCode: sub.CandidateCode, JobID: sub.JobCode, IsEdit: false}
	}
}

func saveJobCmd(pool *pgxpool.Pool, j models.JobPosting, isEdit bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if isEdit {
			err := models.UpdateJob(ctx, pool, j)
			if err != nil {
				return AppErrMsg{Err: err}
			}
			return JobSavedMsg{JobID: j.JobCode, IsEdit: true}
		}

		err := models.CreateJob(ctx, pool, j)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return JobSavedMsg{JobID: j.JobCode, IsEdit: false}
	}
}

func saveOrgCmd(pool *pgxpool.Pool, o models.Organisation, isEdit bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if isEdit {
			err := models.UpdateOrganisation(ctx, pool, o)
			if err != nil {
				return AppErrMsg{Err: err}
			}
			return OrgSavedMsg{OrgCode: o.OrgCode, IsEdit: true}
		}

		err := models.CreateOrganisation(ctx, pool, o)
		if err != nil {
			return AppErrMsg{Err: err}
		}
		return OrgSavedMsg{OrgCode: o.OrgCode, IsEdit: false}
	}
}

func deleteEntityCmd(pool *pgxpool.Pool, entityType EntityType, id int, desc string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		switch entityType {
		case EntityCandidate:
			if err := models.DeleteCandidate(ctx, pool, id); err != nil {
				return AppErrMsg{Err: err}
			}
			return CandidateDeletedMsg{Desc: desc}
		case EntityApplication:
			if err := models.DeleteApplication(ctx, pool, id); err != nil {
				return AppErrMsg{Err: err}
			}
			return ApplicationDeletedMsg{Desc: desc}
		case EntityJob:
			if err := models.DeleteJob(ctx, pool, id); err != nil {
				return AppErrMsg{Err: err}
			}
			return JobDeletedMsg{Desc: desc}
		case EntityOrganisation:
			if err := models.DeleteOrganisation(ctx, pool, id); err != nil {
				return AppErrMsg{Err: err}
			}
			return OrgDeletedMsg{Desc: desc}
		}
		return nil
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.recalculateAvailableHeight()

		m.applySortingAndFilters()
		return m, nil

	case ProfilesLoadedMsg:
		m.profiles = msg.Profiles
		m.applySortingAndFilters()
		if m.searchQuery != "" {
			m.statusMsg = fmt.Sprintf("Full-text search: %d candidate(s) match '%s'", len(m.profiles), m.searchQuery)
		}
		return m, nil

	case SubmissionsLoadedMsg:
		m.submissions = msg.Submissions
		m.applySortingAndFilters()
		return m, nil

	case JobsLoadedMsg:
		m.jobs = msg.Jobs
		m.applySortingAndFilters()
		return m, nil

	case OrgsLoadedMsg:
		m.orgs = msg.Orgs
		m.applySortingAndFilters()
		return m, nil

	case NotesLoadedMsg:
		m.activeAppNotes = msg.Notes
		m.rebuildModalContent()
		return m, nil

	case CandidateSavedMsg:
		m.showFormModal = false
		action := "created"
		if msg.IsEdit {
			action = "updated"
		}
		m.statusMsg = fmt.Sprintf("Candidate %s successfully %s", msg.Code, action)
		return m, fetchProfilesCmd(m.pool, m.searchQuery)

	case ApplicationSavedMsg:
		m.showFormModal = false
		action := "created"
		if msg.IsEdit {
			action = "updated"
		}
		m.statusMsg = fmt.Sprintf("Application for %s (%s) %s", msg.CandidateCode, msg.JobID, action)
		return m, tea.Batch(
			fetchSubmissionsCmd(m.pool, ""),
			m.kanbanModel.FetchApplicationsCmd(),
		)

	case JobSavedMsg:
		m.showFormModal = false
		action := "created"
		if msg.IsEdit {
			action = "updated"
		}
		m.statusMsg = fmt.Sprintf("Job %s %s", msg.JobID, action)
		return m, fetchJobsCmd(m.pool)

	case OrgSavedMsg:
		m.showFormModal = false
		action := "created"
		if msg.IsEdit {
			action = "updated"
		}
		m.statusMsg = fmt.Sprintf("Organisation %s %s", msg.OrgCode, action)
		return m, fetchOrgsCmd(m.pool)

	case FormCancelMsg:
		m.showFormModal = false
		return m, nil

	case CandidateDeletedMsg:
		m.statusMsg = fmt.Sprintf("Successfully deleted %s", msg.Desc)
		m.isError = false
		return m, tea.Batch(
			fetchProfilesCmd(m.pool, m.searchQuery),
			fetchSubmissionsCmd(m.pool, ""),
			m.kanbanModel.FetchApplicationsCmd(),
		)

	case ApplicationDeletedMsg:
		m.statusMsg = fmt.Sprintf("Successfully deleted %s", msg.Desc)
		m.isError = false
		return m, tea.Batch(
			fetchSubmissionsCmd(m.pool, ""),
			m.kanbanModel.FetchApplicationsCmd(),
		)

	case JobDeletedMsg:
		m.statusMsg = fmt.Sprintf("Successfully deleted %s", msg.Desc)
		m.isError = false
		return m, fetchJobsCmd(m.pool)

	case OrgDeletedMsg:
		m.statusMsg = fmt.Sprintf("Successfully deleted %s", msg.Desc)
		m.isError = false
		return m, fetchOrgsCmd(m.pool)

	case AppErrMsg:
		m.statusMsg = fmt.Sprintf("Error: %v", msg.Err)
		m.isError = true
		return m, nil

	case tea.KeyMsg:
		// 0. Deletion Confirmation Modal Handling
		if m.showDeleteConfirm {
			switch msg.String() {
			case "y", "Y":
				m.showDeleteConfirm = false
				return m, deleteEntityCmd(m.pool, m.deleteTargetType, m.deleteTargetID, m.deleteTargetDesc)
			case "n", "N", "esc", "q":
				m.showDeleteConfirm = false
				m.statusMsg = "Deletion cancelled"
				m.isError = false
				return m, nil
			}
			return m, nil
		}

		// 1. Form Modal Handling
		if m.showFormModal {
			m.formModal, cmd = m.formModal.Update(msg)
			return m, cmd
		}

		// 2. Full Details Modal Handling
		if m.showModal {
			switch msg.String() {
			case "esc", "q":
				m.showModal = false
				return m, nil
			}
			m.detailViewport, cmd = m.detailViewport.Update(msg)
			return m, cmd
		}

		// 3. Global Search Bar Handling
		if m.isSearching {
			switch msg.String() {
			case "esc":
				m.isSearching = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.isSearching = false
				m.searchInput.Blur()
				m.searchQuery = strings.TrimSpace(m.searchInput.Value())
				if m.activeTab == TabProfiles {
					return m, fetchProfilesCmd(m.pool, m.searchQuery)
				}
				m.applySortingAndFilters()
				m.statusMsg = fmt.Sprintf("Search applied: '%s'", m.searchQuery)
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		// 4. Overhauled Column Filter Modal Handling
		if m.isColFilterModal {
			cols := m.getFilterableColumnsForActiveTab()
			if m.colFilterColIdx >= len(cols) {
				m.colFilterColIdx = 0
			}
			activeCol := cols[m.colFilterColIdx]

			if m.colFilterSearching {
				switch msg.String() {
				case "esc", "enter":
					m.colFilterSearching = false
					m.colFilterSearchInput.Blur()
					return m, nil
				default:
					m.colFilterSearchInput, cmd = m.colFilterSearchInput.Update(msg)
					m.colFilterValIdx = 0
					return m, cmd
				}
			}

			distinctVals := m.getFilteredDistinctValues(activeCol)

			switch msg.String() {
			case "esc":
				m.isColFilterModal = false
				return m, nil

			case "tab", "right", "l":
				m.colFilterColIdx = (m.colFilterColIdx + 1) % len(cols)
				m.colFilterValIdx = 0
				m.colFilterSearchInput.SetValue("")
				return m, nil

			case "shift+tab", "left", "h":
				m.colFilterColIdx = (m.colFilterColIdx - 1 + len(cols)) % len(cols)
				m.colFilterValIdx = 0
				m.colFilterSearchInput.SetValue("")
				return m, nil

			case "down", "j":
				if len(distinctVals) > 0 {
					m.colFilterValIdx = (m.colFilterValIdx + 1) % len(distinctVals)
				}
				return m, nil

			case "up", "k":
				if len(distinctVals) > 0 {
					m.colFilterValIdx = (m.colFilterValIdx - 1 + len(distinctVals)) % len(distinctVals)
				}
				return m, nil

			case " ":
				// Toggle checkbox on current value
				if len(distinctVals) > 0 && m.colFilterValIdx < len(distinctVals) {
					val := distinctVals[m.colFilterValIdx]
					if m.colFilters[activeCol] == nil {
						m.colFilters[activeCol] = make(map[string]bool)
					}
					m.colFilters[activeCol][val] = !m.colFilters[activeCol][val]
					if !m.colFilters[activeCol][val] {
						delete(m.colFilters[activeCol], val)
						if len(m.colFilters[activeCol]) == 0 {
							delete(m.colFilters, activeCol)
						}
					}
				}
				return m, nil

			case "/":
				m.colFilterSearching = true
				m.colFilterSearchInput.Focus()
				return m, textinput.Blink

			case "enter":
				m.isColFilterModal = false
				m.applySortingAndFilters()
				m.statusMsg = m.getActiveFiltersStatusSummary()
				return m, nil
			}
			return m, nil
		}

		// 5. Global Hotkeys & Navigation
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "1":
			m.activeTab = TabProfiles
			m.statusMsg = "Switched to [1] Profiles"
			m.recalculateAvailableHeight()
			return m, nil

		case "2":
			m.activeTab = TabApplications
			m.statusMsg = "Switched to [2] Applications"
			m.recalculateAvailableHeight()
			return m, nil

		case "3":
			m.activeTab = TabJobs
			m.statusMsg = "Switched to [3] Jobs"
			m.recalculateAvailableHeight()
			return m, nil

		case "4":
			m.activeTab = TabOrganisations
			m.statusMsg = "Switched to [4] Organisations"
			m.recalculateAvailableHeight()
			return m, nil

		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
			m.recalculateAvailableHeight()
			return m, nil

		case "shift+tab":
			m.activeTab = (m.activeTab + 3) % 4
			m.recalculateAvailableHeight()
			return m, nil

		case "/":
			m.isSearching = true
			m.searchInput.Focus()
			m.searchInput.SetValue(m.searchQuery)
			return m, nil

		case "r":
			// Conflict resolution: 'r' strictly resets search and column filters
			m.searchQuery = ""
			m.colFilters = make(map[string]map[string]bool)
			m.statusMsg = "Active filters reset (showing full dataset)"
			if m.activeTab == TabProfiles {
				cmds = append(cmds, fetchProfilesCmd(m.pool, ""))
			}
			m.applySortingAndFilters()
			return m, tea.Batch(cmds...)

		case "R", "ctrl+r":
			// Conflict resolution: 'R' (Shift+R) and ctrl+r reload database
			m.statusMsg = "Reloading all views from PostgreSQL..."
			return m, tea.Batch(
				m.kanbanModel.FetchApplicationsCmd(),
				fetchProfilesCmd(m.pool, m.searchQuery),
				fetchSubmissionsCmd(m.pool, ""),
				fetchJobsCmd(m.pool),
				fetchOrgsCmd(m.pool),
			)

		case "s":
			m.cycleSortColumn()
			m.applySortingAndFilters()
			return m, nil

		case "S":
			m.toggleSortDirection()
			m.applySortingAndFilters()
			return m, nil

		case "f":
			// Open Column/Field Filter Modal
			m.isColFilterModal = true
			m.colFilterColIdx = 0
			m.colFilterValIdx = 0
			m.colFilterSearching = false
			m.colFilterSearchInput.SetValue("")
			return m, nil

		case "[":
			// Universal horizontal column scroll left for all tabs
			if m.tabColOffsets[m.activeTab] > 0 {
				m.tabColOffsets[m.activeTab]--
				m.applySortingAndFilters()
			}
			return m, nil

		case "]":
			// Universal horizontal column scroll right for all tabs
			all := m.getAllColumnsForActiveTab()
			if m.tabColOffsets[m.activeTab] < len(all)-2 {
				m.tabColOffsets[m.activeTab]++
				m.applySortingAndFilters()
			}
			return m, nil

		case "a":
			m.openAddForm()
			return m, nil

		case "e":
			m.openEditForm()
			return m, nil

		case "x", "D", "delete":
			m.openDeleteConfirm()
			return m, nil

		case "p":
			m.showSplitPane = !m.showSplitPane
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}

		case "v", "t":
			if m.activeTab == TabApplications {
				if m.appViewMode == AppViewKanban {
					m.appViewMode = AppViewTable
					m.statusMsg = "Applications: Full Table View [use '[' / ']' to scroll columns]"
				} else {
					m.appViewMode = AppViewKanban
					m.statusMsg = "Applications: 4-Stage Kanban Board"
				}
				return m, nil
			}

		case "enter", "d":
			return m, m.openDetailModalCmd()
		}
	}

	// Delegate to active view
	switch m.activeTab {
	case TabProfiles:
		m.profilesTable, cmd = m.profilesTable.Update(msg)
		cmds = append(cmds, cmd)

	case TabApplications:
		if m.appViewMode == AppViewKanban {
			var kModel tea.Model
			kModel, cmd = m.kanbanModel.Update(msg)
			m.kanbanModel = kModel.(kanban.Model)
			cmds = append(cmds, cmd)
		} else {
			m.submissionsTable, cmd = m.submissionsTable.Update(msg)
			cmds = append(cmds, cmd)
		}

	case TabJobs:
		m.jobsTable, cmd = m.jobsTable.Update(msg)
		cmds = append(cmds, cmd)

	case TabOrganisations:
		m.orgsTable, cmd = m.orgsTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) openDeleteConfirm() {
	switch m.activeTab {
	case TabProfiles:
		idx := m.profilesTable.Cursor()
		if idx >= 0 && idx < len(m.profiles) {
			m.deleteTargetType = EntityCandidate
			m.deleteTargetID = m.profiles[idx].ID
			m.deleteTargetDesc = fmt.Sprintf("Candidate '%s' (%s)", m.profiles[idx].Name, m.profiles[idx].CandidateCode)
			m.showDeleteConfirm = true
		} else {
			m.statusMsg = "No candidate selected to delete"
			m.isError = true
		}
	case TabApplications:
		if m.appViewMode == AppViewKanban {
			card := m.kanbanModel.SelectedCard()
			if card != nil {
				m.deleteTargetType = EntityApplication
				m.deleteTargetID = card.ApplicationID
				m.deleteTargetDesc = fmt.Sprintf("Application of '%s' for job %s", card.CandidateName, card.JobCode)
				m.showDeleteConfirm = true
			} else {
				m.statusMsg = "No application selected in current Kanban stage"
				m.isError = true
			}
		} else {
			idx := m.submissionsTable.Cursor()
			if idx >= 0 && idx < len(m.submissions) {
				m.deleteTargetType = EntityApplication
				m.deleteTargetID = m.submissions[idx].ApplicationID
				m.deleteTargetDesc = fmt.Sprintf("Application of '%s' for job %s", m.submissions[idx].CandidateName, m.submissions[idx].JobCode)
				m.showDeleteConfirm = true
			} else {
				m.statusMsg = "No application selected to delete"
				m.isError = true
			}
		}
	case TabJobs:
		idx := m.jobsTable.Cursor()
		if idx >= 0 && idx < len(m.jobs) {
			m.deleteTargetType = EntityJob
			m.deleteTargetID = m.jobs[idx].ID
			m.deleteTargetDesc = fmt.Sprintf("Job '%s' (%s)", m.jobs[idx].JobTitle, m.jobs[idx].JobCode)
			m.showDeleteConfirm = true
		} else {
			m.statusMsg = "No job selected to delete"
			m.isError = true
		}
	case TabOrganisations:
		idx := m.orgsTable.Cursor()
		if idx >= 0 && idx < len(m.orgs) {
			m.deleteTargetType = EntityOrganisation
			m.deleteTargetID = m.orgs[idx].ID
			m.deleteTargetDesc = fmt.Sprintf("Organisation '%s' (%s)", m.orgs[idx].OrgName, m.orgs[idx].OrgCode)
			m.showDeleteConfirm = true
		} else {
			m.statusMsg = "No organisation selected to delete"
			m.isError = true
		}
	}
}

func (m *AppModel) openAddForm() {
	switch m.activeTab {
	case TabProfiles:
		m.formModal = NewCandidateForm(nil)
	case TabApplications:
		m.formModal = NewApplicationForm(nil, "JOB1", "CAN-001")
	case TabJobs:
		m.formModal = NewJobForm(nil)
	case TabOrganisations:
		m.formModal = NewOrganisationForm(nil)
	}
	m.showFormModal = true
}

func (m *AppModel) openEditForm() {
	switch m.activeTab {
	case TabProfiles:
		idx := m.profilesTable.Cursor()
		if idx >= 0 && idx < len(m.profiles) {
			m.formModal = NewCandidateForm(&m.profiles[idx])
			m.showFormModal = true
		}
	case TabApplications:
		if m.appViewMode == AppViewKanban {
			card := m.kanbanModel.SelectedCard()
			if card != nil {
				sub := models.ApplicationSubmissionRow{
					ApplicationID:    card.ApplicationID,
					CandidateID:      card.CandidateID,
					JobCode:          card.JobCode,
					CandidateCode:    card.CandidateCode,
					CandidateName:    card.CandidateName,
					CurrentOrg:       card.CurrentOrg,
					Designation:      card.Designation,
					Products:         card.Products,
					ExperienceYears:  card.ExperienceYears,
					ExpectedCTC:      card.ExpectedCTC,
					NoticePeriodDays: card.NoticePeriodDays,
					Status:           card.Status,
					Comments:         card.LatestComment,
				}
				m.formModal = NewApplicationForm(&sub, card.JobCode, card.CandidateCode)
				m.showFormModal = true
			}
		} else {
			idx := m.submissionsTable.Cursor()
			if idx >= 0 && idx < len(m.submissions) {
				m.formModal = NewApplicationForm(&m.submissions[idx], m.submissions[idx].JobCode, m.submissions[idx].CandidateCode)
				m.showFormModal = true
			}
		}
	case TabJobs:
		idx := m.jobsTable.Cursor()
		if idx >= 0 && idx < len(m.jobs) {
			m.formModal = NewJobForm(&m.jobs[idx])
			m.showFormModal = true
		}
	case TabOrganisations:
		idx := m.orgsTable.Cursor()
		if idx >= 0 && idx < len(m.orgs) {
			m.formModal = NewOrganisationForm(&m.orgs[idx])
			m.showFormModal = true
		}
	}
}

func (m *AppModel) cycleSortColumn() {
	switch m.activeTab {
	case TabProfiles:
		m.profileSortCol = (m.profileSortCol + 1) % 4
	case TabApplications:
		m.appSortCol = (m.appSortCol + 1) % 5
	case TabJobs:
		m.jobSortCol = (m.jobSortCol + 1) % 5
	case TabOrganisations:
		m.orgSortCol = (m.orgSortCol + 1) % 3
	}
}

func (m *AppModel) toggleSortDirection() {
	switch m.activeTab {
	case TabProfiles:
		m.profileSortAsc = !m.profileSortAsc
	case TabApplications:
		m.appSortAsc = !m.appSortAsc
	case TabJobs:
		m.jobSortAsc = !m.jobSortAsc
	case TabOrganisations:
		m.orgSortAsc = !m.orgSortAsc
	}
}

func (m *AppModel) applySortingAndFilters() {
	availWidth := m.width - 4
	if availWidth < 40 {
		availWidth = 40
	}

	// 1. Sort & Filter Profiles
	var filteredProfiles []models.CandidateProfile
	for _, p := range m.profiles {
		rowMap := map[string]string{
			"Candidate ID":              p.CandidateCode,
			"Candidate Name":            p.Name,
			"Contact Number":            p.Phone,
			"Email ID":                  p.Email,
			"Location":                  p.Location,
			"Educational Qualification": p.Qualification,
			"Date of Birth":             p.DOB,
		}
		if m.matchesColFilters(rowMap) {
			filteredProfiles = append(filteredProfiles, p)
		}
	}

	sort.SliceStable(filteredProfiles, func(i, j int) bool {
		var less bool
		switch m.profileSortCol {
		case 0:
			less = filteredProfiles[i].CandidateCode < filteredProfiles[j].CandidateCode
		case 1:
			less = strings.ToLower(filteredProfiles[i].Name) < strings.ToLower(filteredProfiles[j].Name)
		case 2:
			less = strings.ToLower(filteredProfiles[i].Location) < strings.ToLower(filteredProfiles[j].Location)
		case 3:
			less = strings.ToLower(filteredProfiles[i].Qualification) < strings.ToLower(filteredProfiles[j].Qualification)
		}
		if !m.profileSortAsc {
			return !less
		}
		return less
	})

	visProfileCols := visibleColumns(allProfileColumns(), m.tabColOffsets[TabProfiles], availWidth)
	var pRows []table.Row
	for _, p := range filteredProfiles {
		rowMap := map[string]string{
			"Candidate ID":              p.CandidateCode,
			"Candidate Name":            p.Name,
			"Contact Number":            p.Phone,
			"Email ID":                  p.Email,
			"Location":                  p.Location,
			"Educational Qualification": p.Qualification,
			"Date of Birth":             p.DOB,
		}
		var row table.Row
		for _, col := range visProfileCols {
			val := rowMap[col.Title]
			row = append(row, truncateString(val, col.Width))
		}
		pRows = append(pRows, row)
	}
	updateTableData(&m.profilesTable, visProfileCols, pRows)

	// 2. Sort & Filter Submissions
	var filteredSubmissions []models.ApplicationSubmissionRow
	for _, s := range m.submissions {
		rowMap := map[string]string{
			"Job ID":          s.JobCode,
			"Candidate ID":    s.CandidateCode,
			"Candidate Name":  s.CandidateName,
			"Designation":     s.Designation,
			"Exp":             fmt.Sprintf("%dyr", s.ExperienceYears),
			"CTC":             fmt.Sprintf("$%dk", s.CTC/1000),
			"Exp CTC":         fmt.Sprintf("$%dk", s.ExpectedCTC/1000),
			"Status":          string(s.Status),
			"Current Org":     s.CurrentOrg,
			"Products":        s.Products,
			"Notice":          fmt.Sprintf("%dd", s.NoticePeriodDays),
			"Location":        s.Location,
			"Qualification":   s.Qualification,
			"Latest Comments": s.Comments,
		}
		if m.matchesColFilters(rowMap) {
			filteredSubmissions = append(filteredSubmissions, s)
		}
	}

	sort.SliceStable(filteredSubmissions, func(i, j int) bool {
		var less bool
		switch m.appSortCol {
		case 0:
			less = filteredSubmissions[i].JobCode < filteredSubmissions[j].JobCode
		case 1:
			less = strings.ToLower(filteredSubmissions[i].CandidateName) < strings.ToLower(filteredSubmissions[j].CandidateName)
		case 2:
			less = strings.ToLower(string(filteredSubmissions[i].Status)) < strings.ToLower(string(filteredSubmissions[j].Status))
		case 3:
			less = filteredSubmissions[i].ExpectedCTC < filteredSubmissions[j].ExpectedCTC
		case 4:
			less = filteredSubmissions[i].ExperienceYears < filteredSubmissions[j].ExperienceYears
		}
		if !m.appSortAsc {
			return !less
		}
		return less
	})

	visSubCols := visibleColumns(allSubmissionColumns(), m.tabColOffsets[TabApplications], availWidth)
	var sRows []table.Row
	for _, s := range filteredSubmissions {
		rowMap := map[string]string{
			"Job ID":          s.JobCode,
			"Candidate ID":    s.CandidateCode,
			"Candidate Name":  s.CandidateName,
			"Designation":     s.Designation,
			"Exp":             fmt.Sprintf("%dyr", s.ExperienceYears),
			"CTC":             fmt.Sprintf("$%dk", s.CTC/1000),
			"Exp CTC":         fmt.Sprintf("$%dk", s.ExpectedCTC/1000),
			"Status":          string(s.Status),
			"Current Org":     s.CurrentOrg,
			"Products":        s.Products,
			"Notice":          fmt.Sprintf("%dd", s.NoticePeriodDays),
			"Location":        s.Location,
			"Qualification":   s.Qualification,
			"Latest Comments": s.Comments,
		}
		var row table.Row
		for _, col := range visSubCols {
			val := rowMap[col.Title]
			row = append(row, truncateString(val, col.Width))
		}
		sRows = append(sRows, row)
	}
	updateTableData(&m.submissionsTable, visSubCols, sRows)

	// 3. Sort & Filter Jobs
	var filteredJobs []models.JobPosting
	for _, j := range m.jobs {
		rowMap := map[string]string{
			"Job ID":          j.JobCode,
			"Client":          j.ClientName,
			"Job Title":       j.JobTitle,
			"Role":            j.Role,
			"Department":      j.Department,
			"Sub-department":  j.SubDepartment,
			"Location":        j.Location,
			"Req Date":        j.RequisitionDate,
			"Positions":       fmt.Sprintf("%d", j.NumPositions),
			"Profiles Shared": fmt.Sprintf("%d", j.NumShared),
		}
		if m.matchesColFilters(rowMap) {
			filteredJobs = append(filteredJobs, j)
		}
	}

	sort.SliceStable(filteredJobs, func(i, j int) bool {
		var less bool
		switch m.jobSortCol {
		case 0:
			less = filteredJobs[i].JobCode < filteredJobs[j].JobCode
		case 1:
			less = strings.ToLower(filteredJobs[i].ClientName) < strings.ToLower(filteredJobs[j].ClientName)
		case 2:
			less = strings.ToLower(filteredJobs[i].JobTitle) < strings.ToLower(filteredJobs[j].JobTitle)
		case 3:
			less = filteredJobs[i].RequisitionDate < filteredJobs[j].RequisitionDate
		case 4:
			less = filteredJobs[i].NumPositions < filteredJobs[j].NumPositions
		}
		if !m.jobSortAsc {
			return !less
		}
		return less
	})

	visJobCols := visibleColumns(allJobColumns(), m.tabColOffsets[TabJobs], availWidth)
	var jRows []table.Row
	for _, j := range filteredJobs {
		rowMap := map[string]string{
			"Job ID":          j.JobCode,
			"Client":          j.ClientName,
			"Job Title":       j.JobTitle,
			"Role":            j.Role,
			"Department":      j.Department,
			"Sub-department":  j.SubDepartment,
			"Location":        j.Location,
			"Req Date":        j.RequisitionDate,
			"Positions":       fmt.Sprintf("%d", j.NumPositions),
			"Profiles Shared": fmt.Sprintf("%d", j.NumShared),
		}
		var row table.Row
		for _, col := range visJobCols {
			val := rowMap[col.Title]
			row = append(row, truncateString(val, col.Width))
		}
		jRows = append(jRows, row)
	}
	updateTableData(&m.jobsTable, visJobCols, jRows)

	// 4. Sort & Filter Organisations
	var filteredOrgs []models.Organisation
	for _, o := range m.orgs {
		rowMap := map[string]string{
			"Organisation Name": o.OrgName,
			"Org. Code":         o.OrgCode,
			"Industry":          o.Industry,
			"Sub Category":      o.SubCategory,
		}
		if m.matchesColFilters(rowMap) {
			filteredOrgs = append(filteredOrgs, o)
		}
	}

	sort.SliceStable(filteredOrgs, func(i, j int) bool {
		var less bool
		switch m.orgSortCol {
		case 0:
			less = strings.ToLower(filteredOrgs[i].OrgName) < strings.ToLower(filteredOrgs[j].OrgName)
		case 1:
			less = strings.ToLower(filteredOrgs[i].OrgCode) < strings.ToLower(filteredOrgs[j].OrgCode)
		case 2:
			less = strings.ToLower(filteredOrgs[i].Industry) < strings.ToLower(filteredOrgs[j].Industry)
		}
		if !m.orgSortAsc {
			return !less
		}
		return less
	})

	visOrgCols := visibleColumns(allOrgColumns(), m.tabColOffsets[TabOrganisations], availWidth)
	var oRows []table.Row
	for _, o := range filteredOrgs {
		rowMap := map[string]string{
			"Organisation Name": o.OrgName,
			"Org. Code":         o.OrgCode,
			"Industry":          o.Industry,
			"Sub Category":      o.SubCategory,
		}
		var row table.Row
		for _, col := range visOrgCols {
			val := rowMap[col.Title]
			row = append(row, truncateString(val, col.Width))
		}
		oRows = append(oRows, row)
	}
	updateTableData(&m.orgsTable, visOrgCols, oRows)
}

// matchesColFilters performs multi-column faceted filtering (AND across columns, OR within column)
func (m AppModel) matchesColFilters(rowMap map[string]string) bool {
	if len(m.colFilters) == 0 {
		return true
	}

	for colName, checkedMap := range m.colFilters {
		if len(checkedMap) == 0 {
			continue
		}
		cellVal, exists := rowMap[colName]
		if !exists {
			continue
		}

		matched := false
		for val := range checkedMap {
			if strings.EqualFold(strings.TrimSpace(cellVal), strings.TrimSpace(val)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (m AppModel) getFilterableColumnsForActiveTab() []string {
	cols := m.getAllColumnsForActiveTab()
	var names []string
	for _, c := range cols {
		names = append(names, c.Title)
	}
	return names
}

func (m AppModel) getDistinctValuesForColumn(colName string) []string {
	seen := make(map[string]bool)
	var vals []string

	switch m.activeTab {
	case TabProfiles:
		for _, p := range m.profiles {
			var v string
			switch colName {
			case "Candidate ID":
				v = p.CandidateCode
			case "Candidate Name":
				v = p.Name
			case "Contact Number":
				v = p.Phone
			case "Email ID":
				v = p.Email
			case "Location":
				v = p.Location
			case "Educational Qualification":
				v = p.Qualification
			case "Date of Birth":
				v = p.DOB
			}
			v = strings.TrimSpace(v)
			if v != "" && !seen[v] {
				seen[v] = true
				vals = append(vals, v)
			}
		}

	case TabApplications:
		for _, s := range m.submissions {
			var v string
			switch colName {
			case "Job ID":
				v = s.JobCode
			case "Candidate ID":
				v = s.CandidateCode
			case "Candidate Name":
				v = s.CandidateName
			case "Designation":
				v = s.Designation
			case "Exp":
				v = fmt.Sprintf("%dyr", s.ExperienceYears)
			case "CTC":
				v = fmt.Sprintf("$%dk", s.CTC/1000)
			case "Exp CTC":
				v = fmt.Sprintf("$%dk", s.ExpectedCTC/1000)
			case "Status":
				v = string(s.Status)
			case "Current Org":
				v = s.CurrentOrg
			case "Products":
				v = s.Products
			case "Notice":
				v = fmt.Sprintf("%dd", s.NoticePeriodDays)
			case "Location":
				v = s.Location
			case "Qualification":
				v = s.Qualification
			case "Latest Comments":
				v = s.Comments
			}
			v = strings.TrimSpace(v)
			if v != "" && !seen[v] {
				seen[v] = true
				vals = append(vals, v)
			}
		}

	case TabJobs:
		for _, j := range m.jobs {
			var v string
			switch colName {
			case "Job ID":
				v = j.JobCode
			case "Client":
				v = j.ClientName
			case "Job Title":
				v = j.JobTitle
			case "Role":
				v = j.Role
			case "Department":
				v = j.Department
			case "Sub-department":
				v = j.SubDepartment
			case "Location":
				v = j.Location
			case "Req Date":
				v = j.RequisitionDate
			case "Positions":
				v = fmt.Sprintf("%d", j.NumPositions)
			case "Profiles Shared":
				v = fmt.Sprintf("%d", j.NumShared)
			}
			v = strings.TrimSpace(v)
			if v != "" && !seen[v] {
				seen[v] = true
				vals = append(vals, v)
			}
		}

	case TabOrganisations:
		for _, o := range m.orgs {
			var v string
			switch colName {
			case "Organisation Name":
				v = o.OrgName
			case "Org. Code":
				v = o.OrgCode
			case "Industry":
				v = o.Industry
			case "Sub Category":
				v = o.SubCategory
			}
			v = strings.TrimSpace(v)
			if v != "" && !seen[v] {
				seen[v] = true
				vals = append(vals, v)
			}
		}
	}

	sort.Strings(vals)
	return vals
}

func (m AppModel) getFilteredDistinctValues(colName string) []string {
	allVals := m.getDistinctValuesForColumn(colName)
	query := strings.ToLower(strings.TrimSpace(m.colFilterSearchInput.Value()))
	if query == "" {
		return allVals
	}

	var filtered []string
	for _, v := range allVals {
		if strings.Contains(strings.ToLower(v), query) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (m AppModel) getActiveFiltersStatusSummary() string {
	if len(m.colFilters) == 0 {
		return "Filters: None"
	}

	var parts []string
	for col, valMap := range m.colFilters {
		if len(valMap) == 0 {
			continue
		}
		var valList []string
		for v := range valMap {
			valList = append(valList, v)
		}
		sort.Strings(valList)
		parts = append(parts, fmt.Sprintf("%s=[%s]", col, strings.Join(valList, ", ")))
	}
	if len(parts) == 0 {
		return "Filters: None"
	}
	return fmt.Sprintf("Filtered: %s (press 'r' to clear)", strings.Join(parts, " • "))
}

func (m *AppModel) openDetailModalCmd() tea.Cmd {
	m.showModal = true
	m.activeAppNotes = nil

	switch m.activeTab {
	case TabProfiles:
		idx := m.profilesTable.Cursor()
		if idx >= 0 && idx < len(m.profiles) {
			p := m.profiles[idx]
			m.modalTitle = fmt.Sprintf("Candidate: %s [%s]", p.Name, p.CandidateCode)
			var sb strings.Builder
			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("CANDIDATE MASTER PROFILE\n"))
			sb.WriteString(strings.Repeat("─", 60) + "\n")
			sb.WriteString(fmt.Sprintf("• Candidate ID:  %s\n", p.CandidateCode))
			sb.WriteString(fmt.Sprintf("• Name:          %s\n", p.Name))
			sb.WriteString(fmt.Sprintf("• Email:         %s\n", p.Email))
			sb.WriteString(fmt.Sprintf("• Phone:         %s\n", p.Phone))
			sb.WriteString(fmt.Sprintf("• Location:      %s\n", p.Location))
			sb.WriteString(fmt.Sprintf("• Qualification: %s\n", p.Qualification))
			sb.WriteString(fmt.Sprintf("• Date of Birth: %s\n\n", p.DOB))

			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5A3")).Render("📄 FULL RESUME TEXT (PostgreSQL tsvector Indexed):\n"))
			sb.WriteString(strings.Repeat("─", 60) + "\n")
			sb.WriteString(p.ResumeText + "\n\n")

			m.detailViewport.SetContent(sb.String())
		}

	case TabApplications:
		if m.appViewMode == AppViewKanban {
			card := m.kanbanModel.SelectedCard()
			if card != nil {
				m.modalTitle = fmt.Sprintf("Application: %s -> %s", card.CandidateName, card.JobTitle)
				m.rebuildModalContent()
				return fetchNotesCmd(m.pool, card.ApplicationID)
			}
		} else {
			idx := m.submissionsTable.Cursor()
			if idx >= 0 && idx < len(m.submissions) {
				sub := m.submissions[idx]
				m.modalTitle = fmt.Sprintf("Submission: %s -> %s", sub.CandidateName, sub.JobTitle)
				m.rebuildModalContent()
				return fetchNotesCmd(m.pool, sub.ApplicationID)
			}
		}

	case TabJobs:
		idx := m.jobsTable.Cursor()
		if idx >= 0 && idx < len(m.jobs) {
			j := m.jobs[idx]
			m.modalTitle = fmt.Sprintf("Job Requisition: %s [%s]", j.JobTitle, j.JobCode)
			var sb strings.Builder
			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("JOB REQUISITION DETAILS\n"))
			sb.WriteString(strings.Repeat("─", 60) + "\n")
			sb.WriteString(fmt.Sprintf("• Job ID:          %s\n", j.JobCode))
			sb.WriteString(fmt.Sprintf("• Client:          %s\n", j.ClientName))
			sb.WriteString(fmt.Sprintf("• Job Title:       %s\n", j.JobTitle))
			sb.WriteString(fmt.Sprintf("• Role:            %s\n", j.Role))
			sb.WriteString(fmt.Sprintf("• Department:      %s / %s\n", j.Department, j.SubDepartment))
			sb.WriteString(fmt.Sprintf("• Location:        %s\n", j.Location))
			sb.WriteString(fmt.Sprintf("• Requisition:     %s\n", j.RequisitionDate))
			sb.WriteString(fmt.Sprintf("• Positions Open:  %d\n", j.NumPositions))
			sb.WriteString(fmt.Sprintf("• Profiles Shared: %d\n", j.NumShared))
			m.detailViewport.SetContent(sb.String())
		}

	case TabOrganisations:
		idx := m.orgsTable.Cursor()
		if idx >= 0 && idx < len(m.orgs) {
			o := m.orgs[idx]
			m.modalTitle = fmt.Sprintf("Client: %s [%s]", o.OrgName, o.OrgCode)
			var sb strings.Builder
			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("ORGANISATION PROFILE\n"))
			sb.WriteString(strings.Repeat("─", 60) + "\n")
			sb.WriteString(fmt.Sprintf("• Organisation: %s\n", o.OrgName))
			sb.WriteString(fmt.Sprintf("• Org Code:     %s\n", o.OrgCode))
			sb.WriteString(fmt.Sprintf("• Industry:     %s\n", o.Industry))
			sb.WriteString(fmt.Sprintf("• Sub Category: %s\n", o.SubCategory))
			m.detailViewport.SetContent(sb.String())
		}
	}

	return nil
}

func (m *AppModel) rebuildModalContent() {
	if m.activeTab != TabApplications {
		return
	}

	var sb strings.Builder
	if m.appViewMode == AppViewKanban {
		card := m.kanbanModel.SelectedCard()
		if card == nil {
			return
		}
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("APPLICATION & SUBMISSION PROFILE\n"))
		sb.WriteString(strings.Repeat("─", 70) + "\n")
		sb.WriteString(fmt.Sprintf("• Candidate:       %s [%s]\n", card.CandidateName, card.CandidateCode))
		sb.WriteString(fmt.Sprintf("• Target Job:      %s (Job ID: %s)\n", card.JobTitle, card.JobCode))
		sb.WriteString(fmt.Sprintf("• Current Org:     %s (%s)\n", card.CurrentOrg, card.Designation))
		sb.WriteString(fmt.Sprintf("• Experience:      %d years  •  Primary Skill: %s\n", card.ExperienceYears, card.Skill))
		sb.WriteString(fmt.Sprintf("• Expected Comp:   $%dk  •  Notice Period: %d days\n", card.ExpectedCTC/1000, card.NoticePeriodDays))
		sb.WriteString(fmt.Sprintf("• Current Status:  %s\n", card.Status.Title()))
		if card.Products != "" {
			sb.WriteString(fmt.Sprintf("• Product Line(s): %s\n", card.Products))
		}
		sb.WriteString("\n")

		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5A3")).Render("📄 CANDIDATE RESUME:\n"))
		sb.WriteString(strings.Repeat("─", 70) + "\n")
		sb.WriteString(card.ResumeText + "\n\n")

	} else {
		idx := m.submissionsTable.Cursor()
		if idx < 0 || idx >= len(m.submissions) {
			return
		}
		sub := m.submissions[idx]
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("APPLICATION & SUBMISSION PROFILE\n"))
		sb.WriteString(strings.Repeat("─", 70) + "\n")
		sb.WriteString(fmt.Sprintf("• Candidate:       %s [%s]\n", sub.CandidateName, sub.CandidateCode))
		sb.WriteString(fmt.Sprintf("• Target Job:      %s (Job ID: %s)\n", sub.JobTitle, sub.JobCode))
		sb.WriteString(fmt.Sprintf("• Department:      %s / %s\n", sub.Department, sub.SubDepartment))
		sb.WriteString(fmt.Sprintf("• Contact:         %s | %s | %s\n", sub.Phone, sub.Email, sub.Location))
		sb.WriteString(fmt.Sprintf("• Current Org:     %s (%s)\n", sub.CurrentOrg, sub.Designation))
		sb.WriteString(fmt.Sprintf("• Exp & Comp:      %dyr  •  Current: $%dk  •  Exp: $%dk  •  Notice: %ddays\n",
			sub.ExperienceYears, sub.CTC/1000, sub.ExpectedCTC/1000, sub.NoticePeriodDays))
		sb.WriteString(fmt.Sprintf("• Current Status:  %s\n", sub.Status.Title()))
		if sub.Products != "" {
			sb.WriteString(fmt.Sprintf("• Product Line(s): %s\n", sub.Products))
		}
		sb.WriteString("\n")

		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E5A3")).Render("📄 CANDIDATE RESUME:\n"))
		sb.WriteString(strings.Repeat("─", 70) + "\n")
		sb.WriteString(sub.ResumeText + "\n\n")
	}

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAA33")).Render("📝 CHRONOLOGICAL NOTES & INTERVIEW FEEDBACK:\n"))
	sb.WriteString(strings.Repeat("─", 70) + "\n")
	if len(m.activeAppNotes) == 0 {
		sb.WriteString("  No interview comments or notes recorded yet.\n")
	} else {
		for _, note := range m.activeAppNotes {
			tagColor := "#00E5A3"
			if note.NoteType == "Feedback" {
				tagColor = "#FF5F87"
			}
			tag := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(tagColor)).Render(fmt.Sprintf("[%s]", note.NoteType))
			timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render(note.CreatedAt)
			sb.WriteString(fmt.Sprintf("%s  %s\n  %s\n\n", tag, timeStr, note.Content))
		}
	}

	m.detailViewport.SetContent(sb.String())
}

func (m AppModel) renderSplitPane() string {
	paneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5A56E0")).
		Padding(0, 1).
		Width(m.width - 2).
		Height(6)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D7D7")).
		Render("🔍 QUICK PREVIEW [Press 'p' to toggle]")

	var body string

	switch m.activeTab {
	case TabProfiles:
		idx := m.profilesTable.Cursor()
		if idx >= 0 && idx < len(m.profiles) {
			p := m.profiles[idx]
			body = fmt.Sprintf(
				"%s (%s) • %s • %s\nEducation: %s • DOB: %s\nResume Preview: %s",
				lipgloss.NewStyle().Bold(true).Render(p.Name),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render(p.CandidateCode),
				p.Email,
				p.Location,
				p.Qualification,
				p.DOB,
				lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#A0A0A0")).Render(truncateString(p.ResumeText, 140)),
			)
		} else {
			body = "No candidate selected."
		}

	case TabApplications:
		if m.appViewMode == AppViewKanban {
			card := m.kanbanModel.SelectedCard()
			if card != nil {
				comment := card.LatestComment
				if comment == "" {
					comment = "No comments recorded."
				}
				body = fmt.Sprintf(
					"%s [%s] -> %s (%s)\nRole: %s @ %s (%dyr) • Comp: $%dk • Notice: %dd\nProducts: %s\nLatest Note: %s",
					lipgloss.NewStyle().Bold(true).Render(card.CandidateName),
					card.CandidateCode,
					card.JobTitle,
					card.JobCode,
					card.Designation,
					card.CurrentOrg,
					card.ExperienceYears,
					card.ExpectedCTC/1000,
					card.NoticePeriodDays,
					lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA33")).Render(card.Products),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFA3")).Render(truncateString(comment, 120)),
				)
			} else {
				body = "No application card selected."
			}
		} else {
			idx := m.submissionsTable.Cursor()
			if idx >= 0 && idx < len(m.submissions) {
				s := m.submissions[idx]
				body = fmt.Sprintf(
					"%s [%s] -> %s (%s) • Status: %s\nRole: %s @ %s • Comp: Current $%dk / Exp $%dk\nProducts: %s\nLatest Note: %s",
					lipgloss.NewStyle().Bold(true).Render(s.CandidateName),
					s.CandidateCode,
					s.JobTitle,
					s.JobCode,
					s.Status.Title(),
					s.Designation,
					s.CurrentOrg,
					s.CTC/1000,
					s.ExpectedCTC/1000,
					lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA33")).Render(s.Products),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFA3")).Render(truncateString(s.Comments, 120)),
				)
			} else {
				body = "No submission row selected."
			}
		}

	case TabJobs:
		idx := m.jobsTable.Cursor()
		if idx >= 0 && idx < len(m.jobs) {
			j := m.jobs[idx]
			body = fmt.Sprintf(
				"%s (%s) for %s\nDepartment: %s / %s • Location: %s • Requisition Date: %s\nOpen Positions: %d • Profiles Shared: %d",
				lipgloss.NewStyle().Bold(true).Render(j.JobTitle),
				j.JobCode,
				j.ClientName,
				j.Department,
				j.SubDepartment,
				j.Location,
				j.RequisitionDate,
				j.NumPositions,
				j.NumShared,
			)
		} else {
			body = "No job requisition selected."
		}

	case TabOrganisations:
		idx := m.orgsTable.Cursor()
		if idx >= 0 && idx < len(m.orgs) {
			o := m.orgs[idx]
			body = fmt.Sprintf(
				"Organisation: %s [Code: %s]\nIndustry: %s\nSub Category: %s",
				lipgloss.NewStyle().Bold(true).Render(o.OrgName),
				o.OrgCode,
				o.Industry,
				o.SubCategory,
			)
		} else {
			body = "No organisation selected."
		}
	}

	return paneStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, body))
}

func (m AppModel) renderColFilterModal() string {
	boxWidth := 68
	if m.width > 0 && boxWidth > m.width-6 {
		boxWidth = m.width - 6
	}

	cols := m.getFilterableColumnsForActiveTab()
	if m.colFilterColIdx >= len(cols) {
		return ""
	}
	activeCol := cols[m.colFilterColIdx]
	distinctVals := m.getFilteredDistinctValues(activeCol)

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#00E5A3")).
		Padding(1, 2).
		Width(boxWidth)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#00E5A3")).
		Padding(0, 2).
		Render("🔎 MULTI-COLUMN FILTER")

	// 1. Column Selection Header
	colLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).
		Render(fmt.Sprintf("Active Column: < %s >  (Col %d of %d)", activeCol, m.colFilterColIdx+1, len(cols)))
	colHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Italic(true).
		Render("[Tab / ← / → to change column]")

	// 2. Type-to-Search Input
	searchHeader := "Search Values: (Press '/' to type filter)"
	if m.colFilterSearching {
		searchHeader = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFA3")).
			Render("Search Values: [Type query, Enter/Esc when done]")
	}
	searchBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#383838")).
		Padding(0, 1).
		Width(boxWidth - 8).
		Render(m.colFilterSearchInput.View())

	// 3. Values List (Windowed with Checkboxes)
	var valLines []string
	valLines = append(valLines, lipgloss.NewStyle().Bold(true).Underline(true).Render(fmt.Sprintf("Values for '%s':", activeCol)))

	if len(distinctVals) == 0 {
		valLines = append(valLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render("  (No matching values found)"))
	} else {
		// Window list to display max 8 items centered around cursor
		windowSize := 8
		startIdx := 0
		if m.colFilterValIdx >= windowSize {
			startIdx = m.colFilterValIdx - windowSize + 1
		}
		endIdx := startIdx + windowSize
		if endIdx > len(distinctVals) {
			endIdx = len(distinctVals)
		}

		for i := startIdx; i < endIdx; i++ {
			val := distinctVals[i]
			isCursor := (i == m.colFilterValIdx)
			isChecked := false
			if m.colFilters[activeCol] != nil && m.colFilters[activeCol][val] {
				isChecked = true
			}

			cursorMarker := "  "
			if isCursor {
				cursorMarker = "▸ "
			}

			checkMarker := "[ ] "
			if isChecked {
				checkMarker = "[✓] "
			}

			lineStr := cursorMarker + checkMarker + truncateString(val, boxWidth-16)

			var lineStyle lipgloss.Style
			if isCursor {
				lineStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFA3"))
			} else if isChecked {
				lineStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFCC00"))
			} else {
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#BBBBBB"))
			}

			valLines = append(valLines, lineStyle.Render(lineStr))
		}

		if len(distinctVals) > windowSize {
			valLines = append(valLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true).
				Render(fmt.Sprintf("  ... showing %d of %d values", endIdx-startIdx, len(distinctVals))))
		}
	}
	valuesBlock := lipgloss.JoinVertical(lipgloss.Left, valLines...)

	// 4. Active Selection Summary Box
	var summaryParts []string
	for col, valMap := range m.colFilters {
		if len(valMap) == 0 {
			continue
		}
		var list []string
		for v := range valMap {
			list = append(list, v)
		}
		sort.Strings(list)
		summaryParts = append(summaryParts, fmt.Sprintf("%s=[%s]", col, strings.Join(list, ", ")))
	}
	var summaryText string
	if len(summaryParts) > 0 {
		summaryText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFCC00")).
			Render("Selected Filters: " + truncateString(strings.Join(summaryParts, " • "), boxWidth-8))
	} else {
		summaryText = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).
			Render("Selected Filters: (None checked yet)")
	}

	// 5. Help Footer
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Italic(true).
		Render("[Space] Toggle [✓]  •  [Tab/←/→] Column  •  [/] Search  •  [Enter] Apply  •  [Esc] Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"\n",
		colLabel,
		colHint,
		"\n",
		searchHeader,
		searchBox,
		"\n",
		valuesBlock,
		"\n",
		summaryText,
		"\n",
		help,
	)

	return modalBox.Render(content)
}

func (m AppModel) renderDeleteConfirmModal() string {
	boxWidth := 56
	if m.width > 0 && m.width-10 < boxWidth {
		boxWidth = m.width - 10
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#E11D48")).
		Padding(0, 2).
		Render("⚠️  Confirm Deletion")

	desc := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FDA4AF")).
		Render(m.deleteTargetDesc)

	warningMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Render("Are you sure you want to permanently delete this record?\nThis action cannot be undone.")

	prompt := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FACC15")).
		Render("Press [y] to confirm  •  [n/Esc] to cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		desc,
		"",
		warningMsg,
		"",
		prompt,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E11D48")).
		Padding(1, 3).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m AppModel) View() string {
	if m.width == 0 {
		return "Starting ATS TUI..."
	}

	// 0. If Delete Confirmation Modal is open
	if m.showDeleteConfirm {
		return m.renderDeleteConfirmModal()
	}

	// 1. If Form Modal is open
	if m.showFormModal {
		return m.formModal.View(m.width, m.height)
	}

	// 2. If Details Modal is open
	if m.showModal {
		modalBox := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(m.width - 6).
			Height(m.height - 4)

		modalHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			Render(m.modalTitle)

		instructions := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777")).
			Render("Press [↑/↓] or [PgUp/PgDn] to scroll  •  [Esc] or [q] to close modal")

		modalView := lipgloss.JoinVertical(
			lipgloss.Left,
			modalHeader,
			"\n",
			m.detailViewport.View(),
			"\n",
			instructions,
		)

		return modalBox.Render(modalView)
	}

	// 3. If Column Filter Modal is open
	if m.isColFilterModal {
		return m.renderColFilterModal()
	}

	// 4. Tab Bar Header
	tabs := []string{
		"[1] Profiles",
		"[2] Applications",
		"[3] Jobs",
		"[4] Organisations",
	}

	tabStyles := make([]string, len(tabs))
	for i, tab := range tabs {
		if i == m.activeTab {
			tabStyles[i] = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2).
				Render(tab)
		} else {
			tabStyles[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Background(lipgloss.Color("#252525")).
				Padding(0, 2).
				Render(tab)
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabStyles...)

	appTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00E5A3")).
		Render("💼 RECRUITMENT ATS")

	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, appTitle, "   ", tabBar)

	// Sub-header controls (Sort info, Universal Column scroll info, Active Filter info)
	var subHeaderParts []string

	// Active Sort Indicator
	var sortColName string
	var sortAsc bool
	switch m.activeTab {
	case TabProfiles:
		cols := []string{"Candidate ID", "Name", "Location", "Qualification"}
		sortColName = cols[m.profileSortCol]
		sortAsc = m.profileSortAsc
	case TabApplications:
		cols := []string{"Job ID", "Candidate Name", "Status", "Exp CTC", "Experience"}
		sortColName = cols[m.appSortCol]
		sortAsc = m.appSortAsc
	case TabJobs:
		cols := []string{"Job ID", "Client", "Job Title", "Req Date", "Positions"}
		sortColName = cols[m.jobSortCol]
		sortAsc = m.jobSortAsc
	case TabOrganisations:
		cols := []string{"Org Name", "Org Code", "Industry"}
		sortColName = cols[m.orgSortCol]
		sortAsc = m.orgSortAsc
	}

	arrow := "▲"
	if !sortAsc {
		arrow = "▼"
	}
	sortPill := lipgloss.NewStyle().Foreground(lipgloss.Color("#9999FF")).
		Render(fmt.Sprintf("Sort: %s %s [s/S]", sortColName, arrow))
	subHeaderParts = append(subHeaderParts, sortPill)

	// Universal Horizontal Column Scrolling Indicator
	allCols := m.getAllColumnsForActiveTab()
	visCols := visibleColumns(allCols, m.tabColOffsets[m.activeTab], m.width-4)
	if len(allCols) > len(visCols) {
		startIdx := m.tabColOffsets[m.activeTab] + 1
		endIdx := m.tabColOffsets[m.activeTab] + len(visCols)
		colPill := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).
			Render(fmt.Sprintf("Cols: %d–%d of %d [use '[' / ']']", startIdx, endIdx, len(allCols)))
		subHeaderParts = append(subHeaderParts, colPill)
	}

	// Active Filters Indicator
	if len(m.colFilters) > 0 {
		var activeParts []string
		for col, valMap := range m.colFilters {
			if len(valMap) == 0 {
				continue
			}
			var list []string
			for v := range valMap {
				list = append(list, v)
			}
			sort.Strings(list)
			activeParts = append(activeParts, fmt.Sprintf("%s=[%s]", col, strings.Join(list, ", ")))
		}
		if len(activeParts) > 0 {
			filterPill := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFAA33")).
				Render(fmt.Sprintf("Filtered: %s (press 'r' to clear)", strings.Join(activeParts, " • ")))
			subHeaderParts = append(subHeaderParts, filterPill)
		}
	} else if m.searchQuery != "" {
		searchPill := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFA3")).
			Render(fmt.Sprintf("Search: '%s' (press 'r' to clear)", m.searchQuery))
		subHeaderParts = append(subHeaderParts, searchPill)
	}

	subHeader := strings.Join(subHeaderParts, "  •  ")

	// 5. Global Search Bar Overlay if active
	var searchBarView string
	if m.isSearching {
		searchBarView = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#00E5A3")).
			Padding(0, 1).
			Render("🔎 " + m.searchInput.View())
	}

	// 6. Main Body Content
	var mainContent string
	switch m.activeTab {
	case TabProfiles:
		mainContent = m.profilesTable.View()
	case TabApplications:
		if m.appViewMode == AppViewKanban {
			mainContent = m.kanbanModel.View()
		} else {
			mainContent = m.submissionsTable.View()
		}
	case TabJobs:
		mainContent = m.jobsTable.View()
	case TabOrganisations:
		mainContent = m.orgsTable.View()
	}

	// 7. Split-Pane Preview
	var splitPaneView string
	if m.showSplitPane {
		splitPaneView = m.renderSplitPane()
	}

	// 8. Footer / Status Line & Help Bar
	var statusLine string
	if m.statusMsg != "" {
		if m.isError {
			statusLine = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F87")).Render("● " + m.statusMsg)
		} else {
			statusLine = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFA3")).Render("● " + m.statusMsg)
		}
	} else {
		statusLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("● Ready")
	}

	helpParts := m.getHelpParts()
	helpLines := wrapHelpItems(helpParts, m.width)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777"))
	var renderedHelpLines []string
	for _, line := range helpLines {
		renderedHelpLines = append(renderedHelpLines, helpStyle.Render(line))
	}
	helpBar := strings.Join(renderedHelpLines, "\n")

	layoutItems := []string{headerBar}
	if subHeader != "" {
		layoutItems = append(layoutItems, subHeader)
	}
	if searchBarView != "" {
		layoutItems = append(layoutItems, searchBarView)
	}
	layoutItems = append(layoutItems, mainContent)
	if splitPaneView != "" {
		layoutItems = append(layoutItems, splitPaneView)
	}
	layoutItems = append(layoutItems, statusLine, helpBar)

	return lipgloss.JoinVertical(lipgloss.Left, layoutItems...)
}

func (m AppModel) getHelpParts() []string {
	helpParts := []string{
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
	}
	if m.activeTab == TabApplications {
		helpParts = append(helpParts, "[v/t] Kanban/Table", "[m/b] Stage")
	} else {
		helpParts = append(helpParts, "[↑/↓] Select")
	}
	helpParts = append(helpParts, "[R] Reload DB", "[q] Quit")
	return helpParts
}

func wrapHelpItems(items []string, maxWidth int) []string {
	if len(items) == 0 {
		return nil
	}
	if maxWidth <= 0 {
		return []string{strings.Join(items, " • ")}
	}

	sep := " • "
	sepWidth := lipgloss.Width(sep)

	var lines []string
	var currentLine []string
	currentWidth := 0

	for _, item := range items {
		itemWidth := lipgloss.Width(item)
		if len(currentLine) == 0 {
			currentLine = append(currentLine, item)
			currentWidth = itemWidth
		} else if currentWidth+sepWidth+itemWidth <= maxWidth {
			currentLine = append(currentLine, item)
			currentWidth += sepWidth + itemWidth
		} else {
			lines = append(lines, strings.Join(currentLine, sep))
			currentLine = []string{item}
			currentWidth = itemWidth
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, sep))
	}

	return lines
}

func (m AppModel) helpBarLineCount() int {
	lines := wrapHelpItems(m.getHelpParts(), m.width)
	if len(lines) == 0 {
		return 1
	}
	return len(lines)
}

func (m *AppModel) recalculateAvailableHeight() {
	helpLines := m.helpBarLineCount()
	extraHelpLines := 0
	if helpLines > 1 {
		extraHelpLines = helpLines - 1
	}

	availableHeight := m.height - 9 - extraHelpLines
	if m.showSplitPane {
		availableHeight = m.height - 18 - extraHelpLines
	}
	if availableHeight < 6 {
		availableHeight = 6
	}

	m.profilesTable.SetHeight(availableHeight)
	m.submissionsTable.SetHeight(availableHeight)
	m.jobsTable.SetHeight(availableHeight)
	m.orgsTable.SetHeight(availableHeight)

	m.kanbanModel.SetSize(m.width, availableHeight)
	m.detailViewport.Width = m.width - 10
	m.detailViewport.Height = m.height - 10
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		if maxLen <= 3 {
			return s[:maxLen]
		}
		return s[:maxLen-3] + "..."
	}
	return s
}
