package common

import "github.com/arran4/goa4web/config"

// NotificationsEnabled reports if the internal notification system should run.
// NotificationsEnabled reports if the internal notification system should run
// according to the runtime configuration.
func NotificationsEnabled() bool {
	return config.AppRuntimeConfig.NotificationsEnabled
}
