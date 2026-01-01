package dbstart

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type migrationFile struct {
	Name        string
	Version     int
	Description string // For future use
	Driver      string
}

func (m migrationFile) String() string {
	parts := []string{fmt.Sprintf("%04d", m.Version)}
	if m.Description != "" {
		parts = append(parts, m.Description)
	}
	if m.Driver != "" {
		parts = append(parts, m.Driver)
	}
	return strings.Join(parts, ".") + ".sql"
}

func parseVersion(name string) (int, error) {
	var digits []rune
	for _, r := range name {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		} else {
			break
		}
	}
	if len(digits) == 0 {
		return 0, fmt.Errorf("no version digits found")
	}
	return strconv.Atoi(string(digits))
}

func parseMigrationFilename(name string) (*migrationFile, error) {
	if !strings.HasSuffix(name, ".sql") {
		return nil, fmt.Errorf("not a sql file")
	}

	base := strings.TrimSuffix(name, ".sql")

	version, err := parseVersion(base)
	if err != nil {
		return nil, err
	}

	// Calculate where digits end
	digitCount := 0
	for _, r := range base {
		if unicode.IsDigit(r) {
			digitCount++
		} else {
			break
		}
	}

	m := &migrationFile{
		Name:    name,
		Version: version,
	}

	// Check for driver suffix
	// We currently only support mysql explicitly, so we look for it.
	// If other drivers are added, this logic might need extension.
	if strings.HasSuffix(base, ".mysql") {
		m.Driver = "mysql"
		base = strings.TrimSuffix(base, ".mysql")
	}

	// The remainder is the description part
	if len(base) > digitCount {
		remainder := base[digitCount:]
		// Trim separators
		remainder = strings.TrimPrefix(remainder, ".")
		remainder = strings.TrimPrefix(remainder, "_")
		m.Description = remainder
	}

	return m, nil
}
