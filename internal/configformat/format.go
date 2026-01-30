package configformat

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/dbdrivers"
)

// AsOptions holds options used to render configuration output.
type AsOptions struct {
	Extended bool
}

func defaultEnvMap() map[string]string {
	def := config.NewRuntimeConfig()
	m, _ := config.ToEnvMap(def, "")
	return m
}

// FormatAsEnv renders configuration output as shell exports.
func FormatAsEnv(cfg *config.RuntimeConfig, configFile string, reg *dbdrivers.Registry, opts AsOptions) (string, error) {
	return formatEnv(cfg, configFile, reg, opts, true)
}

// FormatAsEnvFile renders configuration output as dotenv content.
func FormatAsEnvFile(cfg *config.RuntimeConfig, configFile string, reg *dbdrivers.Registry, opts AsOptions) (string, error) {
	return formatEnv(cfg, configFile, reg, opts, false)
}

// FormatAsJSON renders configuration output as JSON.
func FormatAsJSON(cfg *config.RuntimeConfig, configFile string) (string, error) {
	m, err := config.ToEnvMap(cfg, configFile)
	if err != nil {
		return "", fmt.Errorf("env map: %w", err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	return fmt.Sprintf("%s\n", string(b)), nil
}

// FormatAsCLI renders configuration output as CLI args.
func FormatAsCLI(cfg *config.RuntimeConfig, configFile string) (string, error) {
	current, err := config.ToEnvMap(cfg, configFile)
	if err != nil {
		return "", fmt.Errorf("env map: %w", err)
	}
	def := defaultEnvMap()
	var parts []string
	nameMap := config.NameMap()
	for env, val := range current {
		if def[env] == val {
			continue
		}
		n := nameMap[env]
		if n == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("--%s=%s", n, val))
	}
	sort.Strings(parts)
	return fmt.Sprintf("%s\n", strings.Join(parts, " ")), nil
}

func formatEnv(cfg *config.RuntimeConfig, configFile string, reg *dbdrivers.Registry, opts AsOptions, useExport bool) (string, error) {
	current, err := config.ToEnvMap(cfg, configFile)
	if err != nil {
		return "", fmt.Errorf("env map: %w", err)
	}
	def := defaultEnvMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := config.UsageMap()
	ext := map[string]string{}
	if opts.Extended {
		if reg == nil {
			reg = dbdrivers.NewRegistry()
		}
		ext = config.ExtendedUsageMap(reg)
	}
	ex := config.ExamplesMap()
	var b strings.Builder
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		line := "# "
		if u != "" {
			line += fmt.Sprintf("%s (default: %s)", u, d)
		} else {
			line += fmt.Sprintf("default: %s", d)
		}
		if xs := ex[k]; len(xs) > 0 {
			line += fmt.Sprintf(" (examples: %s)", strings.Join(xs, ", "))
		}
		fmt.Fprintln(&b, line)
		if opts.Extended {
			if e := ext[k]; e != "" {
				for _, line := range strings.Split(strings.TrimSuffix(e, "\n"), "\n") {
					fmt.Fprintf(&b, "# %s\n", line)
				}
			}
		}
		if useExport {
			fmt.Fprintf(&b, "export %s=%s\n", k, current[k])
		} else {
			fmt.Fprintf(&b, "%s=%s\n", k, current[k])
		}
	}
	return b.String(), nil
}
