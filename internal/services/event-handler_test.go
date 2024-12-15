package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventHandler(t *testing.T) {
	suite.Run(t, new(TestEventHandlerSuite))
}

type TestEventHandlerSuite struct {
	suite.Suite
}

func (suite *TestEventHandlerSuite) TestDancerRegistrationGet() {
	suite.Run("dancer is not registered", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile: &models.Profile{ID: 100, FirstName: "Test", LastName: "User"},
			Role:    models.RoleLeader,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(d.Profile.ID, got.Profile.ID)
		suite.Equal(d.Profile.FullName(), got.Profile.FullName())
		suite.Equal(models.RoleLeader, got.Role)
		suite.Equal(false, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})

	suite.Run("dancer is registered in couple by username", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile: &models.Profile{ID: 20, FirstName: "Jill", LastName: "Smith", Username: "jillsmith"},
			Role:    models.RoleLeader,
		}
		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Require().NotNil(got.Partner)
		suite.Equal("Jack Smith", got.Partner.FullName)
		suite.Equal(models.RoleFollower, got.Role)
		suite.Equal(false, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})

	suite.Run("dancer is registered in couple by profile", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile: &models.Profile{ID: 3, FirstName: "Jack", LastName: "Smith"},
			Role:    models.RoleLeader,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Require().NotNil(got.Partner)
		suite.Equal("@jillsmith", got.Partner.FullName)
		suite.Equal("Jack Smith", got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.Equal(true, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})

	suite.Run("dancer is registered as single", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile: &models.Profile{ID: 5},
			Role:    models.RoleFollower,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusAsSingle, got.Status)
		suite.Nil(got.Partner)
		suite.Equal("Amalia Green", got.FullName)
		suite.Equal(models.RoleFollower, got.Role)
		suite.Equal(true, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})
}

func (suite *TestEventHandlerSuite) TestRegistrationGet_byName() {
	suite.Run("dancer is not registered", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile:  nil,
			FullName: "@testuser",
			Role:     models.RoleLeader,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Profile)
		suite.Equal(d.FullName, got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.Equal(false, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})

	suite.Run("dancer is registered in couple by username", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile:  nil,
			FullName: "@jillsmith",
			Role:     models.RoleLeader,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Require().NotNil(got.Partner)
		suite.Equal("Jack Smith", got.Partner.FullName)
		suite.Equal(models.RoleFollower, got.Role)
		suite.Equal(false, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})

	suite.Run("dancer is registered in couple by profile", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile:  nil,
			FullName: "@johndoe",
			Role:     models.RoleLeader,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Require().NotNil(got.Partner)
		suite.Equal("Jane Doe", got.Partner.FullName)
		suite.Require().NotNil(got.Profile)
		suite.Equal(int64(1), got.Profile.ID)
		suite.Equal("John Doe", got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.Equal(false, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})

	suite.Run("dancer is registered as single", func() {
		event := sampleEvent()
		d := &models.Dancer{
			Profile:  nil,
			FullName: "@katbrown",
			Role:     models.RoleFollower,
		}

		got := NewEventHandler(&event).RegistrationGet(d)

		suite.Require().NotNil(got)
		suite.Equal(models.StatusAsSingle, got.Status)
		suite.Nil(got.Partner)
		suite.Require().NotNil(got.Profile)
		suite.Equal(int64(4), got.Profile.ID)
		suite.Equal("Kate Brown", got.FullName)
		suite.Equal(models.RoleFollower, got.Role)
		suite.Equal(true, got.AsSingle)
		suite.NotEmpty(got.CreatedAt)
	})
}

func (suite *TestEventHandlerSuite) TestCoupleAdd() {
	suite.Run("both dancers not registered", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"},
			Role:    models.RoleLeader,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 700, FirstName: "Bob", LastName: "Builder"},
			Role:    models.RoleFollower,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Require().NotNil(got.Partner)
		suite.Require().NotNil(got.Related.Partner)
		suite.Equal(got.Partner, got.Related.Dancer)

		suite.Require().Len(event.Couples, 3)
		couple := event.Couples[2]
		suite.Equal(got.Dancer, &couple.Dancers[0])
		suite.Equal(got.Partner, &couple.Dancers[1])
		suite.Equal(got.Profile, couple.Dancers[0].Profile)
		suite.NotEmpty(couple.CreatedAt)

		suite.Require().Len(handler.hist, 1)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(&couple, handler.hist[0].Details)

		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("partner registered as single", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 600, FirstName: "Bob", LastName: "Sponge"},
			Role:    models.RoleLeader,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"},
			Role:    models.RoleFollower,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Require().NotNil(got.Partner)
		suite.Require().NotNil(got.Related.Partner)
		suite.Equal(got.Partner, got.Related.Dancer)
		suite.Equal(got.Related.Partner, got.Dancer)
		suite.True(got.Related.AsSingle)

		suite.Require().Len(event.Couples, 3)
		couple := event.Couples[2]
		suite.Equal(got.Dancer, &couple.Dancers[0])
		suite.Equal(got.Partner, &couple.Dancers[1])

		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistorySingleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(got.Related.Dancer, handler.hist[0].Details)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[1].Action)
		suite.Equal(got.Profile, handler.hist[1].Initiator)
		suite.Equal(&couple, handler.hist[1].Details)

		suite.Require().Len(handler.notif, 1)
		suite.Equal(got.Dancer, handler.notif[0].Payload.Partner)
		suite.Equal(got.Related.Profile, handler.notif[0].Recipient)
		suite.Equal(models.TmplRegisteredWithSingle, handler.notif[0].TmplCode)
		suite.Equal(event, *handler.notif[0].Payload.Event)
	})

	suite.Run("dancer registered as single", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"},
			Role:    models.RoleFollower,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 700, FirstName: "Bob", LastName: "Builder"},
			Role:    models.RoleLeader,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Require().NotNil(got.Partner)
		suite.Require().NotNil(got.Related.Partner)
		suite.Equal(got.Partner, got.Related.Dancer)
		suite.Equal(got.Related.Partner, got.Dancer)
		suite.True(got.AsSingle)

		suite.Require().Len(event.Couples, 3)
		couple := event.Couples[2]
		suite.Equal(got.Related.Dancer, &couple.Dancers[0])
		suite.Equal(got.Dancer, &couple.Dancers[1])

		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistorySingleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(got.Dancer, handler.hist[0].Details)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[1].Action)
		suite.Equal(got.Profile, handler.hist[1].Initiator)
		suite.Equal(&couple, handler.hist[1].Details)

		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("partner registered in another couple", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 100, FirstName: "Charlie", LastName: "Brown"},
			Role:    models.RoleLeader,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe"},
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultPartnerTaken, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Nil(got.Partner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer already registered in another couple", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 1, FirstName: "John", LastName: "Doe"},
			Role:    models.RoleLeader,
		}
		d2 := &models.Dancer{
			Profile:  nil,
			FullName: "@someuser",
			Role:     models.RoleFollower,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultAlreadyInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Equal(models.StatusNotRegistered, got.Related.Status)
		suite.Require().NotNil(got.Partner)
		suite.Nil(got.Related.Partner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer already registered in same couple", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 1, FirstName: "John", LastName: "Doe"},
			Role:    models.RoleLeader,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe"},
			Role:    models.RoleFollower,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultAlreadyInSameCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Require().NotNil(got.Partner)
		suite.Require().NotNil(got.Related.Partner)
		suite.Equal(got.Partner.Profile, got.Related.Profile)
		suite.Equal(got.Related.Partner.Profile, got.Profile)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer and partner have the same role", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 400, FirstName: "Eve", LastName: "Green"},
			Role:    models.RoleFollower,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 4, FirstName: "Kate", LastName: "Brown"},
			Role:    models.RoleLeader,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultPartnerSameRole, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(models.StatusAsSingle, got.Related.Status)
		suite.Nil(got.Partner)
		suite.Nil(got.Related.Partner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer tries to self register", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 4, FirstName: "Kate", LastName: "Brown"},
			Role:    models.RoleFollower,
		}
		d2 := &models.Dancer{
			Profile:  nil,
			FullName: "@katbrown",
			Role:     models.RoleLeader,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultSelfNotAllowed, got.Result)
		suite.Equal(got.Dancer, got.Related.Dancer)
		suite.Equal(models.StatusAsSingle, got.Status)
		suite.Equal(models.StatusAsSingle, got.Related.Status)
		suite.Nil(got.Partner)
		suite.Nil(got.Related.Partner)
		suite.Require().Len(handler.hist, 0)
	})

	suite.Run("event is closed for new registrations", func() {
		event := sampleEvent()
		event.Settings.ClosedFor = models.ClosedForAll
		handler := NewEventHandler(&event)
		d1 := &models.Dancer{
			Profile: &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"},
			Role:    models.RoleLeader,
		}
		d2 := &models.Dancer{
			Profile: &models.Profile{ID: 700, FirstName: "Bob", LastName: "Builder"},
			Role:    models.RoleFollower,
		}

		got := handler.CoupleAdd(d1, d2)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultEventClosed, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(models.StatusNotRegistered, got.Related.Status)
		suite.Len(handler.hist, 0)
		suite.Len(handler.notif, 0)
	})

}

func (suite *TestEventHandlerSuite) TestSingleAdd() {
	suite.Run("dancer not registered", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile:  &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"},
			FullName: "Alice Wonder",
			Role:     models.RoleLeader,
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredAsSingle, got.Result)
		suite.Equal(models.StatusAsSingle, got.Status)
		suite.True(got.AsSingle)
		suite.Nil(got.Partner)
		suite.Nil(got.Related)
		suite.Equal(d.Profile.ID, got.Profile.ID)
		suite.Equal(d.FullName, got.Dancer.FullName)
		suite.Equal(models.RoleLeader, got.Role)

		suite.Require().Len(event.Singles, 3)
		suite.Equal(event.Singles[2], *got.Dancer)

		suite.Require().Len(handler.hist, 1)
		suite.Equal(models.HistorySingleAdded, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(got.Dancer, handler.hist[0].Details)

		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer already registered as single", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"},
			Role:    models.RoleFollower,
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultAlreadyAsSingle, got.Result)
		suite.Equal(models.StatusAsSingle, got.Status)
		suite.True(got.AsSingle)
		suite.Nil(got.Partner)
		suite.Nil(got.Related)
		suite.Equal(d.ID, got.ID)
		suite.Equal(d.Profile.FullName(), got.FullName)
		suite.Equal(models.RoleFollower, got.Role)

		suite.Require().Len(event.Singles, 2)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer already registered in couple", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 1, FirstName: "John", LastName: "Doe"},
			Role:    models.RoleLeader,
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultAlreadyInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Require().NotNil(got.Partner)
		suite.Require().Nil(got.Related)
		suite.Equal(d.ID, got.ID)
		suite.Equal(d.Profile.FullName(), got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.False(got.AsSingle)

		suite.Require().Len(event.Singles, 2)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("event is closed for new registrations", func() {
		event := sampleEvent()
		event.Settings.ClosedFor = models.ClosedForAll
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile:  &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"},
			Role:     models.RoleLeader,
			FullName: "Alice Wonder",
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultEventClosed, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Nil(got.Related)
		suite.Equal(d.ID, got.ID)
		suite.Equal(d.FullName, got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.False(got.AsSingle)

		suite.Require().Len(event.Singles, 2)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("event is closed for singles", func() {
		event := sampleEvent()
		event.Settings.ClosedFor = models.ClosedForSingles
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile:  &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"},
			Role:     models.RoleLeader,
			FullName: "Alice Wonder",
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultClosedForSingles, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Nil(got.Related)
		suite.Equal(d.ID, got.ID)
		suite.Equal(d.FullName, got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.False(got.AsSingle)

		suite.Require().Len(event.Singles, 2)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("event is closed for single leaders", func() {
		event := sampleEvent()
		event.Settings.ClosedFor = models.ClosedForSingleLeaders
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile:  &models.Profile{ID: 600, FirstName: "Alice", LastName: "Wonder"},
			Role:     models.RoleLeader,
			FullName: "Alice Wonder",
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultClosedForSingleRole, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Nil(got.Related)
		suite.Equal(d.ID, got.ID)
		suite.Equal(d.FullName, got.FullName)
		suite.Equal(models.RoleLeader, got.Role)
		suite.False(got.AsSingle)

		suite.Require().Len(event.Singles, 2)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})
}

func (suite *TestEventHandlerSuite) TestSingleAdd_autoPair() {
	suite.Run("no matching partners", func() {
		event := sampleEvent()
		event.Settings.AutoPairing = true
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 600, FirstName: "Mary"},
			Role:    models.RoleFollower,
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredAsSingle, got.Result)
		suite.Equal(models.StatusAsSingle, got.Status)
		suite.True(got.AsSingle)
		suite.Equal(d.Profile, got.Profile)
		suite.Equal(d.Role, got.Role)

		suite.Nil(got.Partner)
		suite.Nil(got.Related)

		suite.Len(event.Singles, 3)
		suite.Equal(got.Dancer, &event.Singles[2])

		suite.Len(handler.hist, 1)
		suite.Equal(models.HistorySingleAdded, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(got.Dancer, handler.hist[0].Details)
		suite.Len(handler.notif, 0)
	})

	suite.Run("found matching partner", func() {
		config.SetBotProfile(botUser)
		event := sampleEvent()
		event.Settings.AutoPairing = true
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 700, FirstName: "Bobby", LastName: "Fisher"},
			Role:    models.RoleLeader,
		}
		wantPartner := event.Singles[0]

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.True(got.AsSingle)
		suite.Equal(d.Profile, got.Profile)
		suite.Equal(d.Role, got.Role)

		suite.Require().NotNil(got.Partner)
		suite.Equal(&wantPartner, got.Partner)
		suite.Require().NotNil(got.Related)
		suite.Equal(&wantPartner, got.Related.Dancer)
		suite.Equal(models.ResultRegisteredInCouple, got.Related.Result)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Equal(got.Dancer, got.Related.Partner)

		suite.Require().Len(event.Couples, 3)
		couple := event.Couples[2]
		suite.Equal(got.Dancer, &couple.Dancers[0])
		suite.Equal(got.Partner, &couple.Dancers[1])
		suite.True(couple.AutoPair)
		suite.Equal(got.Profile, &couple.CreatedBy)

		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistorySingleRemoved, handler.hist[0].Action)
		suite.Equal(&botProfile, handler.hist[0].Initiator)
		suite.Equal(got.Partner, handler.hist[0].Details)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[1].Action)
		suite.Equal(&botProfile, handler.hist[1].Initiator)
		suite.Equal(&couple, handler.hist[1].Details)

		suite.Require().Len(handler.notif, 1)
		suite.Equal(models.TmplAutoPairPartnerFound, handler.notif[0].TmplCode)
		suite.Equal(got.Partner.Profile, handler.notif[0].Recipient)
		suite.Equal(got.Dancer, handler.notif[0].Payload.Partner)
	})

	suite.Run("closed for single leaders but found matching partner", func() {
		config.SetBotProfile(botUser)
		event := sampleEvent()
		event.Settings.AutoPairing = true
		event.Settings.ClosedFor = models.ClosedForSingleLeaders
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 700, FirstName: "Bobby", LastName: "Fisher"},
			Role:    models.RoleLeader,
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegisteredInCouple, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
	})

	suite.Run("closed for single followers and no matching partners", func() {
		event := sampleEvent()
		event.Settings.AutoPairing = true
		event.Settings.ClosedFor = models.ClosedForSingleFollowers
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 600, FirstName: "Mary"},
			Role:    models.RoleFollower,
		}

		got := handler.SingleAdd(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultClosedForSingleRole, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
	})

}

func (suite *TestEventHandlerSuite) TestDancerRemove() {
	suite.Run("dancer not registered", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 100, FirstName: "Test", LastName: "User"},
		}

		got := handler.DancerRemove(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultWasNotRegistered, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Require().Len(handler.hist, 0)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer registered as single", func() {
		event := sampleEvent()
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"},
		}

		got := handler.DancerRemove(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Require().Len(event.Singles, 1)
		suite.Require().Len(handler.hist, 1)
		suite.Equal(models.HistorySingleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(got.Dancer, handler.hist[0].Details)
		suite.Require().Len(handler.notif, 0)
	})

	suite.Run("dancer registered in couple, partner is from singles list", func() {
		event := sampleEvent()
		couple := event.Couples[0]
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 1, FirstName: "John", LastName: "Doe"},
			Role:    models.RoleLeader,
		}
		partner := handler.RegistrationGet(d).Partner
		suite.Require().NotNil(partner)

		got := handler.DancerRemove(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Require().NotNil(got.Related)
		suite.Equal(partner, got.Related.Dancer)
		suite.Equal(models.StatusAsSingle, got.Related.Status)
		suite.Equal(models.ResultRegisteredAsSingle, got.Related.Result)
		suite.Require().Len(event.Couples, 1)
		suite.Require().Len(event.Singles, 3)
		suite.Equal(partner, &event.Singles[1])
		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistoryCoupleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(&couple, handler.hist[0].Details)
		suite.Equal(models.HistorySingleAdded, handler.hist[1].Action)
		suite.Equal(got.Profile, handler.hist[1].Initiator)
		suite.Equal(partner, handler.hist[1].Details)
		suite.Require().Len(handler.notif, 1)
		suite.Equal(got.Dancer, handler.notif[0].Payload.Partner)
		suite.Equal(partner.Profile, handler.notif[0].Recipient)
		suite.Equal(models.TmplCanceledWithSingle, handler.notif[0].TmplCode)
	})

	suite.Run("dancer registered in couple, partner is the couple creator", func() {
		event := sampleEvent()
		couple := event.Couples[0]
		handler := NewEventHandler(&event)
		d := &models.Dancer{
			Profile: &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe"},
			Role:    models.RoleFollower,
		}
		partner := handler.RegistrationGet(d).Partner
		suite.Require().NotNil(partner)

		got := handler.DancerRemove(d)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Nil(got.Partner)
		suite.Require().NotNil(got.Related)
		suite.Equal(partner, got.Related.Dancer)
		suite.Equal(models.StatusNotRegistered, got.Related.Status)
		suite.Equal(models.ResultRegistrationRemoved, got.Related.Result)
		suite.Require().Len(event.Couples, 1)
		suite.Require().Len(event.Singles, 2)
		suite.Require().Len(handler.hist, 1)
		suite.Equal(models.HistoryCoupleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(&couple, handler.hist[0].Details)
		suite.Require().Len(handler.notif, 1)
		suite.Equal(got.Dancer, handler.notif[0].Payload.Partner)
		suite.Equal(partner.Profile, handler.notif[0].Recipient)
		suite.Equal(models.TmplCanceledByPartner, handler.notif[0].TmplCode)
	})

	suite.Run("event is closed for all", func() {
		event := sampleEvent()
		event.Settings.ClosedFor = models.ClosedForAll
		dancer := event.Couples[0].Dancers[0]
		partner := event.Couples[0].Dancers[1]
		handler := NewEventHandler(&event)

		got := handler.DancerRemove(&dancer)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultEventClosed, got.Result)
		suite.Equal(models.StatusInCouple, got.Status)
		suite.Equal(&partner, got.Partner)
	})

	suite.Run("partner as single, event is closed for singles", func() {
		event := sampleEvent()
		event.Settings.ClosedFor = models.ClosedForSingles
		couple := event.Couples[0]
		dancer := couple.Dancers[0]
		partner := couple.Dancers[1]
		handler := NewEventHandler(&event)

		got := handler.DancerRemove(&dancer)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(&dancer, got.Dancer)
		suite.Nil(got.Partner)

		suite.Require().NotNil(got.Related)
		suite.Equal(models.StatusAsSingle, got.Related.Status)
		suite.Equal(models.ResultRegisteredAsSingle, got.Related.Result)
		suite.Equal(&partner, got.Related.Dancer)
		suite.Nil(got.Related.Partner)

		suite.Require().Len(event.Couples, 1)
		suite.Require().Len(event.Singles, 3)
		suite.Equal(partner, event.Singles[1])
	})
}

func (suite *TestEventHandlerSuite) TestDancerRemove_autoPair() {

	suite.Run("partner as single, found new partner", func() {
		config.SetBotProfile(botUser)
		event := sampleEvent()
		event.Settings.AutoPairing = true
		couple := event.Couples[1]
		dancer := couple.Dancers[1]
		partner := couple.Dancers[0]
		newPartner := event.Singles[0]
		handler := NewEventHandler(&event)

		got := handler.DancerRemove(&dancer)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(&dancer, got.Dancer)
		suite.Nil(got.Partner)

		suite.Require().NotNil(got.Related)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Equal(models.ResultRegisteredInCouple, got.Related.Result)
		suite.Equal(&partner, got.Related.Dancer)
		suite.Require().NotNil(got.Related.Partner)
		suite.Equal(&newPartner, got.Related.Partner)

		suite.Require().Len(event.Couples, 2)
		suite.Require().Len(event.Singles, 1)

		suite.Require().Len(handler.hist, 3)
		suite.Equal(models.HistoryCoupleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(&couple, handler.hist[0].Details)
		suite.Equal(models.HistorySingleRemoved, handler.hist[1].Action)
		suite.Equal(&botProfile, handler.hist[1].Initiator)
		suite.Equal(&newPartner, handler.hist[1].Details)
		suite.Equal(models.HistoryCoupleAdded, handler.hist[2].Action)
		suite.Equal(&botProfile, handler.hist[2].Initiator)
		suite.Equal(&event.Couples[1], handler.hist[2].Details)

		suite.Require().Len(handler.notif, 2)
		suite.Equal(models.TmplAutoPairPartnerFound, handler.notif[0].TmplCode)
		suite.Equal(newPartner.Profile, handler.notif[0].Recipient)
		suite.Equal(&partner, handler.notif[0].Payload.Partner)
		suite.Equal(models.TmplAutoPairPartnerChanged, handler.notif[1].TmplCode)
		suite.Equal(partner.Profile, handler.notif[1].Recipient)
		suite.Equal(&dancer, handler.notif[1].Payload.Partner)
		suite.Equal(&newPartner, handler.notif[1].Payload.NewPartner)
	})

	suite.Run("partner as single, no matching partner", func() {
		config.SetBotProfile(botUser)
		event := sampleEvent()
		event.Settings.AutoPairing = true
		couple := event.Couples[0]
		dancer := couple.Dancers[0]
		partner := couple.Dancers[1]
		handler := NewEventHandler(&event)

		got := handler.DancerRemove(&dancer)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(&dancer, got.Dancer)
		suite.Nil(got.Partner)

		suite.Require().NotNil(got.Related)
		suite.Equal(models.StatusAsSingle, got.Related.Status)
		suite.Equal(models.ResultRegisteredAsSingle, got.Related.Result)
		suite.Equal(&partner, got.Related.Dancer)
		suite.Nil(got.Related.Partner)

		suite.Require().Len(event.Couples, 1)
		suite.Require().Len(event.Singles, 3)
		suite.Equal(partner, event.Singles[1])

		suite.Require().Len(handler.hist, 2)
		suite.Equal(models.HistoryCoupleRemoved, handler.hist[0].Action)
		suite.Equal(got.Profile, handler.hist[0].Initiator)
		suite.Equal(&couple, handler.hist[0].Details)
		suite.Equal(models.HistorySingleAdded, handler.hist[1].Action)
		suite.Equal(got.Profile, handler.hist[1].Initiator)
		suite.Equal(&partner, handler.hist[1].Details)

		suite.Require().Len(handler.notif, 1)
		suite.Equal(models.TmplCanceledWithSingle, handler.notif[0].TmplCode)
		suite.Equal(partner.Profile, handler.notif[0].Recipient)
		suite.Equal(&dancer, handler.notif[0].Payload.Partner)
	})

	suite.Run("partner as single, event is closed for singles, found new partner", func() {
		config.SetBotProfile(botUser)
		event := sampleEvent()
		event.Settings.AutoPairing = true
		event.Settings.ClosedFor = models.ClosedForSingles
		couple := event.Couples[1]
		dancer := couple.Dancers[1]
		partner := couple.Dancers[0]
		newPartner := event.Singles[0]
		handler := NewEventHandler(&event)

		got := handler.DancerRemove(&dancer)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(&dancer, got.Dancer)
		suite.Nil(got.Partner)

		suite.Require().NotNil(got.Related)
		suite.Equal(models.StatusInCouple, got.Related.Status)
		suite.Equal(models.ResultRegisteredInCouple, got.Related.Result)
		suite.Equal(&partner, got.Related.Dancer)
		suite.Require().NotNil(got.Related.Partner)
		suite.Equal(&newPartner, got.Related.Partner)
	})

	suite.Run("partner as single, event is closed for singles, no matching partner", func() {
		config.SetBotProfile(botUser)
		event := sampleEvent()
		event.Settings.AutoPairing = true
		event.Settings.ClosedFor = models.ClosedForSingles
		couple := event.Couples[0]
		dancer := couple.Dancers[0]
		partner := couple.Dancers[1]
		handler := NewEventHandler(&event)

		got := handler.DancerRemove(&dancer)

		suite.Require().NotNil(got)
		suite.Equal(models.ResultRegistrationRemoved, got.Result)
		suite.Equal(models.StatusNotRegistered, got.Status)
		suite.Equal(&dancer, got.Dancer)
		suite.Nil(got.Partner)

		suite.Require().NotNil(got.Related)
		suite.Equal(models.StatusAsSingle, got.Related.Status)
		suite.Equal(models.ResultRegisteredAsSingle, got.Related.Result)
		suite.Equal(&partner, got.Related.Dancer)
		suite.Nil(got.Related.Partner)
	})
}

func sampleEvent() models.Event {
	return models.Event{
		ID:      "test12345678",
		Caption: "This is a test event",
		Post:    &models.Post{InlineMessageID: "123test456"},
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
						Profile:   &models.Profile{ID: 2, FirstName: "Jane", LastName: "Doe", Username: "janedoe"},
						FullName:  "Jane Doe",
						Role:      models.RoleFollower,
						AsSingle:  true,
						CreatedAt: nowFn().Add(-3 * time.Minute),
					},
				},
				CreatedBy: models.Profile{ID: 1, FirstName: "John", LastName: "Doe", Username: "johndoe"},
				CreatedAt: nowFn(),
			},
			{
				Dancers: []models.Dancer{
					{
						Profile:   &models.Profile{ID: 3, FirstName: "Jack", LastName: "Smith"},
						FullName:  "Jack Smith",
						Role:      models.RoleLeader,
						AsSingle:  true,
						CreatedAt: nowFn().Add(-2 * time.Minute),
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
				Profile:   &models.Profile{ID: 4, FirstName: "Kate", LastName: "Brown", Username: "katbrown"},
				FullName:  "Kate Brown",
				Role:      models.RoleFollower,
				AsSingle:  true,
				CreatedAt: nowFn().Add(-4 * time.Minute),
			},
			{
				Profile:   &models.Profile{ID: 5, FirstName: "Amalia", LastName: "Green"},
				FullName:  "Amalia Green",
				Role:      models.RoleFollower,
				AsSingle:  true,
				CreatedAt: nowFn().Add(-1 * time.Minute),
			},
		},
		Owner:     models.Profile{ID: 1000, FirstName: "Test", LastName: "Owner"},
		CreatedAt: nowFn(),
	}
}

var (
	botUser = &tele.User{
		ID:        1234567890,
		IsBot:     true,
		FirstName: "test_bot",
		Username:  "test_bot",
	}
	botProfile = models.NewProfile(*botUser)
)
