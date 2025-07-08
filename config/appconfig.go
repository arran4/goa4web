package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arran4/goa4web/core"
)

// ErrConfigFileNotFound is returned when the requested configuration file is missing.
var ErrConfigFileNotFound = errors.New("config file not found")

// LoadAppConfigFile reads CONFIG_FILE style key=value pairs or JSON objects and
// returns them as a map. Missing files return an empty map.
// Supported extensions are ".env" and ".json". An error is returned for any
// other extension.
// LoadAppConfigFile reads CONFIG_FILE style key=value pairs and returns them as a map.
// Missing files return an empty map. Unknown keys are ignored.
func LoadAppConfigFile(fs core.FileSystem, path string) (map[string]string, error) {
	values := make(map[string]string)
	if path == "" {
		log.Printf("config file not specified")
		return values, nil
	}
	log.Printf("reading config file %s", path)
	b, err := fs.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("config file not found: %s", path)
			return values, ErrConfigFileNotFound
		}
		return nil, fmt.Errorf("app config file error: %w", err)
	}
	log.Printf("loaded config file %s", path)
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(b, &values); err != nil {
			log.Printf("app config file parse error: %v", err)
			return nil, fmt.Errorf("parse json: %w", err)
		}
		return values, nil
	case ".env":
		return ParseEnvBytes(b), nil
	default:
		return nil, fmt.Errorf("unsupported config extension %q: use .env or .json", filepath.Ext(path))
	}
}

// UpdateConfigKey writes the given key/value pair to the config file.
// Existing keys are replaced, new keys appended. Empty values remove the key.
func UpdateConfigKey(fs core.FileSystem, path, key, value string) error {
	if path == "" {
		return nil
	}
	cfg, err := LoadAppConfigFile(fs, path)
	if err != nil && !errors.Is(err, ErrConfigFileNotFound) {
		return err
	}
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
func AddMissingJSONOptions(fs core.FileSystem, path string, values map[string]string) error {
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
