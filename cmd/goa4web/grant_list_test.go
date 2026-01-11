package main_test

import (
	"bytes"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/arran4/goa4web/cmd"
	"github.com/arran4/goa4web/internal/db"
	"github.com/golang/mock/gomock"
)

func TestGrantListCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		args           []string
		expectedParams db.ListGrantsExtendedParams
		expectedOutput string
		wantErr        bool
	}{
		{
			name: "no filter",
			args: []string{},
			expectedParams: db.ListGrantsExtendedParams{
				OnlyRoles: true,
			},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tRole: test-role\tID: 1\tYes\n",
		},
		{
			name: "filter by user id",
			args: []string{"-uid", "123"},
			expectedParams: db.ListGrantsExtendedParams{
				UserID:    sql.NullInt32{Int32: 123, Valid: true},
				OnlyUsers: true,
			},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tUser: test-user\tID: 1\tYes\n",
		},
		{
			name: "filter by username",
			args: []string{"-username", "test-user"},
			expectedParams: db.ListGrantsExtendedParams{
				Username:  "test-user",
				OnlyRoles: true,
			},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tUser: test-user\tID: 1\tYes\n",
		},
		{
			name: "filter by role id",
			args: []string{"-rid", "456"},
			expectedParams: db.ListGrantsExtendedParams{
				RoleID:    sql.NullInt32{Int32: 456, Valid: true},
				OnlyRoles: true,
			},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tRole: test-role\tID: 1\tYes\n",
		},
		{
			name: "filter by role name",
			args: []string{"-role-name", "test-role"},
			expectedParams: db.ListGrantsExtendedParams{
				RoleName:  "test-role",
				OnlyRoles: true,
			},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tRole: test-role\tID: 1\tYes\n",
		},
		{
			name: "filter users",
			args: []string{"-filter", "users"},
			expectedParams: db.ListGrantsExtendedParams{
				OnlyUsers: true,
			},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tUser: test-user\tID: 1\tYes\n",
		},
		{
			name:           "filter both",
			args:           []string{"-filter", "both"},
			expectedParams: db.ListGrantsExtendedParams{},
			expectedOutput: "ID\tSection\tItem\tAction\tRule Type\tTarget\tScope\tActive\n1\tsection\titem\taction\t-\tUser: test-user\tID: 1\tYes\n1\tsection\titem\taction\t-\tRole: test-role\tID: 1\tYes\n",
		},
		{
			name:    "invalid filter",
			args:    []string{"-filter", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := db.NewMockQuerier(ctrl)
			m.EXPECT().ListGrantsExtended(gomock.Any(), gomock.Eq(tt.expectedParams)).Return(getMockGrants(tt.expectedParams), nil).AnyTimes()
			var out bytes.Buffer
			rootCmd := cmd.NewRoot(&out, m)
			rootCmd.SetArgs(append([]string{"grant", "list"}, tt.args...))
			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("grant list cmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if !strings.Contains(strings.ReplaceAll(out.String(), " ", ""), strings.ReplaceAll(tt.expectedOutput, " ", "")) {
					t.Errorf("Expected output to contain '%s', but got '%s'", tt.expectedOutput, out.String())
				}
			}
		})
	}
}

func getMockGrants(params db.ListGrantsExtendedParams) []*db.ListGrantsExtendedRow {
	var rows []*db.ListGrantsExtendedRow
	if !params.OnlyUsers {
		rows = append(rows, &db.ListGrantsExtendedRow{
			ID:       1,
			Section:  "section",
			Item:     sql.NullString{String: "item", Valid: true},
			Action:   "action",
			RoleID:   sql.NullInt32{Int32: 1, Valid: true},
			RoleName: sql.NullString{String: "test-role", Valid: true},
			ItemID:   sql.NullInt32{Int32: 1, Valid: true},
			Active:   true,
		})
	}
	if !params.OnlyRoles {
		rows = append(rows, &db.ListGrantsExtendedRow{
			ID:       1,
			Section:  "section",
			Item:     sql.NullString{String: "item", Valid: true},
			Action:   "action",
			UserID:   sql.NullInt32{Int32: 1, Valid: true},
			Username: sql.NullString{String: "test-user", Valid: true},
			ItemID:   sql.NullInt32{Int32: 1, Valid: true},
			Active:   true,
		})
	}
	return rows
}

func TestGrantListCmdHelp(t *testing.T) {
	var out bytes.Buffer
	rootCmd := cmd.NewRoot(&out, nil)
	rootCmd.SetArgs([]string{"grant", "list", "-help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	output := out.String()

	expectedFlags := []string{
		"-filter",
		"-uid",
		"-user-id",
		"-username",
		"-rid",
		"-role-id",
		"-role-name",
	}
	for _, flag := range expectedFlags {
		match, err := regexp.MatchString(flag, output)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("Expected help output to contain %s", flag)
		}
	}
}
