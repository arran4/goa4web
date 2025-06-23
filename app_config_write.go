package main

import (
	"bytes"
	"sort"
)

// updateConfigKey writes the given key/value pair to the config file.
// Existing keys are replaced, new keys appended. Empty values remove the key.
func updateConfigKey(path, key, value string) error {
	if path == "" {
		return nil
	}
	cfg := loadAppConfigFile(path)
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
