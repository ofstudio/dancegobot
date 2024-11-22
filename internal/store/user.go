package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

type userRow struct {
	ID        int64     `db:"id"`
	Profile   []byte    `db:"profile"`
	Session   []byte    `db:"session"`
	Settings  []byte    `db:"settings"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserGet returns user by its id.
// If the user does not exist, returns ErrNotFound.
func (s *Store) UserGet(ctx context.Context, id int64) (*models.User, error) {
	const query = `SELECT profile, session, settings, created_at, updated_at
FROM users
WHERE id = ?1
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	var row userRow
	if err = stmt.QueryRowxContext(ctx, id).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	user := &models.User{}
	if err = s.userUnmarshalRow(row, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UserProfileUpsert updates user profile. If the user does not exist, creates a new one.
func (s *Store) UserProfileUpsert(ctx context.Context, user *models.User) error {
	// language=SQLite
	const query = `INSERT INTO users (id, profile)
VALUES (?1, ?2)
ON CONFLICT (id) DO UPDATE SET profile    = excluded.profile,
                               updated_at = CURRENT_TIMESTAMP
RETURNING id, profile, session, settings, created_at, updated_at
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	data, err := json.Marshal(user.Profile)
	if err != nil {
		return fmt.Errorf("%w: user.profile, user.id=%d, %w", ErrMarshal, user.Profile.ID, err)
	}

	var row userRow
	if err = stmt.QueryRowxContext(ctx, user.Profile.ID, data).StructScan(&row); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return s.userUnmarshalRow(row, user)
}

// UserSessionUpsert updates user session and profile as well.
// If the user does not exist, creates a new one.
// Passed user model will be updated with the latest data from the database.
func (s *Store) UserSessionUpsert(ctx context.Context, user *models.User) error {
	// language=SQLite
	const query = `INSERT INTO users (id, profile, session)
VALUES (?1, ?2, ?3)
ON CONFLICT (id) DO UPDATE SET profile    = excluded.profile,
                               session    = excluded.session,
                               updated_at = CURRENT_TIMESTAMP
RETURNING id, profile, session, settings, created_at, updated_at
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	dataProfile, err := json.Marshal(user.Profile)
	if err != nil {
		return fmt.Errorf("%w: user.profile, user.id=%d, %w", ErrMarshal, user.Profile.ID, err)
	}
	dataSession, err := json.Marshal(user.Session)
	if err != nil {
		return fmt.Errorf("%w: user.session, user.id=%d, %w", ErrMarshal, user.Profile.ID, err)
	}

	var row userRow
	if err = stmt.QueryRowxContext(ctx, user.Profile.ID, dataProfile, dataSession).
		StructScan(&row); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return s.userUnmarshalRow(row, user)
}

// UserSettingsUpsert updates user settings and profile as well.
// If the user does not exist, creates a new one.
// Passed user model will be updated with the latest data from the database.
func (s *Store) UserSettingsUpsert(ctx context.Context, user *models.User) error {
	// language=SQLite
	const query = `INSERT INTO users (id, profile, settings)
VALUES (?1, ?2, ?3)
ON CONFLICT (id) DO UPDATE SET profile    = excluded.profile,
                               settings   = excluded.settings,
                               updated_at = CURRENT_TIMESTAMP
RETURNING id, profile, session, settings, created_at, updated_at
`
	stmt, err := s.stmt(ctx, query)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrStmtPrepare, err)
	}

	dataProfile, err := json.Marshal(user.Profile)
	if err != nil {
		return fmt.Errorf("%w: user.profile, user.id=%d, %w", ErrMarshal, user.Profile.ID, err)
	}

	dataSettings, err := json.Marshal(user.Settings)
	if err != nil {
		return fmt.Errorf("%w: user.settings, user.id=%d, %w", ErrMarshal, user.Profile.ID, err)
	}

	var row userRow
	if err = stmt.QueryRowxContext(ctx, user.Profile.ID, dataProfile, dataSettings).
		StructScan(&row); err != nil {
		return fmt.Errorf("%w: %w", ErrStmtExec, err)
	}

	return s.userUnmarshalRow(row, user)
}

func (s *Store) userUnmarshalRow(row userRow, user *models.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	if err := json.Unmarshal(row.Profile, &user.Profile); err != nil {
		return fmt.Errorf("%w: user.profile: %w", ErrUnmarshal, err)
	}
	if err := json.Unmarshal(row.Session, &user.Session); err != nil {
		return fmt.Errorf("%w: user.context: %w", ErrUnmarshal, err)
	}
	if err := json.Unmarshal(row.Settings, &user.Settings); err != nil {
		return fmt.Errorf("%w: user.settings: %w", ErrUnmarshal, err)
	}
	return nil
}
