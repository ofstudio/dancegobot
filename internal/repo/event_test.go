package repo

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

		events, err := suite.store.EventGetUpdatedAfter(context.Background(), time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC))
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Equal("def", events[0].ID)
	})
}
