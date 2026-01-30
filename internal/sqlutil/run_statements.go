package sqlutil

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

// RunStatements executes semicolon-delimited SQL statements from the reader.
func RunStatements(ctx context.Context, sdb *sql.DB, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var stmt strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}
		stmt.WriteString(line)
		if strings.HasSuffix(line, ";") {
			sqlStmt := strings.TrimSuffix(stmt.String(), ";")
			if _, err := sdb.ExecContext(ctx, sqlStmt); err != nil {
				return fmt.Errorf("executing statement %q: %w", sqlStmt, err)
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
		if _, err := sdb.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("executing statement %q: %w", s, err)
		}
	}
	return nil
}
