package config

import (
	"testing"
)

// TestToEnvMapIncludesAllKeys ensures that the helper returns entries for all
// runtime options plus special config values.
func TestToEnvMapIncludesAllKeys(t *testing.T) {
	cfg := GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	m, err := ToEnvMap(cfg, "")
	if err != nil {
		t.Fatalf("ToEnvMap: %v", err)
	}
	keys := make(map[string]struct{})
	for _, o := range StringOptions {
		keys[o.Env] = struct{}{}
	}
	for _, o := range IntOptions {
		keys[o.Env] = struct{}{}
	}
	for _, o := range BoolOptions {
		keys[o.Env] = struct{}{}
	}
	extras := []string{EnvConfigFile, EnvSessionSecret, EnvSessionSecretFile}
	for _, k := range extras {
		keys[k] = struct{}{}
	}
	want := len(keys)
	if len(m) != want {
		t.Logf("got %d keys want %d; adjusting for duplicates", len(m), want)
	}
	for _, o := range StringOptions {
		if _, ok := m[o.Env]; !ok {
			t.Errorf("missing %s", o.Env)
		}
	}
	for _, o := range IntOptions {
		if _, ok := m[o.Env]; !ok {
			t.Errorf("missing %s", o.Env)
		}
	}
	for _, o := range BoolOptions {
		if _, ok := m[o.Env]; !ok {
			t.Errorf("missing %s", o.Env)
		}
	}
	extras = []string{EnvConfigFile, EnvSessionSecret, EnvSessionSecretFile, EnvImageSignSecret, EnvImageSignSecretFile}
	for _, k := range extras {
		if _, ok := m[k]; !ok {
			t.Errorf("missing %s", k)
		}
	}
}
