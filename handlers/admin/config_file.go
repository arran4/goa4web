package admin

import (
	"errors"
	iofs "io/fs"
	"log"
	"strings"

	"github.com/arran4/goa4web/core"
)

// LoadAppConfigFile reads key=value pairs from the given path.
// Missing files return an empty map and unknown keys are ignored.
func LoadAppConfigFile(fs core.FileSystem, path string) map[string]string {
	values := make(map[string]string)
	if path == "" {
		return values
	}
	b, err := fs.ReadFile(path)
	if err != nil {
		if !errors.Is(err, iofs.ErrNotExist) {
			log.Printf("app config file error: %v", err)
		}
		return values
	}
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.IndexByte(line, '='); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			values[key] = val
		}
	}
	return values
}
