package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NoteRecord represents a comment or feedback entry for an application.
type NoteRecord struct {
	ID            int
	ApplicationID int
	NoteType      string
	Content       string
	CreatedAt     string
}

// FetchApplicationNotes queries all notes for a specific application in chronological order.
func FetchApplicationNotes(ctx context.Context, pool *pgxpool.Pool, appID int) ([]NoteRecord, error) {
	query := `
		SELECT 
			id, 
			application_id, 
			COALESCE(note_type, 'Comment'), 
			content, 
			to_char(created_at, 'YYYY-MM-DD HH24:MI') 
		FROM notes 
		WHERE application_id = $1
		ORDER BY created_at ASC;
	`
	rows, err := pool.Query(ctx, query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch application notes: %w", err)
	}
	defer rows.Close()

	var notes []NoteRecord
	for rows.Next() {
		var n NoteRecord
		err := rows.Scan(
			&n.ID,
			&n.ApplicationID,
			&n.NoteType,
			&n.Content,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan note failed: %w", err)
		}
		notes = append(notes, n)
	}

	return notes, nil
}
