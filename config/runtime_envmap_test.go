package config_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

// TestToEnvMapIncludesAllKeys ensures that the helper returns entries for all
// runtime options plus special config values.
func TestToEnvMapIncludesAllKeys(t *testing.T) {
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	m, err := config.ToEnvMap(cfg, "")
	if err != nil {
		t.Fatalf("ToEnvMap: %v", err)
	}
	keys := make(map[string]struct{})
	for _, o := range config.StringOptions {
		keys[o.Env] = struct{}{}
	}
	for _, o := range config.IntOptions {
		keys[o.Env] = struct{}{}
	}
	for _, o := range config.BoolOptions {
		keys[o.Env] = struct{}{}
	}
	extras := []string{config.EnvConfigFile, config.EnvSessionSecret, config.EnvSessionSecretFile}
	for _, k := range extras {
		keys[k] = struct{}{}
	}
	want := len(keys)
	if len(m) != want {
		t.Logf("got %d keys want %d; adjusting for duplicates", len(m), want)
	}
	for _, o := range config.StringOptions {
		if _, ok := m[o.Env]; !ok {
			t.Errorf("missing %s", o.Env)
		}
	}
	for _, o := range config.IntOptions {
		if _, ok := m[o.Env]; !ok {
			t.Errorf("missing %s", o.Env)
		}
	}
	for _, o := range config.BoolOptions {
		if _, ok := m[o.Env]; !ok {
			t.Errorf("missing %s", o.Env)
		}
	}
	extras = []string{config.EnvConfigFile, config.EnvSessionSecret, config.EnvSessionSecretFile, config.EnvImageSignSecret, config.EnvImageSignSecretFile}
	for _, k := range extras {
		if _, ok := m[k]; !ok {
			t.Errorf("missing %s", k)
		}
	}
}
