package configformat_test

import (
	"database/sql/driver"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/configformat"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDriver struct{}

func (m *mockDriver) Name() string { return "mock" }

func (m *mockDriver) Examples() []string {
	return []string{"example1", "example2"}
}

func (m *mockDriver) OpenConnector(dsn string) (driver.Connector, error) {
	return nil, nil
}

func (m *mockDriver) Backup(dsn, file string) error {
	return nil
}

func (m *mockDriver) Restore(dsn, file string) error {
	return nil
}

func TestFormatAsEnv(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	reg := dbdrivers.NewRegistry()
	reg.RegisterDriver(&mockDriver{})

	t.Run("Standard", func(t *testing.T) {
		out, err := configformat.FormatAsEnv(cfg, "", reg, configformat.AsOptions{})
		require.NoError(t, err)
		assert.Contains(t, out, "export DB_CONN=")
		assert.Contains(t, out, "# Database connection string")
		// Should not contain extended info
		assert.NotContains(t, out, "mock examples:")
	})

	t.Run("Extended", func(t *testing.T) {
		out, err := configformat.FormatAsEnv(cfg, "", reg, configformat.AsOptions{Extended: true})
		require.NoError(t, err)
		assert.Contains(t, out, "export DB_CONN=")
		// Should contain extended info from mock driver
		assert.Contains(t, out, "mock examples:")
		assert.Contains(t, out, "- example1")
	})
}

func TestFormatAsEnvFile(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	reg := dbdrivers.NewRegistry()

	t.Run("Standard", func(t *testing.T) {
		out, err := configformat.FormatAsEnvFile(cfg, "", reg, configformat.AsOptions{})
		require.NoError(t, err)
		// Should have DB_CONN= but not export DB_CONN=
		assert.Contains(t, out, "\nDB_CONN=")
		assert.NotContains(t, out, "export DB_CONN=")
	})
}

func TestFormatAsJSON(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.DBConn = "test-conn"

	out, err := configformat.FormatAsJSON(cfg, "")
	require.NoError(t, err)
	assert.Contains(t, out, "\"DB_CONN\": \"test-conn\"")
}

func TestFormatAsCLI(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	// Set a non-default value to verify it shows up
	cfg.DBConn = "test-conn"
	// Set another one
	cfg.DBDriver = "test-driver"

	out, err := configformat.FormatAsCLI(cfg, "")
	require.NoError(t, err)
	assert.Contains(t, out, "--db-conn=test-conn")
	assert.Contains(t, out, "--db-driver=test-driver")
	// Verify sorting (db-conn comes before db-driver)
	// But db-conn is alphabetically before db-driver anyway.
	// Let's rely on Contains for now.
}
