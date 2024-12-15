package repo

import (
	"context"
	"encoding/json"

	"github.com/ofstudio/dancegobot/internal/models"
)

func (suite *TestStoreSuite) TestStoreHistoryInsert() {
	suite.Run("success", func() {
		eventID := "abc"
		item := &models.HistoryItem{
			Initiator: &models.Profile{
				ID:        123,
				FirstName: "Test",
			},
			EventID: &eventID,
			Action:  models.HistorySingleAdded,
			Details: `{"foo": "bar"}`,
		}

		err := suite.store.HistoryInsert(context.Background(), item)
		suite.Require().NoError(err)

		rows, err := suite.store.db.Query(`SELECT initiator_id, event_id, data
FROM history
`)
		suite.Require().NoError(err)
		//goland:noinspection ALL
		defer rows.Close()

		var profileID int64
		var gotEventID *string
		var data string
		suite.True(rows.Next())
		err = rows.Scan(&profileID, &gotEventID, &data)
		suite.Require().NoError(err)
		suite.Equal(item.Initiator.ID, profileID)
		suite.Equal(item.EventID, gotEventID)

		got := &models.HistoryItem{}
		err = json.Unmarshal([]byte(data), got)
		suite.Require().NoError(err)
		suite.Equal(item, got)
	})

}

func (suite *TestStoreSuite) TestStoreHistoryRemoveByEventIDs() {
	suite.Run("success", func() {
		_, err := suite.store.db.Exec(`
INSERT INTO history (action, initiator_id, event_id, data, created_at)
VALUES ('added', 1, 'abc', '{"foo": "bar"}', '2021-01-01 00:00:00'),
       ('added', 2, 'abc', '{"foo": "bar"}', '2021-01-02 00:00:00'),
       ('added', 3, 'def', '{"foo": "bar"}', '2021-01-03 00:00:00'),
       ('added', 4, 'def', '{"foo": "bar"}', '2021-01-04 00:00:00'),
       ('added', 5, 'ghi', '{"foo": "bar"}', '2021-01-05 00:00:00')
`)
		suite.Require().NoError(err)

		count, err := suite.store.HistoryRemoveByEventIDs(
			context.Background(),
			[]string{"abc", "def"},
		)
		suite.Require().NoError(err)
		suite.Equal(4, count)

		rows, err := suite.store.db.Query(`SELECT event_id FROM history`)
		suite.Require().NoError(err)
		//goland:noinspection ALL
		defer rows.Close()

		var eventID string
		var eventIDs []string
		for rows.Next() {
			err = rows.Scan(&eventID)
			suite.Require().NoError(err)
			eventIDs = append(eventIDs, eventID)
		}

		suite.Require().Len(eventIDs, 1)
		suite.Contains(eventIDs, "ghi")
	})

	suite.Run("empty", func() {
		count, err := suite.store.HistoryRemoveByEventIDs(
			context.Background(),
			[]string{},
		)
		suite.Require().NoError(err)
		suite.Equal(0, count)
	})
}
