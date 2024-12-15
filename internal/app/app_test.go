package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

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

func (suite *AppTestSuite) SetupSubTest() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	cfg := config.Default()
	cfg.Bot.Token = "123456:ABCDEF"
	cfg.DB.Filepath = suite.T().TempDir() + "/app_test.db"

	// disable all the background tasks
	cfg.RendererRepeats = []time.Duration{}
	cfg.ReRenderOnStartup = 0
	cfg.DraftCleanupEvery = 0
	cfg.DraftCleanupOlderThan = 0

	gock.New(telegock.GetMe).
		Reply(200).
		JSON(telegock.Result(botUser))

	gock.New(telegock.SetMyCommands).
		Reply(200).JSON(telegock.Result(true))

	suite.app = New(cfg).WithLogger(slog.Default())
	go func() {
		suite.Require().NoError(suite.app.Start(suite.ctx))
	}()

	suite.NoPending()
	suite.NoUnmatched()
}

func (suite *AppTestSuite) TearDownSubTest() {
	suite.cancel()
	// wait for the app to stop
	time.Sleep(100 * time.Millisecond)
	gock.Off()
}
