package app

import (
	"context"
	"log/slog"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/pkg/telegock"
)

func TestApp(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

type AppTestSuite struct {
	telegock.Suite
	app    *App
	ctx    context.Context
	cancel context.CancelFunc
}

func (suite *AppTestSuite) SetupTest() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	cfg := config.Default()
	cfg.Bot.Token = "123456:ABCDEF"
	cfg.DB.Filepath = suite.T().TempDir() + "/app_test.db"

	suite.app = New(cfg).WithLogger(slog.Default())
	go func() {
		suite.NoError(suite.app.Start(suite.ctx))
	}()

	gock.New(telegock.GetMe).
		Reply(200).
		JSON(telegock.Result(&botUser))
	gock.New(telegock.SetMyCommands).
		Reply(200)
	suite.NoPending()
	suite.NoUnmatched()
}

func (suite *AppTestSuite) TearDownTest() {
	suite.cancel()
	gock.Off()
}
