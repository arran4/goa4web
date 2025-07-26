package config

import (
	"strconv"

	"github.com/arran4/goa4web/internal/dbdrivers"
)

// DefaultMap returns a map of environment variable names to their
// built-in default values.
func DefaultMap(cfg *RuntimeConfig) map[string]string {
	m := make(map[string]string)
	for _, o := range StringOptions {
		m[o.Env] = *o.Target(cfg)
	}
	for _, o := range IntOptions {
		m[o.Env] = strconv.Itoa(*o.Target(cfg))
	}
	for _, o := range BoolOptions {
		m[o.Env] = strconv.FormatBool(*o.Target(cfg))
	}
	return m
}

// UsageMap returns a map of environment variable names to their short usage text.
// usageMap returns a usage map built from the provided option sets. When all
// option slices are nil the built-in sets are used.
func usageMap(sopts []StringOption, iopts []IntOption, bopts []BoolOption) map[string]string {
	if sopts == nil {
		sopts = StringOptions
	} else {
		sopts = append(append([]StringOption(nil), StringOptions...), sopts...)
	}
	if iopts == nil {
		iopts = IntOptions
	} else {
		iopts = append(append([]IntOption(nil), IntOptions...), iopts...)
	}
	if bopts == nil {
		bopts = BoolOptions
	} else {
		bopts = append(append([]BoolOption(nil), BoolOptions...), bopts...)
	}
	m := make(map[string]string)
	for _, o := range sopts {
		m[o.Env] = o.Usage
	}
	for _, o := range iopts {
		m[o.Env] = o.Usage
	}
	for _, o := range bopts {
		m[o.Env] = o.Usage
	}
	return m
}

// UsageMap returns a map of environment variable names to their short usage
// text for the built-in runtime options.
func UsageMap() map[string]string { return usageMap(nil, nil, nil) }

// UsageMapWithOptions is like UsageMap but also includes the supplied option
// slices in the returned map.
func UsageMapWithOptions(sopts []StringOption, iopts []IntOption, bopts []BoolOption) map[string]string {
	return usageMap(sopts, iopts, bopts)
}

// nameMap builds a map of environment variable names to their CLI flag names
// from the provided option sets. Nil slices default to the built-in options.
func nameMap(sopts []StringOption, iopts []IntOption, bopts []BoolOption) map[string]string {
	if sopts == nil {
		sopts = StringOptions
	} else {
		sopts = append(append([]StringOption(nil), StringOptions...), sopts...)
	}
	if iopts == nil {
		iopts = IntOptions
	} else {
		iopts = append(append([]IntOption(nil), IntOptions...), iopts...)
	}
	if bopts == nil {
		bopts = BoolOptions
	} else {
		bopts = append(append([]BoolOption(nil), BoolOptions...), bopts...)
	}
	m := make(map[string]string)
	for _, o := range sopts {
		m[o.Env] = o.Name
	}
	for _, o := range iopts {
		m[o.Env] = o.Name
	}
	for _, o := range bopts {
		if o.Name != "" {
			m[o.Env] = o.Name
		}
	}
	return m
}

// NameMap returns a map of environment variable names to their CLI flag names
// for the built-in runtime options.
func NameMap() map[string]string { return nameMap(nil, nil, nil) }

// NameMapWithOptions is like NameMap but also merges the supplied options into
// the returned map.
func NameMapWithOptions(sopts []StringOption, iopts []IntOption, bopts []BoolOption) map[string]string {
	return nameMap(sopts, iopts, bopts)
}

// ExtendedUsageMap returns extended usage text indexed by environment variable name.
// Errors while rendering the usage templates are ignored.
func ExtendedUsageMap(reg *dbdrivers.Registry) map[string]string {
	m := make(map[string]string)
	for _, o := range StringOptions {
		if o.ExtendedUsage != "" {
			if txt, err := ExtendedUsage(o.ExtendedUsage, reg); err == nil {
				m[o.Env] = txt
			}
		}
	}
	for _, o := range IntOptions {
		if o.ExtendedUsage != "" {
			if txt, err := ExtendedUsage(o.ExtendedUsage, reg); err == nil {
				m[o.Env] = txt
			}
		}
	}
	for _, o := range BoolOptions {
		if o.ExtendedUsage != "" {
			if txt, err := ExtendedUsage(o.ExtendedUsage, reg); err == nil {
				m[o.Env] = txt
			}
		}
	}
	return m
}

// ExamplesMap returns a map of environment variable names to example values.
func ExamplesMap() map[string][]string {
	m := make(map[string][]string)
	for _, o := range StringOptions {
		if len(o.Examples) > 0 {
			m[o.Env] = append([]string(nil), o.Examples...)
		}
	}
	return m
}

// ValuesMap returns a map of environment variable names to the values stored
// in cfg.
func ValuesMap(cfg RuntimeConfig) map[string]string {
	m := make(map[string]string)
	for _, o := range StringOptions {
		m[o.Env] = *o.Target(&cfg)
	}
	for _, o := range IntOptions {
		m[o.Env] = strconv.Itoa(*o.Target(&cfg))
	}
	for _, o := range BoolOptions {
		m[o.Env] = strconv.FormatBool(*o.Target(&cfg))
	}
	return m
}
