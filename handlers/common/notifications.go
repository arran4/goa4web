package common

import (
	"os"
	"strings"

	config "github.com/arran4/goa4web/config"
)

// NotificationsEnabled reports if the internal notification system should run.
func NotificationsEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvNotificationsEnabled))
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
