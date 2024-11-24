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

func (suite *TestEventHandlerSuite) TestCoupleAdd() {
	suite.Run("both dancers not registered", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"}
		profile2 := &models.Profile{ID: 700, FirstName: "Bob", LastName: "Builder"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultSuccess, result.Result)
		suite.Equal(models.StatusInCouple, dancer1.Status)
		suite.Equal(models.StatusInCouple, dancer2.Status)
		suite.Require().NotNil(dancer1.Partner)
		suite.Require().NotNil(dancer2.Partner)
		suite.Equal(dancer1.Partner, dancer2)
		suite.Equal(dancer2.Partner, dancer1)
		suite.Equal(dancer2.Partner, dancer1)

		suite.Require().NotNil(result.Couple)
		suite.Equal(dancer1, &result.Couple.Dancers[0])
		suite.Equal(dancer2, &result.Couple.Dancers[1])
		suite.Equal(dancer1.Profile, &result.Couple.CreatedBy)
		suite.NotEmpty(result.Couple.CreatedAt)

		suite.Require().Len(handler.hist, 1)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[0].Action)
		suite.Equal(dancer1.Profile, handler.hist[0].Initiator)
		suite.Equal(result.Couple, handler.hist[0].Details)

		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("partner registered as single", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 600, FirstName: "Bob", LastName: "Sponge"}
		profile2 := &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultSuccess, result.Result)
		suite.Equal(models.StatusInCouple, dancer1.Status)
		suite.Equal(models.StatusInCouple, dancer2.Status)
		suite.Require().NotNil(dancer1.Partner)
		suite.Require().NotNil(dancer2.Partner)
		suite.Equal(dancer1.Partner, dancer2)
		suite.Equal(dancer2.Partner, dancer1)
		suite.True(dancer2.SingleSignup)

		suite.Require().NotNil(result.Couple)
		suite.Equal(dancer1, &result.Couple.Dancers[0])
		suite.Equal(dancer2, &result.Couple.Dancers[1])

		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistorySingleRemoved, handler.hist[0].Action)
		suite.Equal(dancer1.Profile, handler.hist[0].Initiator)
		suite.Equal(dancer2, handler.hist[0].Details)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[1].Action)
		suite.Equal(dancer1.Profile, handler.hist[1].Initiator)
		suite.Equal(result.Couple, handler.hist[1].Details)

		suite.Require().Len(handler.notif, 1)
		suite.Equal(dancer1.Profile, handler.notif[0].Initiator)
		suite.Equal(dancer2.Profile, &handler.notif[0].Recipient)
		suite.Equal(models.TmplRegisteredWithSingle, handler.notif[0].TmplCode)
		suite.Equal(event, *handler.notif[0].Event)
	})

	suite.Run("dancer registered as single", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"}
		profile2 := &models.Profile{ID: 700, FirstName: "Bob", LastName: "Builder"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleFollower)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleLeader)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultSuccess, result.Result)
		suite.Equal(models.StatusInCouple, dancer1.Status)
		suite.Equal(models.StatusInCouple, dancer2.Status)
		suite.Require().NotNil(dancer1.Partner)
		suite.Require().NotNil(dancer2.Partner)
		suite.Equal(dancer1.Partner, dancer2)
		suite.Equal(dancer2.Partner, dancer1)
		suite.True(dancer1.SingleSignup)

		suite.Require().NotNil(result.Couple)
		suite.Equal(dancer2, &result.Couple.Dancers[0])
		suite.Equal(dancer1, &result.Couple.Dancers[1])

		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistorySingleRemoved, handler.hist[0].Action)
		suite.Equal(dancer1.Profile, handler.hist[0].Initiator)
		suite.Equal(dancer1, handler.hist[0].Details)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[1].Action)
		suite.Equal(dancer1.Profile, handler.hist[1].Initiator)
		suite.Equal(result.Couple, handler.hist[1].Details)

		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("one dancer registered in another couple", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 100, FirstName: "Charlie", LastName: "Brown"}
		profile2 := &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultPartnerTaken, result.Result)
		suite.Equal(models.StatusNotRegistered, dancer1.Status)
		suite.Equal(models.StatusInCouple, dancer2.Status)
		suite.Nil(dancer1.Partner)
		suite.Equal(dancer2, result.ChosenPartner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)

	})

	suite.Run("dancer already registered in another couple", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 1, FirstName: "John", LastName: "Doe"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByName("@someuser", models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultAlreadyInCouple, result.Result)
		suite.Equal(models.StatusInCouple, dancer1.Status)
		suite.Equal(models.StatusNotRegistered, dancer2.Status)
		suite.Require().NotNil(dancer1.Partner)
		suite.Nil(dancer2.Partner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer already registered in same couple", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 1, FirstName: "John", LastName: "Doe"}
		profile2 := &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultAlreadyInSameCouple, result.Result)
		suite.Equal(models.StatusInCouple, dancer1.Status)
		suite.Equal(models.StatusInCouple, dancer2.Status)
		suite.Require().NotNil(dancer1.Partner)
		suite.Require().NotNil(dancer2.Partner)
		suite.Equal(dancer1.Partner.Profile, dancer2.Profile)
		suite.Equal(dancer2.Partner.Profile, dancer1.Profile)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer and partner have the same role", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 400, FirstName: "Eve", LastName: "Green"}
		profile2 := &models.Profile{ID: 4, FirstName: "Kate", LastName: "Brown"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleFollower)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleLeader)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultPartnerSameRole, result.Result)
		suite.Equal(models.StatusNotRegistered, dancer1.Status)
		suite.Equal(models.StatusSingle, dancer2.Status)
		suite.Nil(dancer1.Partner)
		suite.Nil(dancer2.Partner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer tries to register with itself", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 4, FirstName: "Kate", LastName: "Brown"}
		handler := newEventHandler(&event)

		dancer1 := handler.DancerGetByProfile(profile1, models.RoleFollower)
		dancer2 := handler.DancerGetByName("@katbrown", models.RoleLeader)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultSelfNotAllowed, result.Result)
		suite.Equal(dancer1, dancer2)
		suite.Equal(models.StatusSingle, dancer1.Status)
		suite.Equal(models.StatusSingle, dancer2.Status)
		suite.Nil(dancer1.Partner)
		suite.Nil(dancer2.Partner)
		suite.Require().Len(handler.hist, 0)
	})

	suite.Run("event is closed for new registrations", func() {
		event := sampleEventFull
		event.Closed = true
		profile1 := &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"}
		profile2 := &models.Profile{ID: 700, FirstName: "Bob", LastName: "Builder"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByProfile(profile2, models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultEventClosed, result.Result)
		suite.Equal(models.StatusNotRegistered, dancer1.Status)
		suite.Equal(models.StatusNotRegistered, dancer2.Status)
		suite.Len(handler.hist, 0)
		suite.Len(handler.notif, 0)
	})

	suite.Run("event is forbidden for dancer", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 100, FirstName: "Charlie", LastName: "Brown"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer1.Status = models.StatusForbidden
		dancer2 := handler.DancerGetByName("@someuser", models.RoleFollower)

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultEventForbiddenDancer, result.Result)
		suite.Equal(models.StatusForbidden, dancer1.Status)
		suite.Equal(models.StatusNotRegistered, dancer2.Status)
		suite.Len(handler.hist, 0)
		suite.Len(handler.notif, 0)
	})

	suite.Run("event is forbidden for partner", func() {
		event := sampleEventFull
		profile1 := &models.Profile{ID: 100, FirstName: "Charlie", LastName: "Brown"}
		handler := newEventHandler(&event)
		dancer1 := handler.DancerGetByProfile(profile1, models.RoleLeader)
		dancer2 := handler.DancerGetByName("@someuser", models.RoleFollower)
		dancer2.Status = models.StatusForbidden

		result := handler.CoupleAdd(dancer1, dancer2)

		suite.Require().NotNil(result)
		suite.Equal(models.ResultEventForbiddenPartner, result.Result)
		suite.Equal(models.StatusNotRegistered, dancer1.Status)
		suite.Equal(models.StatusForbidden, dancer2.Status)
		suite.Len(handler.hist, 0)
		suite.Len(handler.notif, 0)
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
