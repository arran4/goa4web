package sqlutil

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestRunStatements(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     io.Reader
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "happy path",
			input: strings.NewReader(`
SELECT 1;
-- this is a comment

SELECT 2;
`),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "multiline statement",
			input: strings.NewReader(`
SELECT
    1
FROM
    test;
`),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("SELECT 1 FROM test").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "missing trailing semicolon",
			input: strings.NewReader(`
SELECT 1;
SELECT 2
`),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "database execution error",
			input: strings.NewReader(`
SELECT 1;
SELECT 2;
`),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("SELECT 1").WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "database execution error at end",
			input: strings.NewReader(`
SELECT 1
`),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("SELECT 1").WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "reader error",
			input: errorReader{},
			setupMock: func(mock sqlmock.Sqlmock) {
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)
			err = RunStatements(ctx, db, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunStatements() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
