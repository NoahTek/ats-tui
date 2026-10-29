package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ats-tui/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Messages for Setup Wizard
type SetupCompletedMsg struct {
	Pool   *pgxpool.Pool
	Config db.Config
}

type SetupErrorMsg struct {
	Err error
}

type SetupConnectingMsg struct{}

type SetupModel struct {
	inputs       []textinput.Model
	focusIndex   int
	seedDemo     bool
	isConnecting bool
	statusMsg    string
	isError      bool
	width        int
	height       int
}

// Field Indices
const (
	fieldHost = iota
	fieldPort
	fieldDatabase
	fieldUser
	fieldPassword
	fieldSSLMode
	fieldSeedDemo
	fieldSubmitBtn
	totalFields
)

func NewSetupModel(initialCfg db.Config) SetupModel {
	inputs := make([]textinput.Model, 6)

	// 0. Host
	inputs[fieldHost] = textinput.New()
	inputs[fieldHost].Placeholder = "localhost"
	if initialCfg.Host != "" {
		inputs[fieldHost].SetValue(initialCfg.Host)
	} else {
		inputs[fieldHost].SetValue("localhost")
	}
	inputs[fieldHost].CharLimit = 100
	inputs[fieldHost].Width = 36

	// 1. Port
	inputs[fieldPort] = textinput.New()
	inputs[fieldPort].Placeholder = "5432"
	if initialCfg.Port > 0 {
		inputs[fieldPort].SetValue(strconv.Itoa(initialCfg.Port))
	} else {
		inputs[fieldPort].SetValue("5432")
	}
	inputs[fieldPort].CharLimit = 6
	inputs[fieldPort].Width = 36

	// 2. Database
	inputs[fieldDatabase] = textinput.New()
	inputs[fieldDatabase].Placeholder = "ats"
	if initialCfg.Database != "" {
		inputs[fieldDatabase].SetValue(initialCfg.Database)
	} else {
		inputs[fieldDatabase].SetValue("ats")
	}
	inputs[fieldDatabase].CharLimit = 64
	inputs[fieldDatabase].Width = 36

	// 3. User
	inputs[fieldUser] = textinput.New()
	inputs[fieldUser].Placeholder = "postgres"
	if initialCfg.User != "" {
		inputs[fieldUser].SetValue(initialCfg.User)
	} else {
		inputs[fieldUser].SetValue("postgres")
	}
	inputs[fieldUser].CharLimit = 64
	inputs[fieldUser].Width = 36

	// 4. Password
	inputs[fieldPassword] = textinput.New()
	inputs[fieldPassword].Placeholder = "(leave empty if none)"
	inputs[fieldPassword].SetValue(initialCfg.Password)
	inputs[fieldPassword].EchoMode = textinput.EchoPassword
	inputs[fieldPassword].EchoCharacter = '•'
	inputs[fieldPassword].CharLimit = 128
	inputs[fieldPassword].Width = 36

	// 5. SSL Mode
	inputs[fieldSSLMode] = textinput.New()
	inputs[fieldSSLMode].Placeholder = "disable"
	if initialCfg.SSLMode != "" {
		inputs[fieldSSLMode].SetValue(initialCfg.SSLMode)
	} else {
		inputs[fieldSSLMode].SetValue("disable")
	}
	inputs[fieldSSLMode].CharLimit = 32
	inputs[fieldSSLMode].Width = 36

	// Focus first input
	inputs[fieldHost].Focus()

	return SetupModel{
		inputs:     inputs,
		focusIndex: fieldHost,
		seedDemo:   initialCfg.SeedDemo,
		statusMsg:  "Press [Enter] or [Tab] to navigate fields. Press [Ctrl+S] to connect.",
	}
}

func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SetupModel) Update(msg tea.Msg) (SetupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case SetupConnectingMsg:
		m.isConnecting = true
		m.isError = false
		m.statusMsg = "Connecting to PostgreSQL and running migrations..."
		return m, m.performSetupCmd()

	case SetupErrorMsg:
		m.isConnecting = false
		m.isError = true
		m.statusMsg = fmt.Sprintf("Error: %v", msg.Err)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+s":
			if !m.isConnecting {
				return m, func() tea.Msg { return SetupConnectingMsg{} }
			}
			return m, nil

		case "tab", "down":
			m.focusNext()
			return m, nil

		case "shift+tab", "up":
			m.focusPrev()
			return m, nil

		case "enter":
			if m.focusIndex == fieldSubmitBtn {
				if !m.isConnecting {
					return m, func() tea.Msg { return SetupConnectingMsg{} }
				}
				return m, nil
			}
			if m.focusIndex == fieldSeedDemo {
				m.seedDemo = !m.seedDemo
				return m, nil
			}
			m.focusNext()
			return m, nil

		case " ":
			if m.focusIndex == fieldSeedDemo {
				m.seedDemo = !m.seedDemo
				return m, nil
			}
			if m.focusIndex == fieldSubmitBtn {
				if !m.isConnecting {
					return m, func() tea.Msg { return SetupConnectingMsg{} }
				}
				return m, nil
			}
		}
	}

	// Forward key to active input if focused on one of the inputs
	if m.focusIndex < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *SetupModel) focusNext() {
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Blur()
	}
	m.focusIndex = (m.focusIndex + 1) % totalFields
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Focus()
	}
}

func (m *SetupModel) focusPrev() {
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Blur()
	}
	m.focusIndex = (m.focusIndex + totalFields - 1) % totalFields
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Focus()
	}
}

func (m SetupModel) toConfig() (db.Config, error) {
	portStr := strings.TrimSpace(m.inputs[fieldPort].Value())
	port := 5432
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return db.Config{}, fmt.Errorf("invalid port number: %q", portStr)
		}
		port = p
	}

	host := strings.TrimSpace(m.inputs[fieldHost].Value())
	if host == "" {
		host = "localhost"
	}

	database := strings.TrimSpace(m.inputs[fieldDatabase].Value())
	if database == "" {
		database = "ats"
	}

	user := strings.TrimSpace(m.inputs[fieldUser].Value())
	if user == "" {
		user = "postgres"
	}

	sslMode := strings.TrimSpace(m.inputs[fieldSSLMode].Value())
	if sslMode == "" {
		sslMode = "disable"
	}

	return db.Config{
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Password: m.inputs[fieldPassword].Value(),
		SSLMode:  sslMode,
		SeedDemo: m.seedDemo,
	}, nil
}

func (m SetupModel) performSetupCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := m.toConfig()
		if err != nil {
			return SetupErrorMsg{Err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		connURI := cfg.ConnectionURI()

		// 1. Ensure target database exists (creates via /postgres if absent)
		if err := db.EnsureDatabase(ctx, connURI); err != nil {
			// Continue in case user permissions only allow direct connection
			_ = err
		}

		// 2. Create connection pool
		pool, err := db.NewPool(ctx, connURI)
		if err != nil {
			return SetupErrorMsg{Err: fmt.Errorf("connection failed: %w", err)}
		}

		// 3. Automatically apply embedded schema migrations
		if err := db.RunMigrations(ctx, pool); err != nil {
			pool.Close()
			return SetupErrorMsg{Err: fmt.Errorf("migration failed: %w", err)}
		}

		// 4. Optionally seed demo data if database is empty or user requested
		if cfg.SeedDemo {
			isEmpty, err := db.IsDatabaseEmpty(ctx, pool)
			if err == nil && isEmpty {
				if err := db.SeedSampleData(ctx, pool); err != nil {
					// Non-fatal, but logged
					_ = err
				}
			}
		}

		// 5. Persist configuration to disk
		if err := db.SaveConfig(cfg); err != nil {
			// Non-fatal, continuing
			_ = err
		}

		return SetupCompletedMsg{
			Pool:   pool,
			Config: cfg,
		}
	}
}

// View rendering
func (m SetupModel) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(64)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#5A56E0")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Width(16).
		Bold(true).
		Foreground(lipgloss.Color("#E0E0E0"))

	focusedLabelStyle := lipgloss.NewStyle().
		Width(16).
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	var rows []string

	// Title
	rows = append(rows, titleStyle.Render("🎯  ATS-TUI Database Setup Wizard"))
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(
		"Welcome! Configure your PostgreSQL connection below.\nSettings will be saved to your local config automatically.\n",
	))

	// Fields
	fieldLabels := []string{
		"Host",
		"Port",
		"Database Name",
		"Username",
		"Password",
		"SSL Mode",
	}

	for i := 0; i < len(m.inputs); i++ {
		lbl := labelStyle
		if m.focusIndex == i {
			lbl = focusedLabelStyle
		}
		rows = append(rows, fmt.Sprintf("%s %s", lbl.Render(fieldLabels[i]+":"), m.inputs[i].View()))
	}

	// Checkbox field for Seed Demo
	checkboxChar := "[ ]"
	if m.seedDemo {
		checkboxChar = "[✓]"
	}
	seedLabel := labelStyle
	if m.focusIndex == fieldSeedDemo {
		seedLabel = focusedLabelStyle
	}
	seedCheckStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	rows = append(rows, fmt.Sprintf("%s %s Load demo candidates & pipeline data", seedLabel.Render("Seed Demo:"), seedCheckStyle.Render(checkboxChar)))

	rows = append(rows, "")

	// Submit Button
	btnStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 2).
		MarginRight(1)

	if m.focusIndex == fieldSubmitBtn {
		btnStyle = btnStyle.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4"))
	} else {
		btnStyle = btnStyle.
			Foreground(lipgloss.Color("#E0E0E0")).
			Background(lipgloss.Color("#383838"))
	}

	btnText := "[ Connect & Launch ATS ]"
	if m.isConnecting {
		btnText = "⏳ [ Connecting... ]"
	}
	rows = append(rows, btnStyle.Render(btnText))

	// Status Line
	rows = append(rows, "")
	statusStyle := lipgloss.NewStyle().Italic(true)
	if m.isError {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#FF5555")).Bold(true)
	} else if m.isConnecting {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#F1FA8C"))
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#8BE9FD"))
	}
	rows = append(rows, statusStyle.Render(m.statusMsg))

	// Controls Helper
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(
		"\n[Tab/↓] Next  •  [Shift+Tab/↑] Prev  •  [Space] Toggle  •  [Enter] Submit  •  [Esc] Quit",
	))

	content := strings.Join(rows, "\n")
	boxed := boxStyle.Render(content)

	// Center horizontally and vertically
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
	}
	return boxed
}
