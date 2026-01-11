package dbstart

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
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

func getAvailableMigrations(f fs.FS, driver string) ([]*migrationFile, error) {
	entries, err := fs.ReadDir(f, ".")
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	// Group files by version
	byVersion := make(map[int][]*migrationFile)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m, err := parseMigrationFilename(e.Name())
		if err != nil {
			// Skip non-conforming files
			continue
		}

		// Filter by driver: accept exact match or generic (empty driver)
		if m.Driver != "" && m.Driver != driver {
			continue
		}

		byVersion[m.Version] = append(byVersion[m.Version], m)
	}

	var result []*migrationFile
	var versions []int
	for v := range byVersion {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	for _, v := range versions {
		files := byVersion[v]
		var selected *migrationFile

		// Selection logic: Prefer specific driver over generic
		for _, f := range files {
			if f.Driver == driver {
				selected = f
				break
			}
		}
		if selected == nil {
			// Find generic one
			for _, f := range files {
				if f.Driver == "" {
					selected = f
					break
				}
			}
		}

		if selected != nil {
			result = append(result, selected)
		}
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
		if verbose {
			fmt.Printf("applying %s\n", m.Name)
		}
		if err := executeFile(ctx, db, f, m.Name, verbose); err != nil {
			return fmt.Errorf("execute %s: %w", m.Name, err)
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
