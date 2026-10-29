package kanban

import (
	"testing"

	"ats-tui/models"
	"strings"
)

func TestStageTransitions(t *testing.T) {
	// interested -> follow_up -> call_back -> closed
	s := models.StageInterested
	next, ok := s.Next()
	if !ok || next != models.StageFollowUp {
		t.Fatalf("expected next of %s to be %s, got %s", s, models.StageFollowUp, next)
	}

	next, ok = next.Next()
	if !ok || next != models.StageCallBack {
		t.Fatalf("expected next to be %s, got %s", models.StageCallBack, next)
	}

	next, ok = next.Next()
	if !ok || next != models.StageClosed {
		t.Fatalf("expected next to be %s, got %s", models.StageClosed, next)
	}

	_, ok = next.Next()
	if ok {
		t.Fatalf("closed stage should not have next stage")
	}

	prev, ok := next.Prev()
	if !ok || prev != models.StageCallBack {
		t.Fatalf("expected prev of %s to be %s, got %s", next, models.StageCallBack, prev)
	}
}

func TestColumnCursorNavigation(t *testing.T) {
	col := Column{
		Stage: models.StageInterested,
		Cards: []models.ApplicationCard{
			{ApplicationID: 1, CandidateName: "Alice"},
			{ApplicationID: 2, CandidateName: "Bob"},
		},
		Cursor: 0,
	}

	if col.SelectedCard().CandidateName != "Alice" {
		t.Fatalf("expected Alice, got %v", col.SelectedCard())
	}

	col.MoveCursor(1)
	if col.SelectedCard().CandidateName != "Bob" {
		t.Fatalf("expected Bob, got %v", col.SelectedCard())
	}

	// Should clamp at bottom
	col.MoveCursor(5)
	if col.Cursor != 1 {
		t.Fatalf("expected cursor to clamp at 1, got %d", col.Cursor)
	}

	// Move back up
	col.MoveCursor(-1)
	if col.Cursor != 0 {
		t.Fatalf("expected cursor to be 0, got %d", col.Cursor)
	}

	// Should clamp at top
	col.MoveCursor(-5)
	if col.Cursor != 0 {
		t.Fatalf("expected cursor to clamp at 0, got %d", col.Cursor)
	}
}

func TestCardRenderingAndSelectionHighlight(t *testing.T) {
	m := NewModel(nil)
	m.SetSize(120, 30)

	m.columns[0].Cards = []models.ApplicationCard{
		{ApplicationID: 1, CandidateName: "Alex Rivera", CandidateCode: "CAN-001", Designation: "Engineer", Skill: "Go", ExpectedCTC: 150000, JobCode: "JOB1"},
		{ApplicationID: 2, CandidateName: "Marcus Vance", CandidateCode: "CAN-003", Designation: "CLI Dev", Skill: "Go", ExpectedCTC: 160000, JobCode: "JOB1"},
	}
	m.columns[1].Cards = []models.ApplicationCard{
		{ApplicationID: 3, CandidateName: "Priya Sharma", CandidateCode: "CAN-002", Designation: "Backend", Skill: "PostgreSQL", ExpectedCTC: 175000, JobCode: "JOB2"},
	}
	m.activeCol = 0
	m.columns[0].Cursor = 0
	m.columns[1].Cursor = 0

	view := m.View()
	if !strings.Contains(view, "Alex Rivera") || !strings.Contains(view, "Priya Sharma") {
		t.Fatalf("expected view to contain candidate names")
	}

	// Verify rounded border characters are present for full card borders
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╰") {
		t.Fatalf("expected view to contain rounded border characters")
	}
}
