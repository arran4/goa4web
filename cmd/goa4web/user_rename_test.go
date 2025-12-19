package main

import (
	"database/sql"
	"flag"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserRenameCmd(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		args          []string
		setupMocks    func(sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name: "positional args",
			args: []string{"alice", "alice2"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemGetUserByUsername.*").
					WithArgs(sql.NullString{String: "alice", Valid: true}).
					WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).AddRow(
						int32(7), sql.NullString{String: "alice", Valid: true}, sql.NullTime{},
					))
				mock.ExpectExec("(?s).*AdminUpdateUsernameByID.*").
					WithArgs(sql.NullString{String: "alice2", Valid: true}, int32(7)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "flag args",
			args: []string{"-from", "bob", "-to", "bob2"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemGetUserByUsername.*").
					WithArgs(sql.NullString{String: "bob", Valid: true}).
					WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).AddRow(
						int32(9), sql.NullString{String: "bob", Valid: true}, sql.NullTime{},
					))
				mock.ExpectExec("(?s).*AdminUpdateUsernameByID.*").
					WithArgs(sql.NullString{String: "bob2", Valid: true}, int32(9)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:          "missing args",
			args:          nil,
			expectedError: "from and to usernames required",
		},
		{
			name:          "too many args",
			args:          []string{"one", "two", "three"},
			expectedError: "too many arguments",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := &rootCmd{fs: flag.NewFlagSet("prog", flag.ContinueOnError)}

			var mock sqlmock.Sqlmock
			if tc.setupMocks != nil {
				conn, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
				if err != nil {
					t.Fatalf("sqlmock.New: %v", err)
				}
				mock = m
				root.db = conn
				tc.setupMocks(mock)
			}

			parent := &userCmd{rootCmd: root}
			cmd, err := parseUserRenameCmd(parent, tc.args)
			if tc.expectedError != "" && err != nil {
				if err.Error() != tc.expectedError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseUserRenameCmd: %v", err)
			}

			err = cmd.Run()
			if tc.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.expectedError)
				}
				if err.Error() != tc.expectedError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if tc.setupMocks != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("unmet expectations: %v", err)
				}
			}
		})
	}
}
