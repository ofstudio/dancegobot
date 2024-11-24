package store

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
