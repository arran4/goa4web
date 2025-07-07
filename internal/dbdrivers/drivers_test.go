package dbdrivers_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dbdrivers/allstable"
)

type testConnector struct{}

func (testConnector) Connect(context.Context) (driver.Conn, error) { return nil, nil }
func (testConnector) Driver() driver.Driver                        { return nil }

func TestConnectorUnknown(t *testing.T) {
	if _, err := dbdrivers.Connector("unknown-driver", ""); err == nil {
		t.Fatalf("expected error for unknown driver")
	}
}

func TestRegisterConnector(t *testing.T) {
	dbdrivers.RegisterConnector("testdriver", func(dsn string) (driver.Connector, error) {
		if dsn != "dsn" {
			return nil, fmt.Errorf("unexpected dsn: %s", dsn)
		}
		return testConnector{}, nil
	})
	c, err := dbdrivers.Connector("testdriver", "dsn")
	if err != nil {
		t.Fatalf("Connector returned error: %v", err)
	}
	if c == nil {
		t.Fatalf("expected connector")
	}
}

func TestAllstableRegister(t *testing.T) {
	allstable.Register()
	for _, n := range []string{"mysql", "postgres", "sqlite3"} {
		_, err := dbdrivers.Connector(n, "")
		if err != nil && strings.Contains(err.Error(), "unsupported driver") {
			t.Fatalf("%s not registered", n)
		}
	}
}
