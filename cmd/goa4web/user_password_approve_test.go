package main

import (
	"database/sql"
	"flag"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserPasswordApproveCmd(t *testing.T) {
	cases := []struct {
		name          string
		args          []string
		setupMocks    func(sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name: "by username",
			args: []string{"testuser"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemGetUserByUsername.*").
					WithArgs(sql.NullString{String: "testuser", Valid: true}).
					WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).AddRow(1, "testuser", sql.NullTime{}))
				mock.ExpectQuery("(?s).*GetPendingPassword.*").
					WithArgs(int32(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
						AddRow(1, 1, "newpassword", "bcrypt", "testcode", time.Now(), sql.NullTime{}))
				mock.ExpectBegin()
				mock.ExpectExec("(?s).*UpdateUserPassword.*").
					WithArgs(int32(1), "newpassword", sql.NullString{String: "bcrypt", Valid: true}).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("(?s).*DeletePendingPassword.*").
					WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "by id",
			args: []string{"--id=1"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*GetPendingPassword.*").
					WithArgs(int32(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
						AddRow(1, 1, "newpassword", "bcrypt", "testcode", time.Now(), sql.NullTime{}))
				mock.ExpectBegin()
				mock.ExpectExec("(?s).*UpdateUserPassword.*").
					WithArgs(int32(1), "newpassword", sql.NullString{String: "bcrypt", Valid: true}).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("(?s).*DeletePendingPassword.*").
					WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "by code",
			args: []string{"--code=testcode"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*GetPendingPasswordByCode.*").
					WithArgs("testcode").
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
						AddRow(1, 1, "newpassword", "bcrypt", "testcode", time.Now(), sql.NullTime{}))
				mock.ExpectBegin()
				mock.ExpectExec("(?s).*UpdateUserPassword.*").
					WithArgs(int32(1), "newpassword", sql.NullString{String: "bcrypt", Valid: true}).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("(?s).*DeletePendingPassword.*").
					WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name:          "missing identifier",
			args:          []string{},
			expectedError: "id, username, or code required",
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

			userParent, err := parseUserCmd(root, []string{})
			if err != nil {
				t.Fatalf("parseUserCmd: %v", err)
			}
			passwordParent, err := parseUserPasswordCmd(userParent, []string{})
			if err != nil {
				t.Fatalf("parseUserPasswordCmd: %v", err)
			}
			cmd, err := parseUserPasswordApproveCmd(passwordParent, tc.args)
			if err != nil {
				t.Fatalf("parseUserPasswordApproveCmd: %v", err)
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
