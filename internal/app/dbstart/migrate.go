package dbstart

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// ensureVersionTable creates the schema_version table when missing and
// returns the current version number stored in the table.
func ensureVersionTable(ctx context.Context, db *sql.DB) (int, error) {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)"); err != nil {
		return 0, fmt.Errorf("create schema_version: %w", err)
	}
	var version int
	err := db.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version)
	if err == sql.ErrNoRows {
		if _, err := db.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (?)", 1); err != nil {
			return 0, fmt.Errorf("init schema_version: %w", err)
		}
		version = 1
	} else if err != nil {
		return 0, fmt.Errorf("select schema_version: %w", err)
	}
	return version, nil
}

type migrationFile struct {
	Name    string
	Version int
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

func getAvailableMigrations(f fs.FS, driver string) ([]migrationFile, error) {
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	// Map version to filename, preferring specific driver over generic
	migrations := make(map[int]string)
	driverSuffix := "." + driver + ".sql"

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		n, err := parseVersion(name)
		if err != nil {
			continue
		}

		isSpecific := strings.HasSuffix(name, driverSuffix)

		// Check if it looks like a migration file for another driver
		// We crudely assume that if it's not our driver suffix, but has a middle dot part that isn't just "sql", it might be another driver.
		// However, keeping it simple: accept all .sql, prefer specific.
		// Since we removed other drivers, conflicts are unlikely unless someone explicitly kept postgres files.

		isValid := true // Assume valid unless we detect conflict logic (omitted for now)

		if !isValid {
			continue
		}

		existing, exists := migrations[n]
		if exists {
			existingIsSpecific := strings.HasSuffix(existing, driverSuffix)
			if existingIsSpecific && !isSpecific {
				continue // Keep specific
			}
			if !existingIsSpecific && isSpecific {
				migrations[n] = name // Upgrade to specific
				continue
			}
			// Collision of same specificity, keep existing (lexicographical or arbitrary)
		} else {
			migrations[n] = name
		}
	}

	var nums []int
	for n := range migrations {
		nums = append(nums, n)
	}
	sort.Ints(nums)

	var result []migrationFile
	for _, n := range nums {
		result = append(result, migrationFile{Version: n, Name: migrations[n]})
	}

	return result, nil
}

// Apply reads SQL migration files from the provided filesystem and executes
// each one in order, updating the schema_version table after every successful
// script. When verbose is true, progress information is printed to stdout.
func Apply(ctx context.Context, db *sql.DB, f fs.FS, verbose bool, driver string) error {
	version, err := ensureVersionTable(ctx, db)
	if err != nil {
		return err
	}
	migrations, err := getAvailableMigrations(f, driver)
	if err != nil {
		return err
	}
	applied := false
	for _, m := range migrations {
		if m.Version <= version {
			continue
		}
		path := m.Name
		if verbose {
			fmt.Printf("applying %s\n", path)
		}
		if err := executeFile(ctx, db, f, path, verbose); err != nil {
			return fmt.Errorf("execute %s: %w", path, err)
		}
		if _, err := db.ExecContext(ctx, "UPDATE schema_version SET version = ?", m.Version); err != nil {
			return fmt.Errorf("update schema_version: %w", err)
		}
		if verbose {
			fmt.Printf("schema version updated to %d\n", m.Version)
		}
		applied = true
	}
	if verbose {
		if applied {
			fmt.Println("migration complete")
		} else {
			fmt.Println("database schema already up to date")
		}
	}
	return nil
}

func executeFile(ctx context.Context, db *sql.DB, f fs.FS, path string, verbose bool) error {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var stmt strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}
		stmt.WriteString(line)
		if strings.HasSuffix(line, ";") {
			sqlStmt := strings.TrimSuffix(stmt.String(), ";")
			if verbose {
				fmt.Printf("  %s\n", sqlStmt)
			}
			if _, err := db.ExecContext(ctx, sqlStmt); err != nil {
				return err
			}
			stmt.Reset()
		} else {
			stmt.WriteString(" ")
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if s := strings.TrimSpace(stmt.String()); s != "" {
		if verbose {
			fmt.Printf("  %s\n", s)
		}
		if _, err := db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
