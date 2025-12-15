package main

import (
	"fmt"
	fs2 "io/fs"
	"path/filepath"
	"sort"
	"strings"

	dbassets "github.com/arran4/goa4web/database"
)

func readEmbeddedRole(name string) ([]byte, error) {
	// Accept names with or without .sql suffix
	filename := name
	if !strings.HasSuffix(strings.ToLower(filename), ".sql") {
		filename = filename + ".sql"
	}
	path := filepath.ToSlash(filepath.Join("roles", filename))
	b, err := dbassets.RolesFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("embedded role %q not found: %w", name, err)
	}
	return b, nil
}

func listEmbeddedRoles() ([]string, error) {
	entries, err := fs2.ReadDir(dbassets.RolesFS, "roles")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			name = strings.TrimSuffix(name, ".sql")
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}
