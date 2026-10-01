package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
