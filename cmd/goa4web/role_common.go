package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func loadRole(sdb *sql.DB, roleName string, explicitFile string) error {
	var data []byte
	if explicitFile != "" {
		// Explicit filesystem file provided
		p := explicitFile
		if !strings.HasSuffix(strings.ToLower(p), ".sql") {
			p = p + ".sql"
		}
		abs, _ := filepath.Abs(p)
		log.Printf("Loading role %q from file %s", roleName, abs)
		b, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read role file: %w", err)
		}
		data = b
	} else {
		// Default to embedded role
		log.Printf("Loading role %q from embedded roles", roleName)
		b, err := readEmbeddedRole(roleName)
		if err != nil {
			// Try to list available roles if it's a "not found" error
			available, listErr := listEmbeddedRoles()
			if listErr == nil && len(available) > 0 {
				return fmt.Errorf("failed to read embedded role %q: %w. Available roles: %s", roleName, err, strings.Join(available, ", "))
			}
			return fmt.Errorf("failed to read embedded role %q: %w", roleName, err)
		}
		data = b
	}

	if err := runStatements(sdb, strings.NewReader(string(data))); err != nil {
		return fmt.Errorf("failed to apply role: %w", err)
	}

	log.Printf("Role %q loaded successfully.", roleName)
	return nil
}

func findEmbeddedRoleByName(targetName string) (string, error) {
	roles, err := listEmbeddedRoles()
	if err != nil {
		return "", err
	}
	for _, roleFile := range roles {
		name, err := readEmbeddedRoleName(roleFile)
		if err != nil {
			continue // Skip unreadable roles
		}
		if name == targetName {
			return roleFile, nil
		}
	}
	return "", fmt.Errorf("role %q not found in embedded roles", targetName)
}
