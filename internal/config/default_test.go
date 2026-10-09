package config

import (
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/require"
)

func TestDefaultRendererInitialDelay(t *testing.T) {
	require.Equal(t, 5*time.Second, Default().RendererInitialDelay)
}

func TestRendererInitialDelayIgnoresEnvironment(t *testing.T) {
	t.Setenv("RENDERER_INITIAL_DELAY", "1h")
	t.Setenv("RendererInitialDelay", "1h")
	cfg := Default().Settings
	require.NoError(t, env.Parse(&cfg))
	require.Equal(t, 5*time.Second, cfg.RendererInitialDelay)
}

func TestDefaultNotifierRepeats(t *testing.T) {
	require.Equal(t, []time.Duration{
		time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		time.Hour,
		2 * time.Hour,
		6 * time.Hour,
		12 * time.Hour,
	}, Default().NotifierRepeats)
}
