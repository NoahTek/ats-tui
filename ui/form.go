package ui

import (
	"fmt"
	"strconv"
	"strings"

	"ats-tui/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FormMode int

const (
	FormModeAdd FormMode = iota
	FormModeEdit
)

type EntityType int

const (
	EntityCandidate EntityType = iota
	EntityApplication
	EntityJob
	EntityOrganisation
)

// Messages
type FormCancelMsg struct{}

type FormSubmitCandidateMsg struct {
	Profile models.CandidateProfile
	IsEdit  bool
}

type FormSubmitApplicationMsg struct {
	Submission     models.ApplicationSubmissionRow
	IsEdit         bool
	InitialComment string
}

type FormSubmitJobMsg struct {
	Job    models.JobPosting
	IsEdit bool
}

type FormSubmitOrgMsg struct {
	Org    models.Organisation
	IsEdit bool
}

type FormField struct {
	Key      string
	Label    string
	Input    textinput.Model
	Required bool
}

type FormModel struct {
	mode        FormMode
	entityType  EntityType
	targetID    int
	targetCode  string
	title       string
	fields      []FormField
	activeField int
	errMsg      string
	width       int
	height      int
}

func createField(key, label, placeholder, initialVal string, required bool) FormField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(initialVal)
	ti.CharLimit = 250
	ti.Width = 42

	return FormField{
		Key:      key,
		Label:    label,
		Input:    ti,
		Required: required,
	}
}

// NewCandidateForm creates a form for Adding or Editing candidate profiles.
func NewCandidateForm(c *models.CandidateProfile) FormModel {
	mode := FormModeAdd
	title := "➕ ADD NEW CANDIDATE PROFILE"
	targetID := 0
	targetCode := ""

	var (
		name, email, phone, location, qualification, dob, resume string
	)

	if c != nil {
		mode = FormModeEdit
		targetID = c.ID
		targetCode = c.CandidateCode
		title = fmt.Sprintf("✏️ EDIT CANDIDATE PROFILE [%s]", c.CandidateCode)
		name = c.Name
		email = c.Email
		phone = c.Phone
		location = c.Location
		qualification = c.Qualification
		dob = c.DOB
		resume = c.ResumeText
	}

	fields := []FormField{
		createField("name", "Full Name *", "e.g. Jane Doe", name, true),
		createField("email", "Email Address *", "e.g. jane.doe@example.com", email, true),
		createField("phone", "Contact Number", "e.g. +1-555-0199", phone, false),
		createField("location", "Location", "e.g. San Francisco, CA or Remote", location, false),
		createField("qualification", "Educational Qualification", "e.g. B.S. in Computer Science", qualification, false),
		createField("dob", "Date of Birth (YYYY-MM-DD)", "e.g. 1995-04-12", dob, false),
		createField("resume", "Resume Text / Skills", "e.g. Go, PostgreSQL, Distributed Systems...", resume, false),
	}

	fields[0].Input.Focus()

	return FormModel{
		mode:        mode,
		entityType:  EntityCandidate,
		targetID:    targetID,
		targetCode:  targetCode,
		title:       title,
		fields:      fields,
		activeField: 0,
	}
}

// NewApplicationForm creates a form for Adding or Editing application submissions.
func NewApplicationForm(a *models.ApplicationSubmissionRow, defaultJob, defaultCand string) FormModel {
	mode := FormModeAdd
	title := "➕ ADD APPLICATION SUBMISSION"
	targetID := 0
	targetCode := ""

	var (
		jobCode     = defaultJob
		candCode    = defaultCand
		org         string
		designation string
		products    string
		exp         = "3"
		ctc         = "120000"
		expCtc      = "140000"
		notice      = "30"
		status      = "interested"
		comment     string
	)

	if a != nil {
		mode = FormModeEdit
		targetID = a.ApplicationID
		title = fmt.Sprintf("✏️ EDIT APPLICATION [%s -> %s]", a.CandidateCode, a.JobCode)
		jobCode = a.JobCode
		candCode = a.CandidateCode
		org = a.CurrentOrg
		designation = a.Designation
		products = a.Products
		exp = strconv.Itoa(a.ExperienceYears)
		ctc = strconv.Itoa(a.CTC)
		expCtc = strconv.Itoa(a.ExpectedCTC)
		notice = strconv.Itoa(a.NoticePeriodDays)
		status = string(a.Status)
		comment = a.Comments
	}

	fields := []FormField{
		createField("job_code", "Job ID *", "e.g. JOB1, JOB2", jobCode, true),
		createField("cand_code", "Candidate Code *", "e.g. CAN-001", candCode, true),
		createField("org", "Current Organisation", "e.g. Stripe, TechCorp", org, false),
		createField("designation", "Designation / Role", "e.g. Backend Engineer", designation, false),
		createField("products", "Product Line(s)", "e.g. Payments, Cloud SRE", products, false),
		createField("exp", "Experience (Years)", "e.g. 5", exp, false),
		createField("ctc", "Current CTC ($)", "e.g. 130000", ctc, false),
		createField("exp_ctc", "Expected CTC ($)", "e.g. 150000", expCtc, false),
		createField("notice", "Notice Period (Days)", "e.g. 30", notice, false),
		createField("status", "Pipeline Status", "interested, follow_up, call_back, closed", status, false),
		createField("comment", "Initial Comment / Note", "e.g. Screened by recruiter...", comment, false),
	}

	fields[0].Input.Focus()

	return FormModel{
		mode:        mode,
		entityType:  EntityApplication,
		targetID:    targetID,
		targetCode:  targetCode,
		title:       title,
		fields:      fields,
		activeField: 0,
	}
}

// NewJobForm creates a form for Adding or Editing open job positions.
func NewJobForm(j *models.JobPosting) FormModel {
	mode := FormModeAdd
	title := "➕ ADD JOB REQUISITION"
	targetID := 0

	var (
		code, client, titleStr, role, dept, subDept, loc string
		positions = "1"
	)

	if j != nil {
		mode = FormModeEdit
		targetID = j.ID
		title = fmt.Sprintf("✏️ EDIT JOB REQUISITION [%s]", j.JobCode)
		code = j.JobCode
		client = j.ClientName
		titleStr = j.JobTitle
		role = j.Role
		dept = j.Department
		subDept = j.SubDepartment
		loc = j.Location
		positions = strconv.Itoa(j.NumPositions)
	}

	fields := []FormField{
		createField("job_code", "Job ID *", "e.g. JOB4", code, true),
		createField("client", "Client Org Code", "e.g. YSIL, ABCL", client, false),
		createField("title", "Job Title *", "e.g. Senior Go Architect", titleStr, true),
		createField("role", "Role", "e.g. Staff Engineer", role, false),
		createField("dept", "Department", "e.g. Engineering", dept, false),
		createField("sub_dept", "Sub-Department", "e.g. Platform Infrastructure", subDept, false),
		createField("location", "Location", "e.g. Remote, Austin, TX", loc, false),
		createField("positions", "Positions Open", "e.g. 2", positions, false),
	}

	fields[0].Input.Focus()

	return FormModel{
		mode:        mode,
		entityType:  EntityJob,
		targetID:    targetID,
		title:       title,
		fields:      fields,
		activeField: 0,
	}
}

// NewOrganisationForm creates a form for Adding or Editing client organisations.
func NewOrganisationForm(o *models.Organisation) FormModel {
	mode := FormModeAdd
	title := "➕ ADD CLIENT ORGANISATION"
	targetID := 0

	var (
		name, code, ind, subCat string
	)

	if o != nil {
		mode = FormModeEdit
		targetID = o.ID
		title = fmt.Sprintf("✏️ EDIT CLIENT [%s]", o.OrgCode)
		name = o.OrgName
		code = o.OrgCode
		ind = o.Industry
		subCat = o.SubCategory
	}

	fields := []FormField{
		createField("name", "Organisation Name *", "e.g. Datadog Labs", name, true),
		createField("code", "Organisation Code *", "e.g. DTDG", code, true),
		createField("industry", "Industry", "e.g. Information Technology", ind, false),
		createField("sub_cat", "Sub Category", "e.g. Cloud Infrastructure", subCat, false),
	}

	fields[0].Input.Focus()

	return FormModel{
		mode:        mode,
		entityType:  EntityOrganisation,
		targetID:    targetID,
		title:       title,
		fields:      fields,
		activeField: 0,
	}
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return FormCancelMsg{} }

		case "tab", "down":
			m.fields[m.activeField].Input.Blur()
			m.activeField = (m.activeField + 1) % len(m.fields)
			m.fields[m.activeField].Input.Focus()
			return m, textinput.Blink

		case "shift+tab", "up":
			m.fields[m.activeField].Input.Blur()
			m.activeField = (m.activeField - 1 + len(m.fields)) % len(m.fields)
			m.fields[m.activeField].Input.Focus()
			return m, textinput.Blink

		case "enter":
			// If on last field, or submitting with Enter
			if m.activeField == len(m.fields)-1 {
				return m, m.submit()
			}
			// Otherwise move to next field
			m.fields[m.activeField].Input.Blur()
			m.activeField++
			m.fields[m.activeField].Input.Focus()
			return m, textinput.Blink

		case "ctrl+s":
			return m, m.submit()
		}
	}

	var cmd tea.Cmd
	m.fields[m.activeField].Input, cmd = m.fields[m.activeField].Input.Update(msg)
	return m, cmd
}

func (m *FormModel) submit() tea.Cmd {
	m.errMsg = ""
	// Validate required fields
	for _, f := range m.fields {
		if f.Required && strings.TrimSpace(f.Input.Value()) == "" {
			m.errMsg = fmt.Sprintf("Field '%s' is required!", f.Label)
			return nil
		}
	}

	vals := make(map[string]string)
	for _, f := range m.fields {
		vals[f.Key] = strings.TrimSpace(f.Input.Value())
	}

	isEdit := (m.mode == FormModeEdit)

	switch m.entityType {
	case EntityCandidate:
		p := models.CandidateProfile{
			ID:            m.targetID,
			CandidateCode: m.targetCode,
			Name:          vals["name"],
			Email:         vals["email"],
			Phone:         vals["phone"],
			Location:      vals["location"],
			Qualification: vals["qualification"],
			DOB:           vals["dob"],
			ResumeText:    vals["resume"],
		}
		return func() tea.Msg {
			return FormSubmitCandidateMsg{Profile: p, IsEdit: isEdit}
		}

	case EntityApplication:
		exp, _ := strconv.Atoi(vals["exp"])
		ctc, _ := strconv.Atoi(vals["ctc"])
		expCtc, _ := strconv.Atoi(vals["exp_ctc"])
		notice, _ := strconv.Atoi(vals["notice"])
		status := models.Stage(vals["status"])
		if status == "" {
			status = models.StageInterested
		}

		sub := models.ApplicationSubmissionRow{
			ApplicationID:    m.targetID,
			JobCode:          vals["job_code"],
			CandidateCode:    vals["cand_code"],
			CurrentOrg:       vals["org"],
			Designation:      vals["designation"],
			Products:         vals["products"],
			ExperienceYears:  exp,
			CTC:              ctc,
			ExpectedCTC:      expCtc,
			NoticePeriodDays: notice,
			Status:           status,
			Comments:         vals["comment"],
		}
		return func() tea.Msg {
			return FormSubmitApplicationMsg{Submission: sub, IsEdit: isEdit, InitialComment: vals["comment"]}
		}

	case EntityJob:
		pos, _ := strconv.Atoi(vals["positions"])
		if pos <= 0 {
			pos = 1
		}
		j := models.JobPosting{
			ID:            m.targetID,
			JobCode:       vals["job_code"],
			ClientName:    vals["client"],
			JobTitle:      vals["title"],
			Role:          vals["role"],
			Department:    vals["dept"],
			SubDepartment: vals["sub_dept"],
			Location:      vals["location"],
			NumPositions:  pos,
		}
		return func() tea.Msg {
			return FormSubmitJobMsg{Job: j, IsEdit: isEdit}
		}

	case EntityOrganisation:
		o := models.Organisation{
			ID:          m.targetID,
			OrgName:     vals["name"],
			OrgCode:     vals["code"],
			Industry:    vals["industry"],
			SubCategory: vals["sub_cat"],
		}
		return func() tea.Msg {
			return FormSubmitOrgMsg{Org: o, IsEdit: isEdit}
		}
	}

	return nil
}

func (m FormModel) View(termWidth, termHeight int) string {
	boxWidth := 68
	if termWidth > 0 && boxWidth > termWidth-4 {
		boxWidth = termWidth - 4
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 2).
		Align(lipgloss.Center)

	formBorder := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#00E5A3")).
		Padding(1, 2).
		Width(boxWidth)

	var lines []string
	lines = append(lines, titleStyle.Width(boxWidth-6).Render(m.title))
	lines = append(lines, "")

	if m.errMsg != "" {
		errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F87"))
		lines = append(lines, errStyle.Render("⚠️  "+m.errMsg), "")
	}

	for i, f := range m.fields {
		isActive := (i == m.activeField)
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0"))
		if isActive {
			labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFA3"))
		}

		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(0, 1)

		if isActive {
			inputStyle = inputStyle.BorderForeground(lipgloss.Color("#00D7D7"))
		}

		lines = append(lines, labelStyle.Render(f.Label))
		lines = append(lines, inputStyle.Width(boxWidth-8).Render(f.Input.View()))
		lines = append(lines, "")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Italic(true)
	lines = append(lines, helpStyle.Render("[Tab/↓] Next  •  [Shift+Tab/↑] Prev  •  [Ctrl+S / Enter] Submit  •  [Esc] Cancel"))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return formBorder.Render(content)
}
