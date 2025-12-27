package main

import (
	"database/sql"
	"flag"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestUserRenameCmd(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		args           []string
		setupStub      func(*db.QuerierStub)
		wantLookup     *sql.NullString
		wantUpdateArgs *db.AdminUpdateUsernameByIDParams
		expectedError  string
	}{
		{
			name: "positional args",
			args: []string{"alice", "alice2"},
			setupStub: func(stub *db.QuerierStub) {
				stub.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
					Idusers:  7,
					Username: sql.NullString{String: "alice", Valid: true},
				}
			},
			wantLookup:     &sql.NullString{String: "alice", Valid: true},
			wantUpdateArgs: &db.AdminUpdateUsernameByIDParams{Username: sql.NullString{String: "alice2", Valid: true}, Idusers: 7},
		},
		{
			name: "flag args",
			args: []string{"-from", "bob", "-to", "bob2"},
			setupStub: func(stub *db.QuerierStub) {
				stub.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
					Idusers:  9,
					Username: sql.NullString{String: "bob", Valid: true},
				}
			},
			wantLookup:     &sql.NullString{String: "bob", Valid: true},
			wantUpdateArgs: &db.AdminUpdateUsernameByIDParams{Username: sql.NullString{String: "bob2", Valid: true}, Idusers: 9},
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

			var stub *db.QuerierStub
			if tc.setupStub != nil {
				stub = &db.QuerierStub{}
				tc.setupStub(stub)
				root.queries = stub
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
			if stub != nil {
				if tc.wantLookup != nil {
					if got := len(stub.SystemGetUserByUsernameCalls); got != 1 {
						t.Fatalf("expected 1 user lookup, got %d", got)
					}
					if stub.SystemGetUserByUsernameCalls[0] != *tc.wantLookup {
						t.Fatalf("unexpected lookup: got %#v want %#v", stub.SystemGetUserByUsernameCalls[0], *tc.wantLookup)
					}
				}
				if tc.wantUpdateArgs != nil {
					if got := len(stub.AdminUpdateUsernameByIDCalls); got != 1 {
						t.Fatalf("expected 1 username update, got %d", got)
					}
					if stub.AdminUpdateUsernameByIDCalls[0] != *tc.wantUpdateArgs {
						t.Fatalf("unexpected update args: got %#v want %#v", stub.AdminUpdateUsernameByIDCalls[0], *tc.wantUpdateArgs)
					}
				}
			}
		})
	}
}
