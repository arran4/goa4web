package runtimeconfig

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

func resolveFeedsEnabled(cli, file, env string) bool {
	if b, ok := parseBool(cli); ok {
		return b
	}
	if b, ok := parseBool(file); ok {
		return b
	}
	if b, ok := parseBool(env); ok {
		return b
	}
	return true
}
