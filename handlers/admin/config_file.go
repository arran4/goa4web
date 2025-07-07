package admin

import (
	"log"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// LoadAppConfigFile reads key=value pairs from the given path.
// Missing files return an empty map and unknown keys are ignored.
func LoadAppConfigFile(fs core.FileSystem, path string) map[string]string {
	values, err := config.LoadAppConfigFile(fs, path)
	if err != nil {
		log.Printf("app config file error: %v", err)
		return map[string]string{}
	}
	return values
}
