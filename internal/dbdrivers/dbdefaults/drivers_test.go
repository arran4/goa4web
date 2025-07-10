package dbdefaults_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/arran4/goa4web/internal/dbdrivers"
	dbdefaults "github.com/arran4/goa4web/internal/dbdrivers/dbdefaults"
)

type testConnector struct{}

func (testConnector) Connect(context.Context) (driver.Conn, error) { return nil, nil }
func (testConnector) Driver() driver.Driver                        { return nil }

func TestConnectorUnknown(t *testing.T) {
	dbdefaults.Register()
	if _, err := dbdrivers.Connector("unknown-driver", ""); err == nil {
		t.Fatalf("expected error for unknown driver")
	}
}

func TestRegistryNames(t *testing.T) {
	dbdefaults.Register()
	want := []string{"mysql", "postgres"}
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

type testDriver struct{}

func (testDriver) Name() string                                   { return "test" }
func (testDriver) Examples() []string                             { return nil }
func (testDriver) OpenConnector(string) (driver.Connector, error) { return testConnector{}, nil }

func TestConnectorRegistered(t *testing.T) {
	orig := dbdrivers.Registry
	t.Cleanup(func() { dbdrivers.Registry = orig })

	dbdrivers.RegisterDriver(testDriver{})
	c, err := dbdrivers.Connector("test", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(testConnector); !ok {
		t.Fatalf("unexpected connector type %T", c)
	}
}
