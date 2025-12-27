package dbstart

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (fakeResult) RowsAffected() (int64, error) { return 0, nil }

type fakeRow struct {
	version int
	err     error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) == 0 {
		return nil
	}
	ptr, ok := dest[0].(*int)
	if !ok {
		return fmt.Errorf("expected *int destination, got %T", dest[0])
	}
	*ptr = r.version
	return nil
}

type fakeStore struct {
	version    int
	hasVersion bool
	execs      []string
}

func (f *fakeStore) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.execs = append(f.execs, fmt.Sprintf("%s %v", query, args))
	switch {
	case strings.HasPrefix(query, "INSERT INTO schema_version"):
		f.version = 1
		f.hasVersion = true
	case strings.HasPrefix(query, "UPDATE schema_version SET version = ?"):
		if len(args) > 0 {
			switch n := args[0].(type) {
			case int:
				f.version = n
			case int64:
				f.version = int(n)
			}
			f.hasVersion = true
		}
	}
	return fakeResult{}, nil
}

func (f *fakeStore) QueryRowContext(_ context.Context, _ string, _ ...any) rowScanner {
	if !f.hasVersion {
		return fakeRow{err: sql.ErrNoRows}
	}
	return fakeRow{version: f.version}
}
