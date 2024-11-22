package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (s *Store) HistoryInsert(ctx context.Context, item *models.HistoryItem) error {
	// language=SQLite
	const query = `INSERT INTO history (action, profile_id, event_id, data)
VALUES (?1, ?2, ?3, $4);`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshal, err)
	}

	var profileID *int64
	if item.Profile != nil {
		profileID = &item.Profile.ID
	}

	if _, err = stmt.ExecContext(ctx, item.Action, profileID, item.EventID, data); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return nil
}
