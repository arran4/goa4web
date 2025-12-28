package main

import (
	"database/sql"
	"flag"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserApproveCmd(t *testing.T) {
	cases := []struct {
		name          string
		args          []string
		wantID        int
		wantUsername  string
		setupMocks    func(sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name:         "positional username",
			args:         []string{"alice"},
			wantID:       7,
			wantUsername: "alice",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemGetUserByUsername.*").
					WithArgs(sql.NullString{String: "alice", Valid: true}).
					WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).AddRow(
						int32(7), sql.NullString{String: "alice", Valid: true}, sql.NullTime{},
					))
				mock.ExpectExec("(?s).*SystemCreateUserRole.*").
					WithArgs(int32(7), "user").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "id flag",
			args:   []string{"-id", "9"},
			wantID: 9,
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("(?s).*SystemCreateUserRole.*").
					WithArgs(int32(9), "user").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:          "missing identifier",
			args:          nil,
			expectedError: "id or username required",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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

			cmd, err := parseUserApproveCmd(parent, tc.args)
			if err != nil {
				t.Fatalf("parseUserApproveCmd: %v", err)
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

			if tc.wantID != 0 && cmd.ID != tc.wantID {
				t.Fatalf("unexpected id: got %d want %d", cmd.ID, tc.wantID)
			}
			if tc.wantUsername != "" && cmd.Username != tc.wantUsername {
				t.Fatalf("unexpected username: got %s want %s", cmd.Username, tc.wantUsername)
			}

			if tc.setupMocks != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("unmet expectations: %v", err)
				}
			}
		})
	}
}
