package dbstart

import (
	"context"
	"database/sql"
)

type execCall struct {
	query string
	args  []any
}

type fakeQuerier struct {
	versionValue   int
	scanErr        error
	scannedVersion bool
	execLog        []execCall
}

func (f *fakeQuerier) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.execLog = append(f.execLog, execCall{query: query, args: args})
	return sqlmockResult{}, nil
}

func (f *fakeQuerier) QueryRowContext(_ context.Context, _ string, _ ...any) rowScanner {
	return fakeRow{scan: func(dest ...any) error {
		if f.scanErr != nil {
			return f.scanErr
		}
		if len(dest) > 0 {
			switch v := dest[0].(type) {
			case *int:
				*v = f.versionValue
			}
		}
		f.scannedVersion = true
		return nil
	}}
}

type fakeRow struct {
	scan func(dest ...any) error
}

func (f fakeRow) Scan(dest ...any) error {
	return f.scan(dest...)
}

type sqlmockResult struct{}

func (sqlmockResult) LastInsertId() (int64, error) { return 0, nil }
func (sqlmockResult) RowsAffected() (int64, error) { return 0, nil }
