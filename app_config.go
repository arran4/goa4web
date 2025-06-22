package main

import (
	"log"
	"os"
	"strings"
)

// loadAppConfigFile reads CONFIG_FILE style key=value pairs and returns them as a map.
// Missing files return an empty map. Unknown keys are ignored.
func loadAppConfigFile(path string) map[string]string {
	values := make(map[string]string)
	if path == "" {
		return values
	}
	b, err := readFile(path)
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
