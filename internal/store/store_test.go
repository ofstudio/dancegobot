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

func TestSQLiteTimeBoundary(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{name: "whole second UTC", time: base, want: "2026-10-01 00:00:00"},
		{name: "positive offset", time: base.In(time.FixedZone("UTC+03", 3*60*60)), want: "2026-10-01 00:00:00"},
		{name: "negative offset across date", time: base.In(time.FixedZone("UTC-05:30", -(5*60+30)*60)), want: "2026-10-01 00:00:00"},
		{name: "fractional second", time: base.Add(123456789 * time.Nanosecond), want: "2026-10-01 00:00:00.123456789"},
		{name: "fractional second without trailing zeros", time: base.Add(500 * time.Millisecond), want: "2026-10-01 00:00:00.5"},
		{name: "nanosecond precision", time: base.Add(time.Nanosecond), want: "2026-10-01 00:00:00.000000001"},
		{name: "previous UTC date", time: base.Add(-time.Nanosecond).In(time.FixedZone("UTC+03", 3*60*60)), want: "2026-09-30 23:59:59.999999999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqliteTimeBoundary(tt.time); got != tt.want {
				t.Errorf("sqliteTimeBoundary(%v) = %q, want %q", tt.time, got, tt.want)
			}
		})
	}
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
