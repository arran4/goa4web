package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"log"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestUserRenameCmd(t *testing.T) {
	cases := []struct {
		name           string
		args           []string
		setupFake      func(*fakeUserRenameQueries)
		expectedLog    string
		expectedUpdate *db.AdminUpdateUsernameByIDParams
		expectedError  string
	}{
		{
			name:        "positional args",
			args:        []string{"alice", "alice2"},
			expectedLog: "renamed alice to alice2",
			expectedUpdate: &db.AdminUpdateUsernameByIDParams{
				Username: sql.NullString{String: "alice2", Valid: true},
				Idusers:  int32(7),
			},
			setupFake: func(fake *fakeUserRenameQueries) {
				fake.user = &db.SystemGetUserByUsernameRow{
					Idusers:                int32(7),
					Username:               sql.NullString{String: "alice", Valid: true},
					PublicProfileEnabledAt: sql.NullTime{},
				}
			},
		},
		{
			name:        "flag args",
			args:        []string{"-from", "bob", "-to", "bob2"},
			expectedLog: "renamed bob to bob2",
			expectedUpdate: &db.AdminUpdateUsernameByIDParams{
				Username: sql.NullString{String: "bob2", Valid: true},
				Idusers:  int32(9),
			},
			setupFake: func(fake *fakeUserRenameQueries) {
				fake.user = &db.SystemGetUserByUsernameRow{
					Idusers:                int32(9),
					Username:               sql.NullString{String: "bob", Valid: true},
					PublicProfileEnabledAt: sql.NullTime{},
				}
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
			root := &rootCmd{fs: flag.NewFlagSet("prog", flag.ContinueOnError)}

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

			var logBuf bytes.Buffer
			restore := swapLogOutput(&logBuf)
			defer restore()

			if tc.setupFake != nil {
				fake := &fakeUserRenameQueries{t: t}
				tc.setupFake(fake)
				cmd.queries = fake
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
			if tc.expectedLog != "" && !containsLine(logBuf.String(), tc.expectedLog) {
				t.Fatalf("expected log output %q, got %q", tc.expectedLog, logBuf.String())
			}
			if tc.expectedUpdate != nil {
				if cmd.queries.(*fakeUserRenameQueries).updated == nil {
					t.Fatalf("expected update to be called")
				}
				if *cmd.queries.(*fakeUserRenameQueries).updated != *tc.expectedUpdate {
					t.Fatalf("unexpected update args: %+v", *cmd.queries.(*fakeUserRenameQueries).updated)
				}
			}
		})
	}
}

type fakeUserRenameQueries struct {
	t         *testing.T
	user      *db.SystemGetUserByUsernameRow
	getErr    error
	updateErr error
	updated   *db.AdminUpdateUsernameByIDParams
}

func (f *fakeUserRenameQueries) SystemGetUserByUsername(_ context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
	if f.user != nil && f.user.Username.String != username.String {
		f.t.Fatalf("unexpected username lookup %q", username.String)
	}
	return f.user, f.getErr
}

func (f *fakeUserRenameQueries) AdminUpdateUsernameByID(_ context.Context, arg db.AdminUpdateUsernameByIDParams) error {
	f.updated = &arg
	return f.updateErr
}

func swapLogOutput(buf *bytes.Buffer) func() {
	prevFlags := log.Flags()
	prevOutput := log.Writer()
	log.SetFlags(0)
	log.SetOutput(buf)
	return func() {
		log.SetFlags(prevFlags)
		log.SetOutput(prevOutput)
	}
}

func containsLine(content, want string) bool {
	for _, line := range bytes.Split([]byte(content), []byte("\n")) {
		if string(bytes.TrimSpace(line)) == want {
			return true
		}
	}
	return false
}
