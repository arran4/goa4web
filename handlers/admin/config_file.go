package admin

import (
	"errors"
	iofs "io/fs"
	"log"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// LoadAppConfigFile reads key=value pairs from the given path.
// Missing files return an empty map and unknown keys are ignored.
func LoadAppConfigFile(fs core.FileSystem, path string) map[string]string {
        values := make(map[string]string)
        if path == "" {
                log.Printf("config file not specified")
                return values
        }
        log.Printf("reading config file %s", path)
        b, err := fs.ReadFile(path)
        if err != nil {
                if errors.Is(err, iofs.ErrNotExist) {
                        log.Printf("config file not found: %s", path)
                } else {
                        log.Printf("app config file error: %v", err)
                }
                return values
        }
        log.Printf("loaded config file %s", path)
        return config.ParseEnvBytes(b)
}
