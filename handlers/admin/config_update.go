package admin

import (
	"github.com/arran4/goa4web/core"
)

// updateConfigKey writes the given key/value pair to the config file.
// Existing keys are replaced, new keys appended. Empty values remove the key.
func (h *Handlers) updateConfigKey(path, key, value string) error {
	if h.UpdateConfigKeyFunc == nil {
		return nil
	}
	return h.UpdateConfigKeyFunc(core.OSFS{}, path, key, value)
}
