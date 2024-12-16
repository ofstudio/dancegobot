package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

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
	db, err := NewSQLite(":memory:", 1)
	suite.Require().NoError(err)
	suite.store = NewSQLiteStore(db)
}

func (suite *TestStoreSuite) TearDownSubTest() {
	suite.store.Close()
	suite.store = nil
}

func (suite *TestStoreSuite) TestStoreTx() {
	suite.Run("tx and non-tx requests", func() {
		time.Sleep(300 * time.Millisecond)
		db, err := NewSQLite(":memory:", 1)
		suite.Require().NoError(err)

		store := NewSQLiteStore(db)
		defer store.Close()

		go func() {
			tx, err := store.BeginTx(context.Background())
			suite.Require().NoError(err)

			time.Sleep(700 * time.Millisecond)
			err = tx.UserUpsert(context.Background(), &models.User{})
			suite.Require().NoError(err)
			suite.Require().NoError(tx.Commit())
		}()

		time.Sleep(300 * time.Millisecond)
		err = store.UserUpsert(context.Background(), &models.User{})
		suite.Require().NoError(err)

	})
}
