package stats

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/dbdrivers"
)

func TestBuildServerStatsData_NilDBRegistry(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	data := BuildServerStatsData(cfg, "", nil, nil, nil, nil, []string{"testmodule"})
	if data.Registries.RouterModules[0] != "testmodule" {
		t.Errorf("expected router module 'testmodule', got %v", data.Registries.RouterModules)
	}
}

func TestBuildServerStatsData_WithDBRegistry(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	reg := dbdrivers.NewRegistry()
	data := BuildServerStatsData(cfg, "", nil, reg, nil, nil, []string{"testmodule"})
	if data.Registries.RouterModules[0] != "testmodule" {
		t.Errorf("expected router module 'testmodule', got %v", data.Registries.RouterModules)
	}
}
