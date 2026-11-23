package config_test

import (
	"flag"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/stretchr/testify/require"
)

func TestDurableJobHistoryConfiguration(t *testing.T) {
	defaults := config.NewRuntimeConfig(config.WithGetenv(func(string) string { return "" }))
	require.Equal(t, 24, defaults.DurableJobHistoryHours)
	require.Equal(t, 1000, defaults.DurableJobHistoryLimit)

	flags := flag.NewFlagSet("history", flag.ContinueOnError)
	flags.Int("durable-job-history-limit", 1000, "")
	require.NoError(t, flags.Parse([]string{"--durable-job-history-limit=50"}))

	env := map[string]string{
		config.EnvDurableJobHistoryHours: "12",
		config.EnvDurableJobHistoryLimit: "700",
	}
	files := map[string]string{
		config.EnvDurableJobHistoryHours: "6",
		config.EnvDurableJobHistoryLimit: "300",
	}
	cfg := config.NewRuntimeConfig(config.WithFlagSet(flags), config.WithFileValues(files), config.WithGetenv(func(key string) string { return env[key] }))
	require.Equal(t, 6, cfg.DurableJobHistoryHours)
	require.Equal(t, 50, cfg.DurableJobHistoryLimit)
}
