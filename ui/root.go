package ui

import (
	"ats-tui/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RootMode int

const (
	RootModeSetup RootMode = iota
	RootModeApp
)

type RootModel struct {
	mode       RootMode
	setupModel SetupModel
	appModel   AppModel
	width      int
	height     int
}

func NewRootModel(pool *pgxpool.Pool, initialConfig *db.Config, startInSetup bool) RootModel {
	if startInSetup || pool == nil {
		cfg := db.DefaultConfig()
		if initialConfig != nil {
			cfg = *initialConfig
		}
		return RootModel{
			mode:       RootModeSetup,
			setupModel: NewSetupModel(cfg),
		}
	}

	return RootModel{
		mode:     RootModeApp,
		appModel: NewApp(pool),
	}
}

func (m RootModel) Init() tea.Cmd {
	if m.mode == RootModeApp {
		return m.appModel.Init()
	}
	return m.setupModel.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.mode == RootModeApp {
			var cmd tea.Cmd
			var app tea.Model
			app, cmd = m.appModel.Update(msg)
			m.appModel = app.(AppModel)
			return m, cmd
		} else {
			var cmd tea.Cmd
			m.setupModel, cmd = m.setupModel.Update(msg)
			return m, cmd
		}

	case SetupCompletedMsg:
		m.mode = RootModeApp
		m.appModel = NewApp(msg.Pool)
		if m.width > 0 && m.height > 0 {
			var app tea.Model
			app, _ = m.appModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.appModel = app.(AppModel)
		}
		return m, m.appModel.Init()
	}

	if m.mode == RootModeApp {
		var cmd tea.Cmd
		var app tea.Model
		app, cmd = m.appModel.Update(msg)
		m.appModel = app.(AppModel)
		return m, cmd
	}

	var cmd tea.Cmd
	m.setupModel, cmd = m.setupModel.Update(msg)
	return m, cmd
}

func (m RootModel) View() string {
	if m.mode == RootModeApp {
		return m.appModel.View()
	}
	return m.setupModel.View()
}
