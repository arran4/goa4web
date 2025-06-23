package main

import (
	"bytes"
	"sort"
)

// updateConfigKey rewrites the configuration file path with the provided key value.
func updateConfigKey(path, key, value string) error {
	m := loadAppConfigFile(path)
	if m == nil {
		m = make(map[string]string)
	}
	m[key] = value
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, k := range keys {
		buf.WriteString(k + "=" + m[k] + "\n")
	}
	return writeFile(path, buf.Bytes(), 0644)
}
