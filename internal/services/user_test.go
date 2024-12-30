package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
)

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

type UserServiceTestSuite struct {
	suite.Suite
	store store.Store
	srv   *UserService
}

func (suite *UserServiceTestSuite) SetupSubTest() {
	db, err := store.NewSQLite(":memory:", config.Default().DB.Version)
	suite.Require().NoError(err)
	suite.store = store.NewSQLiteStore(db)
	suite.srv = NewUserService(config.Default().Settings, suite.store)
}

func (suite *UserServiceTestSuite) TearDownSubTest() {
	suite.store.Close()
	suite.store = nil
	suite.srv = nil
}

func (suite *UserServiceTestSuite) TestGet() {
	ctx := context.Background()

	suite.Run("existing user", func() {
		profile := models.Profile{ID: 1, FirstName: "John", LastName: "Doe"}
		session := models.Session{EventID: "test_event_id"}
		user := &models.User{Profile: profile, Session: session}
		suite.Require().NoError(suite.store.UserUpsertProfile(ctx, user))
		suite.Require().NotNil(user)

		user, err := suite.srv.Get(ctx, profile)
		suite.Require().NoError(err)
		suite.Require().NotNil(user)
		suite.Equal(int64(1), user.Profile.ID)
		suite.Equal("John", user.Profile.FirstName)
		suite.Equal("Doe", user.Profile.LastName)
		suite.Equal("test_event_id", user.Session.EventID)
	})

	suite.Run("new user", func() {
		profile := models.Profile{ID: 2, FirstName: "Jane", LastName: "Smith"}
		user, err := suite.srv.Get(ctx, profile)
		suite.Require().NoError(err)
		suite.Require().NotNil(user)
		suite.Equal(int64(2), user.Profile.ID)
		suite.Equal("Jane", user.Profile.FirstName)
		suite.Equal("Smith", user.Profile.LastName)
	})

	suite.Run("profile updated", func() {
		profile := models.Profile{ID: 1, FirstName: "John", LastName: "Doe"}
		session := models.Session{EventID: "test_event_id"}
		user := &models.User{Profile: profile, Session: session}
		suite.Require().NoError(suite.store.UserUpsertProfile(ctx, user))
		suite.Require().NotNil(user)

		newProfile := models.Profile{ID: 1, FirstName: "Jack", LastName: "Smith", Username: "jacksmith"}
		user, err := suite.srv.Get(ctx, newProfile)
		suite.Require().NoError(err)
		suite.Require().NotNil(user)
		suite.Equal(int64(1), user.Profile.ID)
		suite.Equal("Jack", user.Profile.FirstName)
		suite.Equal("Smith", user.Profile.LastName)
		suite.Equal("jacksmith", user.Profile.Username)
		suite.Equal("test_event_id", user.Session.EventID)
	})

	suite.Run("profile ID is 0", func() {
		profile := models.Profile{ID: 0}
		user, err := suite.srv.Get(ctx, profile)
		suite.Require().Error(err)
		suite.Require().Nil(user)
	})
}
