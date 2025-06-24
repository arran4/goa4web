package goa4web

import (
	"bytes"
	"log"
	"os"
	"sort"
	"strings"
)

// LoadAppConfigFile reads CONFIG_FILE style key=value pairs and returns them as a map.
// Missing files return an empty map. Unknown keys are ignored.
func LoadAppConfigFile(path string) map[string]string {
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

// updateConfigKey writes the given key/value pair to the config file.
// Existing keys are replaced, new keys appended. Empty values remove the key.
func updateConfigKey(path, key, value string) error {
	if path == "" {
		return nil
	}
	cfg := LoadAppConfigFile(path)
	if value == "" {
		delete(cfg, key)
	} else {
		cfg[key] = value
	}
	var keys []string
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, k := range keys {
		buf.WriteString(k + "=" + cfg[k] + "\n")
	}
	return writeFile(path, buf.Bytes(), 0644)
}
