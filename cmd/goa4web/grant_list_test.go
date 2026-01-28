package main

import (
	"bytes"
	"database/sql"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestGrantListCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		roles          []*db.Role
		grants         []*db.Grant
		grantsExtended []*db.ListGrantsExtendedRow
		expectedOutput string
		wantErr        bool
	}{
		{
			name: "no filter (default to roles)",
			args: []string{},
			roles: []*db.Role{
				{ID: 1, Name: "test-role", CanLogin: true},
			},
			grants: []*db.Grant{
				{
					ID:      1,
					RoleID:  sql.NullInt32{Int32: 1, Valid: true},
					Section: "section",
					Item:    sql.NullString{String: "item", Valid: true},
					Action:  "action",
					Active:  true,
				},
			},
			grantsExtended: []*db.ListGrantsExtendedRow{
				{
					ID:        1,
					CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
					UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
					RoleID:    sql.NullInt32{Int32: 1, Valid: true},
					Section:   "section",
					Item:      sql.NullString{String: "item", Valid: true},
					RuleType:  "-",
					ItemID:    sql.NullInt32{Int32: 1, Valid: true},
					Action:    "action",
					Active:    true,
					RoleName:  sql.NullString{String: "test-role", Valid: true},
				},
			},
			expectedOutput: "ID  Section  Item  Action  Rule Type  Target           Scope  Active\n1   section  item  action  -          Role: test-role  ID: 1  Yes\n",
		},
		{
			name: "filter users",
			args: []string{"-filter", "users"},
			roles: []*db.Role{
				{ID: 2, Name: "another-role", CanLogin: true},
			},
			grants: []*db.Grant{
				{
					ID:      1,
					UserID:  sql.NullInt32{Int32: 1, Valid: true},
					Section: "section",
					Item:    sql.NullString{String: "item", Valid: true},
					Action:  "action",
					Active:  true,
				},
			},
			grantsExtended: []*db.ListGrantsExtendedRow{
				{
					ID:        1,
					CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
					UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
					UserID:    sql.NullInt32{Int32: 1, Valid: true},
					Section:   "section",
					Item:      sql.NullString{String: "item", Valid: true},
					RuleType:  "-",
					ItemID:    sql.NullInt32{Int32: 1, Valid: true},
					Action:    "action",
					Active:    true,
					Username:  sql.NullString{String: "test-user", Valid: true},
				},
			},
			expectedOutput: "ID  Section  Item  Action  Rule Type  Target           Scope  Active\n1   section  item  action  -          User: test-user  ID: 1  Yes\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			qs.AdminListRolesReturns = tt.roles
			qs.ListGrantsReturns = tt.grants
			qs.ListGrantsExtendedReturns = tt.grantsExtended

			r := &rootCmd{
				fs:      flag.NewFlagSet("test", flag.ContinueOnError),
				querier: qs,
			}

			// Capture output
			var out bytes.Buffer
			r.fs.SetOutput(&out)

			parent := &grantCmd{rootCmd: r}

			cmd, err := parseGrantListCmd(parent, tt.args)
			if err != nil {
				t.Fatalf("parseGrantListCmd error: %v", err)
			}
			cmd.fs.SetOutput(&out)

			err = cmd.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("grantListCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.expectedOutput != "" {
				if !strings.Contains(out.String(), "ID") {
					t.Errorf("Output missing header")
				}
				if tt.name == "no filter (default to roles)" {
					if !strings.Contains(out.String(), "Role: test-role") {
						t.Errorf("Expected 'Role: test-role' in output, got: %s", out.String())
					}
				}
				if tt.name == "filter users" {
					if !strings.Contains(out.String(), "User: test-user") {
						t.Errorf("Expected 'User: test-user' in output, got: %s", out.String())
					}
				}
			}
		})
	}
}
