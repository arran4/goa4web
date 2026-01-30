package roles

import (
	"fmt"
	fs2 "io/fs"
	"path/filepath"
	"sort"
	"strings"

	dbassets "github.com/arran4/goa4web/database"
)

// ReadEmbeddedRole reads an embedded role SQL file by name.
func ReadEmbeddedRole(name string) ([]byte, error) {
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

// ListEmbeddedRoles lists all embedded role identifiers.
func ListEmbeddedRoles() ([]string, error) {
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

// ListEmbeddedRoleNames reads the role names declared inside embedded SQL.
func ListEmbeddedRoleNames() ([]string, error) {
	roles, err := ListEmbeddedRoles()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, role := range roles {
		name, err := ReadEmbeddedRoleName(role)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

// ReadEmbeddedRoleName reads the declared role name from an embedded role file.
func ReadEmbeddedRoleName(identifier string) (string, error) {
	data, err := ReadEmbeddedRole(identifier)
	if err != nil {
		return "", err
	}
	name, err := ParseRoleName(data)
	if err != nil {
		return "", fmt.Errorf("embedded role %q: %w", identifier, err)
	}
	return name, nil
}

// FindEmbeddedRoleByName searches embedded roles for a matching declared name.
func FindEmbeddedRoleByName(targetName string) (string, error) {
	roles, err := ListEmbeddedRoles()
	if err != nil {
		return "", err
	}
	for _, roleFile := range roles {
		name, err := ReadEmbeddedRoleName(roleFile)
		if err != nil {
			continue
		}
		if name == targetName {
			return roleFile, nil
		}
	}
	return "", fmt.Errorf("role %q not found in embedded roles", targetName)
}
