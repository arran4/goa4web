package main

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestPrintGrantsTable(t *testing.T) {
	tests := []struct {
		name     string
		rows     []*db.ListGrantsExtendedRow
		expected []string
	}{
		{
			name: "User Grant",
			rows: []*db.ListGrantsExtendedRow{
				{
					ID:       1,
					Section:  "forum",
					Action:   "view",
					RuleType: "allow",
					Username: sql.NullString{String: "user1", Valid: true},
					UserID:   sql.NullInt32{Int32: 10, Valid: true},
				},
			},
			expected: []string{
				"ID", "Section", "Item", "Action", "Rule Type", "Target",
				"1", "forum", "view", "allow", "User: user1",
			},
		},
		{
			name: "Role Grant",
			rows: []*db.ListGrantsExtendedRow{
				{
					ID:       1,
					Section:  "forum",
					Action:   "view",
					RuleType: "allow",
					RoleName: sql.NullString{String: "admin", Valid: true},
					RoleID:   sql.NullInt32{Int32: 5, Valid: true},
				},
			},
			expected: []string{
				"ID", "Section", "Item", "Action", "Rule Type", "Target",
				"1", "forum", "view", "allow", "Role: admin",
			},
		},
		{
			name: "Row with Item",
			rows: []*db.ListGrantsExtendedRow{
				{
					ID:       2,
					Section:  "blog",
					Item:     sql.NullString{String: "post", Valid: true},
					Action:   "edit",
					RuleType: "deny",
				},
			},
			expected: []string{
				"2", "blog", "post", "edit", "deny",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := printGrantsTable(&buf, tt.rows); err != nil {
				t.Fatalf("printGrantsTable() error = %v", err)
			}
			output := buf.String()
			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, but got:\n%s", exp, output)
				}
			}
		})
	}
}
