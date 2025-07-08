package admin

import (
	"github.com/arran4/goa4web/core"
)

// updateConfigKey writes the given key/value pair to the config file.
// Existing keys are replaced, new keys appended. Empty values remove the key.
func updateConfigKey(path, key, value string) error {
	if UpdateConfigKeyFunc == nil {
		return nil
	}
	return UpdateConfigKeyFunc(core.OSFS{}, path, key, value)
}
