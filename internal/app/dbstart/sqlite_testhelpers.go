//go:build sqlite

package dbstart

import (
	"bufio"
	"database/sql"
	"regexp"
	"strings"
	"testing"
)

var (
	reInt     = regexp.MustCompile(`int\(\d+\)`)
	reTinyInt = regexp.MustCompile(`tinyint\(\d+\)`)
)

func sanitizeForSQLite(s string) string {
	s = strings.ReplaceAll(s, "`", "\"")
	s = strings.ReplaceAll(s, "AUTO_INCREMENT", "")
	s = reInt.ReplaceAllString(s, "INTEGER")
	s = reTinyInt.ReplaceAllString(s, "INTEGER")
	s = strings.ReplaceAll(s, "mediumtext", "TEXT")
	s = strings.ReplaceAll(s, "longtext", "TEXT")
	s = strings.ReplaceAll(s, "tinytext", "TEXT")
	s = strings.ReplaceAll(s, "datetime", "TEXT")
	s = strings.ReplaceAll(s, "DATETIME", "TEXT")
	s = strings.ReplaceAll(s, "NOW()", "CURRENT_TIMESTAMP")
	s = strings.ReplaceAll(s, "ON UPDATE CURRENT_TIMESTAMP", "")
	return s
}

func execSQL(t *testing.T, db *sql.DB, sqlText string) {
	scanner := bufio.NewScanner(strings.NewReader(sqlText))
	var parts []string
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(raw, "--") || raw == "" {
			continue
		}
		if i := strings.Index(raw, "--"); i >= 0 {
			raw = strings.TrimSpace(raw[:i])
			if raw == "" {
				continue
			}
		}
		if strings.HasPrefix(raw, "KEY ") || strings.HasPrefix(raw, "UNIQUE KEY") {
			continue
		}
		line := sanitizeForSQLite(raw)
		if strings.HasPrefix(line, ")") && len(parts) > 0 {
			last := strings.TrimSpace(parts[len(parts)-1])
			if strings.HasSuffix(last, ",") {
				parts[len(parts)-1] = strings.TrimSuffix(last, ",")
			}
		}
		if strings.HasSuffix(line, ";") {
			parts = append(parts, strings.TrimSuffix(line, ";"))
			stmt := strings.Join(parts, " ")
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("exec %s: %v", stmt, err)
			}
			parts = nil
		} else {
			parts = append(parts, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(parts) > 0 {
		stmt := strings.Join(parts, " ")
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %s: %v", stmt, err)
		}
	}
}
