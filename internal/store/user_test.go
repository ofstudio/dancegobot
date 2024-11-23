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

		// Get user from the database
		user, err := suite.store.UserGet(context.Background(), 1)
		suite.Require().NoError(err)
		suite.Equal(models.Profile{ID: 1, FirstName: "Test"}, user.Profile)
		suite.Equal(now, user.CreatedAt)
		suite.Equal(now, user.UpdatedAt)
	})

	suite.Run("not found", func() {
		// Get user from the database
		user, err := suite.store.UserGet(context.Background(), 1)
		suite.ErrorIs(err, ErrNotFound)
		suite.Nil(user)
	})
}

func (suite *TestStoreSuite) TestUserProfileUpsert() {
	suite.Run("insert", func() {
		user := &models.User{
			Profile: models.Profile{ID: 1, FirstName: "Test"}}

		err := suite.store.UserProfileUpsert(context.Background(), user)
		suite.Require().NoError(err)

		got, err := suite.store.UserGet(context.Background(), user.Profile.ID)
		suite.Require().NoError(err)
		suite.Equal(user, got)
	})

	suite.Run("update", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).Add(-1 * time.Minute).UTC()
		_, err := suite.store.db.Exec(`
INSERT INTO users (id, profile, session, created_at, updated_at)
VALUES (1, '{"id": 1, "first_name": "Test"}', '{"action":"event_leave"}', ?1, ?2)`, now, now)
		suite.Require().NoError(err)

		// Update user in the database
		newUser := &models.User{Profile: models.Profile{ID: 1, FirstName: "New Test"}}

		err = suite.store.UserProfileUpsert(context.Background(), newUser)
		suite.Require().NoError(err)

		// Check user is updated
		suite.Equal(int64(1), newUser.Profile.ID)
		suite.Equal("New Test", newUser.Profile.FirstName)
		suite.Equal(now, newUser.CreatedAt)
		suite.True(newUser.UpdatedAt.After(now))
		suite.Equal(models.Session{Action: "event_leave"}, newUser.Session)

		// Check user is updated in the database
		got, err := suite.store.UserGet(context.Background(), newUser.Profile.ID)
		suite.Require().NoError(err)
		suite.Equal(newUser, got)
	})
}

func (suite *TestStoreSuite) TestUserSessionUpsert() {
	suite.Run("insert", func() {
		user := &models.User{
			Profile: models.Profile{ID: 1, FirstName: "Test"},
			Session: models.Session{Action: "event_signup"},
		}

		err := suite.store.UserSessionUpsert(context.Background(), user)
		suite.Require().NoError(err)

		got, err := suite.store.UserGet(context.Background(), user.Profile.ID)
		suite.Require().NoError(err)
		suite.Equal(user, got)
	})

	suite.Run("update", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).Add(-1 * time.Minute).UTC()
		_, err := suite.store.db.Exec(`
INSERT INTO users (id, profile, session, created_at, updated_at)
VALUES (1, '{"id": 1, "first_name": "Test"}', '{"action":"event_leave"}', ?1, ?2)
`, now, now)
		suite.Require().NoError(err)

		// Update user in the database
		newUser := &models.User{
			Profile: models.Profile{ID: 1, FirstName: "Test2"},
			Session: models.Session{Action: "event_signup"},
		}

		err = suite.store.UserSessionUpsert(context.Background(), newUser)
		suite.Require().NoError(err)

		// Check user is updated
		suite.Equal(int64(1), newUser.Profile.ID)
		suite.Equal("Test2", newUser.Profile.FirstName)
		suite.Equal(now, newUser.CreatedAt)
		suite.True(newUser.UpdatedAt.After(now))
	})
}
