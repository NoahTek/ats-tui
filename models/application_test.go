package models

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDatabaseApplicationsFetch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5434/ats")
	if err != nil {
		t.Skipf("skipping db test: postgres not reachable (%v)", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping db test: postgres ping failed (%v)", err)
		return
	}

	stageMap, err := FetchBoardApplications(ctx, pool, "ALL")
	if err != nil {
		t.Fatalf("FetchBoardApplications failed: %v", err)
	}

	totalCards := 0
	for _, cards := range stageMap {
		totalCards += len(cards)
	}

	if totalCards == 0 {
		t.Fatalf("expected seeded cards, got 0")
	}
}

func TestStageHelpers(t *testing.T) {
	if StageInterested.Title() != "Interested" {
		t.Errorf("expected Interested, got %s", StageInterested.Title())
	}
	if StageFollowUp.Title() != "Follow-Up" {
		t.Errorf("expected Follow-Up, got %s", StageFollowUp.Title())
	}
	if StageCallBack.Title() != "Call-Back" {
		t.Errorf("expected Call-Back, got %s", StageCallBack.Title())
	}
	if StageClosed.Title() != "Closed" {
		t.Errorf("expected Closed, got %s", StageClosed.Title())
	}
}
