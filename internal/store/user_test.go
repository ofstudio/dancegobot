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

func (suite *TestStoreSuite) TestUserUpsert() {
	suite.Run("insert new user", func() {
		// Insert new user into the database
		user := &models.User{
			Profile:  models.Profile{ID: 2, FirstName: "NewUser"},
			Session:  models.Session{Action: "start"},
			Settings: models.UserSettings{},
		}
		err := suite.store.UserUpsert(context.Background(), user)
		suite.Require().NoError(err)

		// Get user from the database
		insertedUser, err := suite.store.UserGet(context.Background(), 2)
		suite.Require().NoError(err)
		suite.Equal(user.Profile, insertedUser.Profile)
		suite.Equal(user.Session, insertedUser.Session)
	})

	suite.Run("update existing user", func() {
		// Add user to the database
		now := time.Now().Truncate(time.Second).UTC()
		_, err := suite.store.db.Exec(`
		INSERT INTO users (id, profile, session, created_at, updated_at)
		VALUES (3, '{"id": 3, "first_name": "ExistingUser"}', '{}', ?1, ?2)`, now, now)
		suite.Require().NoError(err)

		// Update user in the database
		updatedUser := &models.User{
			Profile:  models.Profile{ID: 3, FirstName: "UpdatedUser"},
			Session:  models.Session{},
			Settings: models.UserSettings{},
		}
		err = suite.store.UserUpsert(context.Background(), updatedUser)
		suite.Require().NoError(err)

		// Get user from the database
		user, err := suite.store.UserGet(context.Background(), 3)
		suite.Require().NoError(err)
		suite.Equal(updatedUser.Profile, user.Profile)
	})
}
