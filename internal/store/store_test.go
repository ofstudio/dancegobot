package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestStore(t *testing.T) {
	suite.Run(t, new(TestStoreSuite))
}

type TestStoreSuite struct {
	suite.Suite
	store *SQLiteStore
}

func (suite *TestStoreSuite) SetupSubTest() {
	db, err := NewSQLite(":memory:", config.Default().DB.Version)
	suite.Require().NoError(err)
	suite.store = NewSQLiteStore(db)
}

func (suite *TestStoreSuite) TearDownSubTest() {
	suite.store.Close()
	suite.store = nil
}

func (suite *TestStoreSuite) TestStoreTx() {
	suite.Run("tx and non-tx requests", func() {

		go func() {
			tx, err := suite.store.Begin(context.Background())
			suite.Require().NoError(err)

			time.Sleep(200 * time.Millisecond)

			suite.Require().
				NoError(tx.UserUpsertProfile(context.Background(), &models.User{Profile: models.Profile{ID: 1}}))
			suite.Require().NoError(tx.Commit())
		}()

		time.Sleep(100 * time.Millisecond)
		suite.Require().
			NoError(suite.store.UserUpsertProfile(context.Background(), &models.User{Profile: models.Profile{ID: 2}}))
	})
}
