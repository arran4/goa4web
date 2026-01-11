package main

import (
	"bufio"
	"bytes"
	"fmt"
	fs2 "io/fs"
	"path/filepath"
	"regexp"
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

func listEmbeddedRoleNames() ([]string, error) {
	roles, err := listEmbeddedRoles()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, role := range roles {
		name, err := readEmbeddedRoleName(role)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func readEmbeddedRoleName(identifier string) (string, error) {
	data, err := readEmbeddedRole(identifier)
	if err != nil {
		return "", err
	}
	name, err := parseEmbeddedRoleName(data)
	if err != nil {
		return "", fmt.Errorf("embedded role %q: %w", identifier, err)
	}
	return name, nil
}

var roleInsertNameRegexp = regexp.MustCompile(`(?is)insert\s+into\s+roles\s*\([^)]*name[^)]*\)\s*values\s*\(\s*'([^']+)'`)

func parseEmbeddedRoleName(data []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToLower(line), "-- role:") {
			name := strings.TrimSpace(line[len("-- role:"):])
			if name != "" {
				return name, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan role comments: %w", err)
	}
	if matches := roleInsertNameRegexp.FindSubmatch(data); len(matches) > 1 {
		return string(matches[1]), nil
	}
	return "", fmt.Errorf("role name not found")
}
