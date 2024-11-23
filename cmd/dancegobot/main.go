package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ofstudio/dancegobot/internal/app"
	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/pkg/shutdown"
)

func main() {
	log := slog.Default()
	log.Info("starting...")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Error("fatal: failed to load config: " + err.Error())
		os.Exit(-1)
	}

	a := app.New(cfg).WithLogger(log)
	ctx, cancel := shutdown.Context(context.Background(), func(s os.Signal) {
		log.Warn("Received signal: " + s.String())
	})
	defer cancel()

	if err = a.Start(ctx); err != nil {
		log.Error("fatal: failed to start app: " + err.Error())
		os.Exit(-1)
	}

	log.Info("exiting...")
}
