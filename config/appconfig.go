package config

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"sort"
)

// FileSystem abstracts file operations. It matches the core.FileSystem
// interface so core implementations can be passed in directly.
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// LoadAppConfigFile reads CONFIG_FILE style key=value pairs and returns them as a map.
// Missing files return an empty map. Unknown keys are ignored.
func LoadAppConfigFile(fs FileSystem, path string) map[string]string {
	values := make(map[string]string)
	if path == "" {
		return values
	}
	b, err := fs.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("app config file error: %v", err)
		}
		return values
	}
	return ParseEnvBytes(b)
}

// UpdateConfigKey writes the given key/value pair to the config file.
// Existing keys are replaced, new keys appended. Empty values remove the key.
func UpdateConfigKey(fs FileSystem, path, key, value string) error {
	if path == "" {
		return nil
	}
	cfg := LoadAppConfigFile(fs, path)
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
	return fs.WriteFile(path, buf.Bytes(), 0644)
}

// AddMissingJSONOptions ensures all keys from values exist in the JSON file at
// path. Missing keys are added with their values. The file is created when it
// does not exist.
func AddMissingJSONOptions(fs FileSystem, path string, values map[string]string) error {
	if path == "" {
		return nil
	}
	existing := make(map[string]string)
	if b, err := fs.ReadFile(path); err == nil {
		if len(b) > 0 {
			if err := json.Unmarshal(b, &existing); err != nil {
				return err
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	changed := false
	for k, v := range values {
		if _, ok := existing[k]; !ok {
			existing[k] = v
			changed = true
		}
	}
	if !changed {
		return nil
	}
	keys := make([]string, 0, len(existing))
	for k := range existing {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]string, len(existing))
	for _, k := range keys {
		ordered[k] = existing[k]
	}
	b, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return fs.WriteFile(path, b, 0644)
}
