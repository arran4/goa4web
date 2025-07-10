package main

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestToEnvMapLoops(t *testing.T) {
	cfg := config.RuntimeConfig{
		EmailEnabled:         false,
		NotificationsEnabled: false,
		CSRFEnabled:          true,
		AdminNotify:          true,
		FeedsEnabled:         false,
		EmailSMTPStartTLS:    false,
		StatsStartYear:       2020,
	}

	m, err := config.ToEnvMap(cfg, "")
	if err != nil {
		t.Fatalf("ToEnvMap: %v", err)
	}
	tests := map[string]string{
		config.EnvEmailEnabled:         "false",
		config.EnvNotificationsEnabled: "false",
		config.EnvCSRFEnabled:          "true",
		config.EnvAdminNotify:          "true",
		config.EnvFeedsEnabled:         "false",
		config.EnvSMTPStartTLS:         "false",
		config.EnvStatsStartYear:       "2020",
	}
	for k, want := range tests {
		if m[k] != want {
			t.Errorf("%s=%s, want %s", k, m[k], want)
		}
	}
}
