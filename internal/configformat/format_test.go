package configformat

import (
	"encoding/json"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatAsEnv(t *testing.T) {
	cfg := config.NewRuntimeConfig(
		config.WithGetenv(func(s string) string { return "" }),
	)
	// Set specific value
	cfg.HTTPListen = ":9090"

	reg := dbdrivers.NewRegistry()

	t.Run("Basic", func(t *testing.T) {
		out, err := FormatAsEnv(cfg, "", reg, AsOptions{Extended: false})
		require.NoError(t, err)
		assert.Contains(t, out, "export LISTEN=:9090")
		// Check for default comment format
		assert.Contains(t, out, "# The address and port for the HTTP server to listen on. (default: :8080)")
	})

	t.Run("Extended", func(t *testing.T) {
		out, err := FormatAsEnv(cfg, "", reg, AsOptions{Extended: true})
		require.NoError(t, err)
		assert.Contains(t, out, "export LISTEN=:9090")
		// Extended usage might not be present for LISTEN, but let's check basic presence
		assert.Contains(t, out, "# The address and port for the HTTP server to listen on. (default: :8080)")
	})
}

func TestFormatAsEnvFile(t *testing.T) {
	cfg := config.NewRuntimeConfig(
		config.WithGetenv(func(s string) string { return "" }),
	)
	cfg.HTTPListen = ":9090"

	reg := dbdrivers.NewRegistry()

	t.Run("Basic", func(t *testing.T) {
		out, err := FormatAsEnvFile(cfg, "", reg, AsOptions{Extended: false})
		require.NoError(t, err)
		assert.Contains(t, out, "LISTEN=:9090")
		assert.NotContains(t, out, "export LISTEN=:9090")
		// Basic usage comments should be there
		assert.Contains(t, out, "# The address and port for the HTTP server to listen on. (default: :8080)")
	})

	t.Run("Extended", func(t *testing.T) {
		out, err := FormatAsEnvFile(cfg, "", reg, AsOptions{Extended: true})
		require.NoError(t, err)
		assert.Contains(t, out, "LISTEN=:9090")
		assert.NotContains(t, out, "export LISTEN=:9090")
		assert.Contains(t, out, "# The address and port for the HTTP server to listen on. (default: :8080)")
	})
}

func TestFormatAsJSON(t *testing.T) {
	cfg := config.NewRuntimeConfig(
		config.WithGetenv(func(s string) string { return "" }),
	)
	cfg.HTTPListen = ":9090"

	out, err := FormatAsJSON(cfg, "")
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal([]byte(out), &m)
	require.NoError(t, err)
	assert.Equal(t, ":9090", m["LISTEN"])
}

func TestFormatAsCLI(t *testing.T) {
	cfg := config.NewRuntimeConfig(
		config.WithGetenv(func(s string) string { return "" }),
	)
	cfg.HTTPListen = ":9090"
	// Defaults usually filtered out? Let's check logic in FormatAsCLI
	// "if def[env] == val { continue }"
	// Default for HTTPListen is ":8080". So ":9090" should appear.

	out, err := FormatAsCLI(cfg, "")
	require.NoError(t, err)
	assert.Contains(t, out, "--listen=:9090")
}
