package kanban

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ats-tui/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Messages
type DataLoadedMsg struct {
	StageMap map[models.Stage][]models.ApplicationCard
}

type StageUpdatedMsg struct {
	AppID         int
	OldStage      models.Stage
	NewStage      models.Stage
	CandidateName string
}

type StatusClearMsg struct{}

type ErrMsg struct {
	Err error
}

type Column struct {
	Stage  models.Stage
	Title  string
	Cards  []models.ApplicationCard
	Cursor int
}

func (c *Column) SelectedCard() *models.ApplicationCard {
	if len(c.Cards) == 0 || c.Cursor < 0 || c.Cursor >= len(c.Cards) {
		return nil
	}
	return &c.Cards[c.Cursor]
}

func (c *Column) MoveCursor(delta int) {
	if len(c.Cards) == 0 {
		c.Cursor = 0
		return
	}
	c.Cursor += delta
	if c.Cursor < 0 {
		c.Cursor = 0
	}
	if c.Cursor >= len(c.Cards) {
		c.Cursor = len(c.Cards) - 1
	}
}

type Model struct {
	pool          *pgxpool.Pool
	columns       []Column
	activeCol     int
	width         int
	height        int
	statusMessage string
	isError       bool
	loading       bool
	JobCodeFilter string
}

func NewModel(pool *pgxpool.Pool) Model {
	cols := make([]Column, len(models.PipelineStages))
	for i, stage := range models.PipelineStages {
		cols[i] = Column{
			Stage:  stage,
			Title:  stage.Title(),
			Cards:  []models.ApplicationCard{},
			Cursor: 0,
		}
	}

	return Model{
		pool:          pool,
		columns:       cols,
		activeCol:     0,
		loading:       true,
		JobCodeFilter: "ALL",
	}
}

func (m Model) Init() tea.Cmd {
	return m.FetchApplicationsCmd()
}

func (m Model) SelectedCard() *models.ApplicationCard {
	if m.activeCol < 0 || m.activeCol >= len(m.columns) {
		return nil
	}
	return m.columns[m.activeCol].SelectedCard()
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) SetJobFilter(filter string) tea.Cmd {
	m.JobCodeFilter = filter
	m.loading = true
	return m.FetchApplicationsCmd()
}

func (m Model) FetchApplicationsCmd() tea.Cmd {
	filter := m.JobCodeFilter
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		stageMap, err := models.FetchBoardApplications(ctx, m.pool, filter)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return DataLoadedMsg{StageMap: stageMap}
	}
}

func updateStageCmd(pool *pgxpool.Pool, appID int, oldStage, newStage models.Stage, candidateName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := models.UpdateApplicationStage(ctx, pool, appID, newStage); err != nil {
			return ErrMsg{Err: err}
		}
		return StageUpdatedMsg{
			AppID:         appID,
			OldStage:      oldStage,
			NewStage:      newStage,
			CandidateName: candidateName,
		}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return StatusClearMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case DataLoadedMsg:
		m.loading = false
		for i := range m.columns {
			stage := m.columns[i].Stage
			m.columns[i].Cards = msg.StageMap[stage]
			if m.columns[i].Cursor >= len(m.columns[i].Cards) {
				if len(m.columns[i].Cards) > 0 {
					m.columns[i].Cursor = len(m.columns[i].Cards) - 1
				} else {
					m.columns[i].Cursor = 0
				}
			}
		}
		m.statusMessage = fmt.Sprintf("Synced with PostgreSQL (%s)", m.JobCodeFilter)
		m.isError = false
		return m, clearStatusAfter(2 * time.Second)

	case StageUpdatedMsg:
		var movedCard *models.ApplicationCard
		for i := range m.columns {
			if m.columns[i].Stage == msg.OldStage {
				for idx, card := range m.columns[i].Cards {
					if card.ApplicationID == msg.AppID {
						c := card
						c.Status = msg.NewStage
						movedCard = &c
						m.columns[i].Cards = append(m.columns[i].Cards[:idx], m.columns[i].Cards[idx+1:]...)
						if m.columns[i].Cursor >= len(m.columns[i].Cards) && len(m.columns[i].Cards) > 0 {
							m.columns[i].Cursor = len(m.columns[i].Cards) - 1
						}
						break
					}
				}
				break
			}
		}
		if movedCard != nil {
			for i := range m.columns {
				if m.columns[i].Stage == msg.NewStage {
					m.columns[i].Cards = append(m.columns[i].Cards, *movedCard)
					m.columns[i].Cursor = len(m.columns[i].Cards) - 1
					m.activeCol = i
					break
				}
			}
		}
		m.statusMessage = fmt.Sprintf("Moved %s -> %s", msg.CandidateName, msg.NewStage.Title())
		m.isError = false
		return m, clearStatusAfter(3 * time.Second)

	case StatusClearMsg:
		m.statusMessage = ""
		m.isError = false
		return m, nil

	case ErrMsg:
		m.statusMessage = fmt.Sprintf("Error: %v", msg.Err)
		m.isError = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if m.activeCol > 0 {
				m.activeCol--
			} else {
				m.activeCol = len(m.columns) - 1
			}
			return m, nil

		case "l", "right":
			if m.activeCol < len(m.columns)-1 {
				m.activeCol++
			} else {
				m.activeCol = 0
			}
			return m, nil

		case "j", "down":
			if len(m.columns) > 0 {
				m.columns[m.activeCol].MoveCursor(1)
			}
			return m, nil

		case "k", "up":
			if len(m.columns) > 0 {
				m.columns[m.activeCol].MoveCursor(-1)
			}
			return m, nil

		case "m":
			card := m.columns[m.activeCol].SelectedCard()
			if card == nil {
				return m, nil
			}
			nextStage, ok := card.Status.Next()
			if !ok {
				m.statusMessage = fmt.Sprintf("%s is in final stage (%s)", card.CandidateName, card.Status.Title())
				m.isError = false
				return m, clearStatusAfter(2 * time.Second)
			}
			return m, updateStageCmd(m.pool, card.ApplicationID, card.Status, nextStage, card.CandidateName)

		case "b":
			card := m.columns[m.activeCol].SelectedCard()
			if card == nil {
				return m, nil
			}
			prevStage, ok := card.Status.Prev()
			if !ok {
				m.statusMessage = fmt.Sprintf("%s is in initial stage (%s)", card.CandidateName, card.Status.Title())
				m.isError = false
				return m, clearStatusAfter(2 * time.Second)
			}
			return m, updateStageCmd(m.pool, card.ApplicationID, card.Status, prevStage, card.CandidateName)

		case "r":
			m.loading = true
			m.statusMessage = "Refreshing board..."
			return m, m.FetchApplicationsCmd()
		}
	}

	return m, nil
}

// Lipgloss styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			MarginBottom(1)

	headerInactiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#888888")).
				Background(lipgloss.Color("#252525")).
				Padding(0, 1).
				MarginBottom(1)

	columnActiveBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	columnInactiveBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#383838")).
				Padding(0, 1)

	cardSelectedActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00D7D7")).
				Padding(0, 1).
				MarginBottom(1)

	cardSelectedInactive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#383838")).
				Padding(0, 1).
				MarginBottom(1)

	cardNormal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#383838")).
			Padding(0, 1).
			MarginBottom(1)

	cardName = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	cardCode = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFCC00")).
			Bold(true)

	cardSub = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0"))

	skillBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00E5A3")).
			Bold(true)

	prodBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA33")).
			Italic(true)
)

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing ATS Kanban..."
	}

	numCols := len(m.columns)
	colWidth := (m.width - (numCols * 4) - 2) / numCols
	if colWidth < 26 {
		colWidth = 26
	}

	renderedCols := make([]string, numCols)
	targetHeight := m.height - 4
	if targetHeight < 8 {
		targetHeight = 8
	}

	for i, col := range m.columns {
		isActiveCol := (i == m.activeCol)

		// Column Header
		headerText := fmt.Sprintf("%s (%d)", col.Title, len(col.Cards))
		var header string
		if isActiveCol {
			header = headerStyle.Width(colWidth - 2).Align(lipgloss.Center).Render(headerText)
		} else {
			header = headerInactiveStyle.Width(colWidth - 2).Align(lipgloss.Center).Render(headerText)
		}

		// Cards inside column
		var cardsStr strings.Builder
		if len(col.Cards) == 0 {
			cardsStr.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555")).
				Italic(true).
				Render("  No candidates"))
		} else {
			for cIdx, card := range col.Cards {
				isSelectedCard := (cIdx == col.Cursor)

				var cardStyle lipgloss.Style
				if isSelectedCard && isActiveCol {
					cardStyle = cardSelectedActive.Width(colWidth - 4)
				} else {
					cardStyle = cardNormal.Width(colWidth - 4)
				}

				topLine := fmt.Sprintf("%s  %s", cardName.Render(card.CandidateName), cardCode.Render("["+card.CandidateCode+"]"))
				roleLine := cardSub.Render(fmt.Sprintf("%s (%dyr) @ %s", card.Designation, card.ExperienceYears, card.CurrentOrg))
				metaLine := fmt.Sprintf("%s  %s  %s",
					skillBadge.Render(card.Skill),
					cardSub.Render(fmt.Sprintf("$%dk", card.ExpectedCTC/1000)),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#9999FF")).Render(card.JobCode),
				)

				var prodLine string
				if card.Products != "" {
					prodLine = prodBadge.Render("📦 " + card.Products)
				}

				items := []string{topLine, roleLine, metaLine}
				if prodLine != "" {
					items = append(items, prodLine)
				}

				cardContent := lipgloss.JoinVertical(lipgloss.Left, items...)
				cardsStr.WriteString(cardStyle.Render(cardContent))
				cardsStr.WriteString("\n")
			}
		}

		columnContent := lipgloss.JoinVertical(lipgloss.Left, header, cardsStr.String())

		var borderedCol string
		if isActiveCol {
			borderedCol = columnActiveBorder.Width(colWidth).Height(targetHeight).Render(columnContent)
		} else {
			borderedCol = columnInactiveBorder.Width(colWidth).Height(targetHeight).Render(columnContent)
		}

		renderedCols[i] = borderedCol
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}
