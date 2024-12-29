package store

import (
	"context"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (suite *TestStoreSuite) TestUserGet() {
	suite.Run("found", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).UTC()
		_, err := suite.store.db.Exec(`
INSERT INTO users (id, profile, session, created_at, updated_at)
VALUES (1, '{"id": 1, "first_name": "Test"}', '{}', ?1, ?2)`, now, now)

		suite.Require().NoError(err)

		user, err := suite.store.UserGet(context.Background(), 1)
		suite.Require().NoError(err)
		suite.Equal(models.Profile{ID: 1, FirstName: "Test"}, user.Profile)
		suite.Equal(now, user.CreatedAt)
		suite.Equal(now, user.UpdatedAt)
	})

	suite.Run("not found", func() {
		user, err := suite.store.UserGet(context.Background(), 1)
		suite.NoError(err)
		suite.Nil(user)
	})
}

func (suite *TestStoreSuite) TestUserUpsertProfile() {
	suite.Run("new user", func() {
		user := &models.User{
			Profile:  models.Profile{ID: 1, FirstName: "Test"},
			Settings: models.UserSettings{Event: models.EventSettings{AutoPairing: true}},
		}
		err := suite.store.UserUpsertProfile(context.Background(), user)
		suite.Require().NoError(err)

		got, err := suite.store.UserGet(context.Background(), 1)
		suite.Require().NoError(err)
		suite.Equal(user.Profile, got.Profile)
		suite.Equal(user.Settings, got.Settings)
		suite.NotEmpty(user.CreatedAt)
		suite.NotEmpty(user.UpdatedAt)
	})

	suite.Run("update existing user profile", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).Add(-1 * time.Second).UTC()
		_, err := suite.store.db.Exec(`
INSERT INTO users (id, profile, session, settings, created_at, updated_at)
VALUES (1, '{"id": 1, "first_name": "Test"}', '{}', '{"event": {"auto_pairing": true}}', ?1, ?2)`, now, now)
		suite.Require().NoError(err)

		user := &models.User{
			Profile: models.Profile{ID: 1, FirstName: "Test2"},
		}

		err = suite.store.UserUpsertProfile(context.Background(), user)
		suite.Require().NoError(err)
		suite.Equal(models.Profile{ID: 1, FirstName: "Test2"}, user.Profile)
		suite.Equal(models.UserSettings{Event: models.EventSettings{AutoPairing: true}}, user.Settings)
		suite.NotEmpty(user.CreatedAt)
		suite.NotEmpty(user.UpdatedAt)
		suite.NotEqual(user.CreatedAt, user.UpdatedAt)
	})
}

func (suite *TestStoreSuite) TestUserUpdateSession() {
	suite.Run("update session", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).Add(-1 * time.Second).UTC()
		_, err := suite.store.db.Exec(`
		INSERT INTO users (id, profile, session, created_at, updated_at)
		VALUES (1, '{"id": 1, "first_name": "Test"}', '{}', ?1, ?2)`, now, now)
		suite.Require().NoError(err)

		user := &models.User{
			Profile: models.Profile{ID: 1},
			Session: models.Session{EventID: "test_event"},
		}

		err = suite.store.UserUpdateSession(context.Background(), user)
		suite.Require().NoError(err)

		got, err := suite.store.UserGet(context.Background(), 1)
		suite.Require().NoError(err)
		suite.Equal(user.Session, got.Session)
		suite.Equal("test_event", got.Session.EventID)
		suite.NotEmpty(got.UpdatedAt)
		suite.NotEqual(got.CreatedAt, got.UpdatedAt)
	})

	suite.Run("user not found", func() {
		user := &models.User{
			Profile: models.Profile{ID: 2},
			Session: models.Session{EventID: "test_event"},
		}

		err := suite.store.UserUpdateSession(context.Background(), user)
		suite.Require().ErrorIs(err, ErrNotFound)
	})
}

func (suite *TestStoreSuite) TestUserUpdateSettings() {
	suite.Run("update settings", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).Add(-1 * time.Second).UTC()
		_, err := suite.store.db.Exec(`
		INSERT INTO users (id, profile, session, settings, created_at, updated_at)
		VALUES (1, '{"id": 1, "first_name": "Test"}', '{}', '{}', ?1, ?2)`, now, now)
		suite.Require().NoError(err)

		user := &models.User{
			Profile:  models.Profile{ID: 1},
			Settings: models.UserSettings{Event: models.EventSettings{AutoPairing: true}},
		}

		err = suite.store.UserUpdateSettings(context.Background(), user)
		suite.Require().NoError(err)

		got, err := suite.store.UserGet(context.Background(), 1)
		suite.Require().NoError(err)
		suite.Equal(user.Settings, got.Settings)
		suite.Equal(models.EventSettings{AutoPairing: true}, got.Settings.Event)
		suite.NotEmpty(got.UpdatedAt)
		suite.NotEqual(got.CreatedAt, got.UpdatedAt)
	})

	suite.Run("user not found", func() {
		user := &models.User{
			Profile:  models.Profile{ID: 2},
			Settings: models.UserSettings{Event: models.EventSettings{AutoPairing: true}},
		}

		err := suite.store.UserUpdateSettings(context.Background(), user)
		suite.Require().ErrorIs(err, ErrNotFound)
	})
}
