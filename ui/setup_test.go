package ui

import (
	"ats-tui/db"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSetupModelFieldNavigation(t *testing.T) {
	cfg := db.DefaultConfig()
	model := NewSetupModel(cfg)

	if model.focusIndex != fieldHost {
		t.Errorf("Expected initial focus on fieldHost (0), got %d", model.focusIndex)
	}

	// Tab moves to fieldPort
	var cmd tea.Cmd
	model, cmd = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = cmd
	if model.focusIndex != fieldPort {
		t.Errorf("Expected focus on fieldPort (1), got %d", model.focusIndex)
	}

	// Shift+Tab moves back to fieldHost
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.focusIndex != fieldHost {
		t.Errorf("Expected focus back on fieldHost (0), got %d", model.focusIndex)
	}
}

func TestSetupModelSeedDemoToggle(t *testing.T) {
	cfg := db.DefaultConfig()
	cfg.SeedDemo = true
	model := NewSetupModel(cfg)

	// Move to seed demo checkbox
	model.focusIndex = fieldSeedDemo
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})

	if model.seedDemo != false {
		t.Errorf("Expected seedDemo to toggle to false, got %v", model.seedDemo)
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.seedDemo != true {
		t.Errorf("Expected seedDemo to toggle back to true with Enter, got %v", model.seedDemo)
	}
}

func TestSetupModelConfigParsing(t *testing.T) {
	cfg := db.DefaultConfig()
	model := NewSetupModel(cfg)

	parsed, err := model.toConfig()
	if err != nil {
		t.Fatalf("Unexpected error parsing config: %v", err)
	}

	if parsed.Host != "localhost" || parsed.Port != 5432 || parsed.Database != "ats" {
		t.Errorf("Unexpected parsed config: %+v", parsed)
	}

	// Test invalid port error handling
	model.inputs[fieldPort].SetValue("999999")
	_, err = model.toConfig()
	if err == nil {
		t.Error("Expected error for invalid port 999999, got nil")
	}
}

func TestRootModelTransitions(t *testing.T) {
	root := NewRootModel(nil, nil, true)
	if root.mode != RootModeSetup {
		t.Errorf("Expected RootModeSetup, got %v", root.mode)
	}

	view := root.View()
	if view == "" {
		t.Error("Expected non-empty setup view")
	}
}
