package main

import (
	"database/sql"
	"flag"
	"io"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserUnverifiedEmailsCmd(t *testing.T) {
	cases := []struct {
		name          string
		args          []string
		setupMocks    func(sqlmock.Sqlmock)
		expectedError string
		outputContains []string
	}{
		{
			name: "list all",
			args: []string{"list"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemListAllUnverifiedEmails.*").
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verification_expires_at"}).
						AddRow(int32(1), int32(10), "test@example.com", sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}))
			},
			outputContains: []string{"test@example.com", "10"},
		},
		{
			name: "resend dry-run",
			args: []string{"resend", "-dry-run", "-all-time"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemListAllUnverifiedEmails.*").
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).
						AddRow(int32(1), int32(10), "test@example.com", sql.NullTime{}, sql.NullString{}, sql.NullTime{Time: time.Now(), Valid: true}, int32(0)))
			},
			outputContains: []string{"Would resend verification to", "test@example.com"},
		},
		{
			name: "expunge dry-run",
			args: []string{"expunge", "-dry-run", "-older-than", "24h"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s).*SystemListUnverifiedEmailsExpiresBefore.*").
					WithArgs(sqlmock.AnyArg()). // cutoff time
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).
						AddRow(int32(1), int32(10), "old@example.com", sql.NullTime{}, sql.NullString{}, sql.NullTime{Time: time.Now().Add(-48 * time.Hour), Valid: true}, int32(0)))
			},
			outputContains: []string{"Would expunge", "old@example.com"},
		},
		{
			name:          "missing subcommand",
			args:          []string{},
			expectedError: "missing unverified-emails subcommand",
		},
		{
			name:          "unknown subcommand",
			args:          []string{"unknown"},
			expectedError: "unknown unverified-emails subcommand \"unknown\"",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := &rootCmd{fs: flag.NewFlagSet("prog", flag.ContinueOnError)}

			conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer conn.Close()
			root.db = conn

			if tc.setupMocks != nil {
				tc.setupMocks(mock)
			}

			parent := &userCmd{rootCmd: root}
			// We need to mimic parent having a FlagSet if it's used, but here userCmd.fs is not used by subcmd except for Name().
			parent.fs = flag.NewFlagSet("user", flag.ContinueOnError)

			cmd, err := parseUserUnverifiedEmailsCmd(parent, tc.args)
			// If parsing fails (e.g. missing flags), it might return error or usage.
			// Here we assume parsing succeeds for valid cases.
			if err != nil {
				// Some tests might expect parsing error?
				// But parseUserUnverifiedEmailsCmd returns error only if fs.Parse fails.
				t.Fatalf("parseUserUnverifiedEmailsCmd: %v", err)
			}

			// Capture output
			r, w := io.Pipe()
			cmd.fs.SetOutput(w)

			errCh := make(chan error)
			go func() {
				errCh <- cmd.Run()
				w.Close()
			}()

			outBytes, _ := io.ReadAll(r)
			outStr := string(outBytes)

			runErr := <-errCh

			if tc.expectedError != "" {
				if runErr == nil {
					t.Fatalf("expected error %q, got nil", tc.expectedError)
				}
				if runErr.Error() != tc.expectedError {
					t.Fatalf("unexpected error: %q != %q", runErr.Error(), tc.expectedError)
				}
			} else {
				if runErr != nil {
					t.Fatalf("Run: %v", runErr)
				}
			}

			for _, s := range tc.outputContains {
				if !contains(outStr, s) {
					t.Errorf("output missing %q. Got:\n%s", s, outStr)
				}
			}

			if tc.setupMocks != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("unmet expectations: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
