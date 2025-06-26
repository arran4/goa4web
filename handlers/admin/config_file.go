package admin

import (
	"log"
	"os"
	"strings"
)

// LoadAppConfigFile reads key=value pairs from the given path.
// Missing files return an empty map and unknown keys are ignored.
func LoadAppConfigFile(path string) map[string]string {
	values := make(map[string]string)
	if path == "" {
		return values
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
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
