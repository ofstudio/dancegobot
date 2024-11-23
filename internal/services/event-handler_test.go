package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventHandler(t *testing.T) {
	suite.Run(t, new(TestEventHandlerSuite))
}

type TestEventHandlerSuite struct {
	suite.Suite
}

func (suite *TestEventHandlerSuite) TestDancerGetByProfile() {
	suite.Run("dancer is not registered", func() {
		event := sampleEventFull
		profile := &models.Profile{ID: 100, FirstName: "Test", LastName: "User"}

		dancer := newEventHandler(&event).DancerGetByProfile(profile, models.RoleLeader)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusNotRegistered, dancer.Status)
		suite.Equal(profile.ID, dancer.Profile.ID)
		suite.Equal(profile.FullName(), dancer.FullName)
		suite.Equal(models.RoleLeader, dancer.Role)
		suite.Equal(false, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})

	suite.Run("dancer is registered in couple by username", func() {
		event := sampleEventFull
		profile := &models.Profile{ID: 20, FirstName: "Jill", LastName: "Smith", Username: "jillsmith"}

		dancer := newEventHandler(&event).DancerGetByProfile(profile, models.RoleLeader)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusInCouple, dancer.Status)
		suite.Require().NotNil(dancer.Partner)
		suite.Equal("Jack Smith", dancer.Partner.FullName)
		suite.Equal(models.RoleFollower, dancer.Role)
		suite.Equal(false, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})

	suite.Run("dancer is registered in couple by profile", func() {
		event := sampleEventFull
		profile := &models.Profile{ID: 3, FirstName: "Jack", LastName: "Smith"}

		dancer := newEventHandler(&event).DancerGetByProfile(profile, models.RoleLeader)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusInCouple, dancer.Status)
		suite.Require().NotNil(dancer.Partner)
		suite.Equal("@jillsmith", dancer.Partner.FullName)
		suite.Equal("Jack Smith", dancer.FullName)
		suite.Equal(models.RoleLeader, dancer.Role)
		suite.Equal(true, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})

	suite.Run("dancer is registered as single", func() {
		event := sampleEventFull
		profile := &models.Profile{ID: 5}

		dancer := newEventHandler(&event).DancerGetByProfile(profile, models.RoleFollower)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusSingle, dancer.Status)
		suite.Nil(dancer.Partner)
		suite.Equal("Amalia Green", dancer.FullName)
		suite.Equal(models.RoleFollower, dancer.Role)
		suite.Equal(true, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})
}

func (suite *TestEventHandlerSuite) TestDancerGetByName() {
	suite.Run("dancer is not registered", func() {
		event := sampleEventFull
		name := "@testuser"

		dancer := newEventHandler(&event).DancerGetByName(name, models.RoleLeader)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusNotRegistered, dancer.Status)
		suite.Nil(dancer.Profile)
		suite.Equal(name, dancer.FullName)
		suite.Equal(models.RoleLeader, dancer.Role)
		suite.Equal(false, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})

	suite.Run("dancer is registered in couple by username", func() {
		event := sampleEventFull
		name := "@jillsmith"

		dancer := newEventHandler(&event).DancerGetByName(name, models.RoleLeader)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusInCouple, dancer.Status)
		suite.Require().NotNil(dancer.Partner)
		suite.Equal("Jack Smith", dancer.Partner.FullName)
		suite.Equal(models.RoleFollower, dancer.Role)
		suite.Equal(false, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})

	suite.Run("dancer is registered in couple by profile", func() {
		event := sampleEventFull
		name := "@johndoe"

		dancer := newEventHandler(&event).DancerGetByName(name, models.RoleLeader)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusInCouple, dancer.Status)
		suite.Require().NotNil(dancer.Partner)
		suite.Equal("Jane Doe", dancer.Partner.FullName)
		suite.Require().NotNil(dancer.Profile)
		suite.Equal(int64(1), dancer.Profile.ID)
		suite.Equal("John Doe", dancer.FullName)
		suite.Equal(models.RoleLeader, dancer.Role)
		suite.Equal(false, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})

	suite.Run("dancer is registered as single", func() {
		event := sampleEventFull
		name := "@katbrown"

		dancer := newEventHandler(&event).DancerGetByName(name, models.RoleFollower)

		suite.Require().NotNil(dancer)
		suite.Equal(models.StatusSingle, dancer.Status)
		suite.Nil(dancer.Partner)
		suite.Require().NotNil(dancer.Profile)
		suite.Equal(int64(4), dancer.Profile.ID)
		suite.Equal("Kate Brown", dancer.FullName)
		suite.Equal(models.RoleFollower, dancer.Role)
		suite.Equal(true, dancer.SingleSignup)
		suite.NotEmpty(dancer.CreatedAt)
	})
}

var sampleEventFull = models.Event{
	ID:        "test12345678",
	Caption:   "This is a test event",
	MessageID: "123test456",
	Couples: []models.Couple{
		{
			Dancers: []models.Dancer{
				{
					Profile:   &models.Profile{ID: 1, FirstName: "John", LastName: "Doe", Username: "johndoe"},
					FullName:  "John Doe",
					Role:      models.RoleLeader,
					CreatedAt: nowFn(),
				},
				{
					Profile:      &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe", Username: "janedoe"},
					FullName:     "Jane Doe",
					Role:         models.RoleFollower,
					SingleSignup: true,
					CreatedAt:    nowFn().Add(-1 * time.Minute),
				},
			},
			CreatedBy: models.Profile{ID: 1, FirstName: "John", LastName: "Doe", Username: "johndoe"},
			CreatedAt: nowFn(),
		},
		{
			Dancers: []models.Dancer{
				{
					Profile:      &models.Profile{ID: 3, FirstName: "Jack", LastName: "Smith"},
					FullName:     "Jack Smith",
					Role:         models.RoleLeader,
					SingleSignup: true,
					CreatedAt:    nowFn().Add(-3 * time.Minute),
				},
				{
					FullName:  "@jillsmith",
					Role:      models.RoleFollower,
					CreatedAt: nowFn(),
				},
			},
			CreatedBy: models.Profile{ID: 3, FirstName: "Jack", LastName: "Smith"},
			CreatedAt: nowFn(),
		},
	},
	Singles: []models.Dancer{
		{
			Profile:      &models.Profile{ID: 4, FirstName: "Kate", LastName: "Brown", Username: "katbrown"},
			FullName:     "Kate Brown",
			Role:         models.RoleFollower,
			SingleSignup: true,
			CreatedAt:    nowFn().Add(-2 * time.Minute),
		},
		{
			Profile:      &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"},
			FullName:     "Amalia Green",
			Role:         models.RoleFollower,
			SingleSignup: true,
			CreatedAt:    nowFn().Add(-4 * time.Minute),
		},
	},
	Owner:     models.Profile{ID: 1000, FirstName: "Test", LastName: "Owner"},
	CreatedAt: nowFn(),
}
