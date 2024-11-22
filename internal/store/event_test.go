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
			CreatedAt: time.Now().Truncate(time.Millisecond),
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
			CreatedAt: time.Now().Truncate(time.Millisecond),
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
