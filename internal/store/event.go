package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

// EventGet returns an event by its id.
// If the event does not exist, returns nil.
func (s *SQLiteStore) EventGet(ctx context.Context, eventID string) (*models.Event, error) {
	const query = `SELECT data FROM events WHERE id = ?1`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}
	var data []byte
	if err = stmt.QueryRowxContext(ctx, eventID).Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	event := &models.Event{}
	if err = s.unmarshal("data", data, event); err != nil {
		return nil, err
	}

	return event, nil
}

// EventUpsert creates new event or updates existing one.
func (s *SQLiteStore) EventUpsert(ctx context.Context, event *models.Event) error {
	const query =
	// language=SQLite
	`INSERT INTO events (id, owner_id, data)
VALUES (?1, ?2, ?3)
ON CONFLICT (id) DO UPDATE SET owner_id   = excluded.owner_id,
                               data       = CASE
                                                WHEN ifnull(json_extract(events.data, '$.subscribers_notified'), FALSE)
                                                    THEN json_set(excluded.data, '$.subscribers_notified', json('true'))
                                                ELSE excluded.data
                                   END,
                               updated_at = CURRENT_TIMESTAMP;`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := s.marshal("data", event)
	if err != nil {
		return err
	}

	if _, err = stmt.ExecContext(ctx, event.ID, event.Owner.ID, data); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return nil
}

// EventSubscribersNotifiedSet atomically sets the event SubscribersNotified attribute.
// Returns false if the attribute is already set or the event does not exist.
func (s *SQLiteStore) EventSubscribersNotifiedSet(ctx context.Context, eventID string) (bool, error) {
	const query =
	// language=SQLite
	`UPDATE events
SET data = json_set(data, '$.subscribers_notified', json('true'))
WHERE id = ?1
  AND ifnull(json_extract(data, '$.subscribers_notified'), FALSE) = FALSE
RETURNING id;`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	var id string
	if err = stmt.QueryRowxContext(ctx, eventID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	return true, nil
}

// EventGetUpdatedAfter returns all non-draft events updated after the specified time.
func (s *SQLiteStore) EventGetUpdatedAfter(ctx context.Context, after time.Time) ([]*models.Event, error) {
	// language=SQLite
	const query = `SELECT data
FROM events
WHERE updated_at > ?1
  AND json_extract(data, '$.post.inline_message_id') IS NOT NULL`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	rows, err := stmt.QueryxContext(ctx, after)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrScan, err)
	}
	//goland:noinspection ALL
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrScan, err)
		}

		event := &models.Event{}
		if err = s.unmarshal("data", data, event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// EventRemoveDraftsBefore removes all draft events updated before the specified time.
// Returns the ids of the removed events.
func (s *SQLiteStore) EventRemoveDraftsBefore(ctx context.Context, before time.Time) ([]string, error) {
	// language=SQLite
	const query = `DELETE
FROM events
WHERE updated_at < ?1
  AND json_extract(data, '$.post.inline_message_id') IS NULL
  AND ifnull(json_array_length(data, '$.couples'), 0) = 0
  AND ifnull(json_array_length(data, '$.singles'), 0) = 0
RETURNING id;`

	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	rows, err := stmt.QueryxContext(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	//goland:noinspection ALL
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrScan, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// EventGetMy returns slice of event id related to the specified profile:
//   - non-draft and not removed events owned by the user
//   - not removed events where the user is a participant: either in a couple or as a single
func (s *SQLiteStore) EventGetMy(ctx context.Context, profile *models.Profile) ([]string, error) {
	if profile == nil {
		return nil, ErrNil
	}
	// language=SQLite
	const query = `SELECT id, owner_id, data
FROM events
WHERE json_extract(data, '$.post.inline_message_id') IS NOT NULL -- Skip draft events
  AND ifnull(json_extract(data, '$.removed'), FALSE) == FALSE -- Skip removed events
ORDER BY created_at DESC
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	rows, err := stmt.QueryxContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}
	//goland:noinspection ALL
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		var ownerID int64
		var data []byte
		if err = rows.Scan(&id, &ownerID, &data); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrScan, err)
		}

		if ownerID == profile.ID {
			ids = append(ids, id)
			continue
		}

		event := &models.Event{}
		if err = s.unmarshal("data", data, event); err != nil {
			return nil, err
		}
		if eventHasProfile(event, profile) {
			ids = append(ids, id)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrScan, err)
	}

	return ids, nil
}

func eventHasProfile(event *models.Event, profile *models.Profile) bool {
	if event == nil || profile == nil {
		return false
	}
	for _, couple := range event.Couples {
		for _, dancer := range couple.Dancers {
			if dancerMatchesProfile(dancer, profile) {
				return true
			}
		}
	}
	for _, single := range event.Singles {
		if dancerMatchesProfile(single, profile) {
			return true
		}
	}
	return false
}

func dancerMatchesProfile(dancer models.Dancer, profile *models.Profile) bool {
	if profile == nil {
		return false
	}
	if dancer.Profile != nil {
		if dancer.Profile.ID == profile.ID {
			return true
		}
		return profile.Username != "" && strings.EqualFold(dancer.Profile.Username, profile.Username)
	}
	username, ok := usernameFromFullName(dancer.FullName)
	return ok && profile.Username != "" && strings.EqualFold(username, profile.Username)
}

var reFullNameUsername = regexp.MustCompile(`(?:^|\b|\s)@([a-zA-Z][a-zA-Z0-9_]{3,30}[a-zA-Z0-9])(?:\b|$)`)

func usernameFromFullName(s string) (string, bool) {
	matches := reFullNameUsername.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}
