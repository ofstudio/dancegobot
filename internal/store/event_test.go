package store

import (
	"context"
	"time"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (suite *TestStoreSuite) TestEventUpsert() {
	suite.Run("insert", func() {
		event := &models.Event{
			ID:      "abc",
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
			ID:      "abc",
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
VALUES ('abc', 1, '{"id": "abc", "owner": {"id": 1}}')
`)
		suite.Require().NoError(err)

		event, err := suite.store.EventGet(context.Background(), "abc")
		suite.Require().NoError(err)
		suite.Equal(&models.Event{
			ID: "abc",
			Owner: models.Profile{
				ID: 1,
			},
		}, event)
	})

	suite.Run("not found", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data)
VALUES ('abc', 1, '{"id": "abc", "owner": {"id": 1}}')
`)
		suite.Require().NoError(err)

		event, err := suite.store.EventGet(context.Background(), "def")
		suite.ErrorIs(err, ErrNotFound)
		suite.Nil(event)
	})
}

func (suite *TestStoreSuite) TestEventGetUpdatedAfter() {
	suite.Run("success", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data, updated_at)
VALUES ('abc', 1, '{"id": "abc", "post": {"inline_message_id": "qwe"} }', '2021-01-01 00:00:00'),
       ('def', 1, '{"id": "def", "post": {"inline_message_id": "qwe"} }', '2021-01-02 00:00:01'),
       ('ghi', 1, '{"id": "ghi"}', '2021-01-03 00:00:00')
`)
		suite.Require().NoError(err)

		events, err := suite.store.EventGetUpdatedAfter(
			context.Background(),
			time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		)
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Equal("def", events[0].ID)
	})
}

func (suite *TestStoreSuite) TestEventRemoveDraftsBefore() {
	suite.Run("success", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO events (id, owner_id, data, updated_at)
VALUES ('abc', 1, '{"id": "abc", "post": {"inline_message_id": "qwe"} }', '2021-01-01 00:00:00'), -- This should NOT be removed
       ('xxx', 1, '{"id": "def", "couples": [1,2,3] }', '2021-01-02 00:00:01'),                   -- This should NOT be removed
       ('yyy', 1, '{"id": "def", "singles": [5,6,7] }', '2021-01-02 00:00:01'),                   -- This should NOT be removed
       ('def', 1, '{"id": "def" }', '2021-01-02 00:00:01'),                                       -- This should BE removed
       ('ghi', 1, '{"id": "ghi"}', '2021-01-03 00:00:00'),                                        -- This should BE removed
       ('jkl', 1, '{"id": "jkl", "post": {"inline_message_id": "zxc"} }', '2021-01-04 00:00:00'), -- This should NOT be removed
       ('mno', 1, '{"id": "mno" }', '2021-01-05 00:00:01'),                                       -- This should NOT be removed
       ('pqr', 1, '{"id": "pqr", "post": {"inline_message_id": "rty"} }', '2021-01-06 00:00:00')  -- This should NOT be removed
`)
		suite.Require().NoError(err)

		ids, err := suite.store.EventRemoveDraftsBefore(
			context.Background(),
			time.Date(2021, 1, 5, 0, 0, 0, 0, time.UTC),
		)
		suite.Require().NoError(err)
		suite.Require().Len(ids, 2)
		suite.Contains(ids, "def")
		suite.Contains(ids, "ghi")

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
		suite.Contains(idsFromDB, "abc")
		suite.Contains(idsFromDB, "jkl")
		suite.Contains(idsFromDB, "mno")
		suite.Contains(idsFromDB, "pqr")
		suite.Contains(idsFromDB, "xxx")
		suite.Contains(idsFromDB, "yyy")
	})
}
