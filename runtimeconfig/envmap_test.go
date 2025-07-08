package runtimeconfig

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

// TestToEnvMapIncludesAllKeys ensures that the helper returns entries for all
// runtime options plus special config values.
func TestToEnvMapIncludesAllKeys(t *testing.T) {
	cfg := GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	m, err := ToEnvMap(cfg, "")
	if err != nil {
		t.Fatalf("ToEnvMap: %v", err)
	}
	want := len(StringOptions) + len(IntOptions) + len(BoolOptions) + 3
	if len(m) != want {
		t.Fatalf("got %d keys want %d", len(m), want)
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
	extras := []string{config.EnvConfigFile, config.EnvSessionSecret, config.EnvSessionSecretFile}
	for _, k := range extras {
		if _, ok := m[k]; !ok {
			t.Errorf("missing %s", k)
		}
	}
}
