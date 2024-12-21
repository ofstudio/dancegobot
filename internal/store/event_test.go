package store

import (
	"context"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (suite *TestStoreSuite) TestEventUpsert() {
	suite.Run("insert", func() {
		event := &models.Event{
			ID:      "event_1",
			Caption: "test",
			Owner: models.Profile{
				ID:        12,
				FirstName: "Test",
			},
			CreatedAt: time.Now().Truncate(time.Millisecond).UTC(),
		}
		err := suite.store.EventUpsert(context.Background(), event)
		suite.Require().NoError(err)

		got, err := suite.store.EventGet(context.Background(), event.ID)
		suite.Require().NoError(err)
		suite.Equal(event, got)
	})

	suite.Run("update", func() {
		event := &models.Event{
			ID:      "event_1",
			Caption: "test",
			Owner: models.Profile{
				ID:        12,
				FirstName: "Test",
			},
			CreatedAt: time.Now().Truncate(time.Millisecond).UTC(),
		}
		err := suite.store.EventUpsert(context.Background(), event)
		suite.Require().NoError(err)

		event.Caption = "new text"
		err = suite.store.EventUpsert(context.Background(), event)
		suite.Require().NoError(err)

		got, err := suite.store.EventGet(context.Background(), event.ID)
		suite.Require().NoError(err)
		suite.Equal(event, got)
	})
}

func (suite *TestStoreSuite) TestEventGet() {
	suite.Run("success", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data)
VALUES ('event_1', 1, '{"id": "event_1", "owner": {"id": 1}}')
`)
		suite.Require().NoError(err)

		event, err := suite.store.EventGet(context.Background(), "event_1")
		suite.Require().NoError(err)
		suite.Equal(&models.Event{
			ID: "event_1",
			Owner: models.Profile{
				ID: 1,
			},
		}, event)
	})

	suite.Run("not found", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data)
VALUES ('event_1', 1, '{"id": "event_1", "owner": {"id": 1}}')
`)
		suite.Require().NoError(err)

		event, err := suite.store.EventGet(context.Background(), "event_2")
		suite.ErrorIs(err, ErrNotFound)
		suite.Nil(event)
	})
}

func (suite *TestStoreSuite) TestEventGetUpdatedAfter() {
	suite.Run("success", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data, updated_at)
VALUES ('event_1', 1, '{"id": "event_1", "post": {"inline_message_id": "qwe"} }', '2021-01-01 00:00:00'),
       ('event_2', 1, '{"id": "event_2", "post": {"inline_message_id": "qwe"} }', '2021-01-02 00:00:01'),
       ('event_3', 1, '{"id": "event_3", }', '2021-01-03 00:00:00')
`)
		suite.Require().NoError(err)

		events, err := suite.store.EventGetUpdatedAfter(
			context.Background(),
			time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		)
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Equal("event_2", events[0].ID)
	})
}

func (suite *TestStoreSuite) TestEventRemoveDraftsBefore() {
	suite.Run("success", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data, updated_at)
VALUES ('event_1', 1, '{"post": {"inline_message_id": "qwe"} }', '2021-01-01 00:00:00'), -- This should NOT be removed
       ('event_2', 1, '{"couples": [1,2,3] }', '2021-01-02 00:00:01'),                   -- This should NOT be removed
       ('event_3', 1, '{"singles": [5,6,7] }', '2021-01-02 00:00:01'),                   -- This should NOT be removed
       ('event_4', 1, '{}', '2021-01-02 00:00:01'),                                       -- This should BE removed
       ('event_5', 1, '{}', '2021-01-03 00:00:00'),                                        -- This should BE removed
       ('event_6', 1, '{"post": {"inline_message_id": "zxc"} }', '2021-01-04 00:00:00'), -- This should NOT be removed
       ('event_7', 1, '{}', '2021-01-05 00:00:01'),                                       -- This should NOT be removed
       ('event_8', 1, '{"post": {"inline_message_id": "rty"} }', '2021-01-06 00:00:00')  -- This should NOT be removed
`)
		suite.Require().NoError(err)

		ids, err := suite.store.EventRemoveDraftsBefore(
			context.Background(),
			time.Date(2021, 1, 5, 0, 0, 0, 0, time.UTC),
		)
		suite.Require().NoError(err)
		suite.Require().Len(ids, 2)
		suite.Contains(ids, "event_4")
		suite.Contains(ids, "event_5")

		res, err := suite.store.db.Query("SELECT id FROM events")
		suite.Require().NoError(err)
		//goland:noinspection ALL
		defer res.Close()
		var id string
		var idsFromDB []string
		for res.Next() {
			err = res.Scan(&id)
			suite.Require().NoError(err)
			idsFromDB = append(idsFromDB, id)
		}
		suite.Require().Len(idsFromDB, 6)
		suite.Contains(idsFromDB, "event_1")
		suite.Contains(idsFromDB, "event_2")
		suite.Contains(idsFromDB, "event_3")
		suite.Contains(idsFromDB, "event_6")
		suite.Contains(idsFromDB, "event_7")
		suite.Contains(idsFromDB, "event_8")
	})
}
