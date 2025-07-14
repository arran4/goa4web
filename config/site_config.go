package config

import "strings"

func parseBool(v string) (bool, bool) {
	if v == "" {
		return false, false
	}
	switch strings.ToLower(v) {
	case "0", "false", "off", "no":
		return false, true
	case "1", "true", "on", "yes":
		return true, true
	default:
		return false, false
	}
}

// resolveBool returns the first valid boolean parsed from the provided values
// in order, defaulting to def when none are valid.
func resolveBool(def bool, vals ...string) bool {
	for _, v := range vals {
		if b, ok := parseBool(v); ok {
			return b
		}
	}
	return def
}
