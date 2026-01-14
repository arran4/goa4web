package main

import (
	"bytes"
	"database/sql"
	"flag"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGrantListCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupMocks     func(sqlmock.Sqlmock)
		expectedOutput string
		wantErr        bool
	}{
		{
			name: "no filter (default to roles)",
			args: []string{},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT g.id, g.created_at, g.updated_at, g.user_id, g.role_id, g.section, g.item, g.rule_type, g.item_id, g.item_rule, g.action, g.extra, g.active, u.username, r.name as role_name FROM grants g LEFT JOIN users u ON g.user_id = u.idusers LEFT JOIN roles r ON g.role_id = r.id WHERE (? IS NULL OR g.user_id = ?) AND (? IS NULL OR u.username = ?) AND (? IS NULL OR g.role_id = ?) AND (? IS NULL OR r.name = ?) AND (? = false OR g.role_id IS NOT NULL) AND (? = false OR g.user_id IS NOT NULL) ORDER BY g.id")).
					WithArgs(
						sql.NullInt32{}, sql.NullInt32{}, // UserID (IS NULL, user_id=?)
						sql.NullString{}, sql.NullString{}, // Username
						sql.NullInt32{}, sql.NullInt32{}, // RoleID
						sql.NullString{}, sql.NullString{}, // RoleName
						true, // OnlyRoles
						nil,  // OnlyUsers
					).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active", "username", "role_name",
					}).AddRow(
						1, time.Now(), time.Now(), sql.NullInt32{}, 1, "section", "item", "-", 1, "", "action", "", true, nil, "test-role"))
			},
			expectedOutput: "ID  Section  Item  Action  Rule Type  Target           Scope  Active\n1   section  item  action  -          Role: test-role  ID: 1  Yes\n",
		},
		{
			name: "filter users",
			args: []string{"-filter", "users"},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT g.id, g.created_at, g.updated_at, g.user_id, g.role_id, g.section, g.item, g.rule_type, g.item_id, g.item_rule, g.action, g.extra, g.active, u.username, r.name as role_name FROM grants g LEFT JOIN users u ON g.user_id = u.idusers LEFT JOIN roles r ON g.role_id = r.id WHERE (? IS NULL OR g.user_id = ?) AND (? IS NULL OR u.username = ?) AND (? IS NULL OR g.role_id = ?) AND (? IS NULL OR r.name = ?) AND (? = false OR g.role_id IS NOT NULL) AND (? = false OR g.user_id IS NOT NULL) ORDER BY g.id")).
					WithArgs(
						sql.NullInt32{}, sql.NullInt32{}, // UserID
						sql.NullString{}, sql.NullString{}, // Username
						sql.NullInt32{}, sql.NullInt32{}, // RoleID
						sql.NullString{}, sql.NullString{}, // RoleName
						nil,  // OnlyRoles
						true, // OnlyUsers
					).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active", "username", "role_name",
					}).AddRow(
						1, time.Now(), time.Now(), 1, sql.NullInt32{}, "section", "item", "-", 1, "", "action", "", true, "test-user", nil))
			},
			expectedOutput: "ID  Section  Item  Action  Rule Type  Target           Scope  Active\n1   section  item  action  -          User: test-user  ID: 1  Yes\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if tt.setupMocks != nil {
				tt.setupMocks(mock)
			}

			r := &rootCmd{
				fs: flag.NewFlagSet("test", flag.ContinueOnError),
				db: db,
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

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
