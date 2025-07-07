package dbdrivers_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/arran4/goa4web/internal/dbdrivers"
)

type testConnector struct{}

func (testConnector) Connect(context.Context) (driver.Conn, error) { return nil, nil }
func (testConnector) Driver() driver.Driver                        { return nil }

func TestConnectorUnknown(t *testing.T) {
	if _, err := dbdrivers.Connector("unknown-driver", ""); err == nil {
		t.Fatalf("expected error for unknown driver")
	}
}

func TestRegistryNames(t *testing.T) {
	want := []string{"mysql", "postgres", "sqlite3"}
	names := dbdrivers.Names()
	for _, n := range want {
		found := false
		for _, rn := range names {
			if rn == n {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s not listed in registry", n)
		}
	}
}
