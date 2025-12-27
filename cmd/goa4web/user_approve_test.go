package main

import (
	"database/sql"
	"flag"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestUserApproveCmd(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		args          []string
		wantID        int
		wantUsername  string
		setupStub     func(*db.QuerierStub)
		wantLookup    *sql.NullString
		wantRoleParam *db.SystemCreateUserRoleParams
		expectedError string
	}{
		{
			name:         "positional username",
			args:         []string{"alice"},
			wantID:       7,
			wantUsername: "alice",
			setupStub: func(stub *db.QuerierStub) {
				stub.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
					Idusers:  7,
					Username: sql.NullString{String: "alice", Valid: true},
				}
			},
			wantLookup:    &sql.NullString{String: "alice", Valid: true},
			wantRoleParam: &db.SystemCreateUserRoleParams{UsersIdusers: 7, Name: "user"},
		},
		{
			name:   "id flag",
			args:   []string{"-id", "9"},
			wantID: 9,
			setupStub: func(stub *db.QuerierStub) {
				stub.SystemCreateUserRoleErr = nil
			},
			wantRoleParam: &db.SystemCreateUserRoleParams{UsersIdusers: 9, Name: "user"},
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
			t.Parallel()

			root := &rootCmd{fs: flag.NewFlagSet("prog", flag.ContinueOnError)}

			var stub *db.QuerierStub
			if tc.setupStub != nil {
				stub = &db.QuerierStub{}
				tc.setupStub(stub)
				root.queries = stub
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

			if stub != nil {
				if tc.wantLookup != nil {
					if got := len(stub.SystemGetUserByUsernameCalls); got != 1 {
						t.Fatalf("expected 1 user lookup, got %d", got)
					}
					if stub.SystemGetUserByUsernameCalls[0] != *tc.wantLookup {
						t.Fatalf("unexpected lookup: got %#v want %#v", stub.SystemGetUserByUsernameCalls[0], *tc.wantLookup)
					}
				} else if calls := len(stub.SystemGetUserByUsernameCalls); calls > 0 {
					t.Fatalf("unexpected user lookup calls: %d", calls)
				}
				if tc.wantRoleParam != nil {
					if got := len(stub.SystemCreateUserRoleCalls); got != 1 {
						t.Fatalf("expected 1 role insert, got %d", got)
					}
					if stub.SystemCreateUserRoleCalls[0] != *tc.wantRoleParam {
						t.Fatalf("unexpected role params: got %#v want %#v", stub.SystemCreateUserRoleCalls[0], *tc.wantRoleParam)
					}
				}
			}
		})
	}
}
