package admin

import (
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// LoadAppConfigFile reads key=value pairs from the given path.
// Missing files return an empty map and unknown keys are ignored.
func LoadAppConfigFile(fs core.FileSystem, path string) (map[string]string, error) {
	return config.LoadAppConfigFile(fs, path)
}
