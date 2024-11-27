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
	// 1. Create logger
	log := slog.Default()
	log.Info("Starting", "version", config.Version())

	// 2. Load the configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Fatal: failed to load config: " + err.Error())
		os.Exit(-1)
	}

	// 3. Create application context
	ctx, cancel := shutdown.Context(context.Background(), func(s os.Signal) {
		log.Warn("Received signal: " + s.String())
	})
	defer cancel()

	// 4. Start the application
	if err = app.New(cfg).WithLogger(log).Start(ctx); err != nil {
		log.Error("Fatal: application error: " + err.Error())
		os.Exit(-1)
	}

	log.Info("Exiting...")
}
