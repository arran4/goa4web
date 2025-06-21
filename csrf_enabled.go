package main

import (
	"os"
	"strings"

	config "github.com/arran4/goa4web/config"
)

// csrfEnabled reports if CSRF protection should be active.
func csrfEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvCSRFEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}
