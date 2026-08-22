package common

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestHasGrantPrivateForumSemantics(t *testing.T) {
	tests := []struct {
		name       string
		section    string
		itemID     int32
		userID     int32
		isAdmin    bool
		testGrants []*db.Grant
		want       bool
	}{
		{
			name:    "Ordinary grant fallback works",
			section: "news",
			itemID:  42,
			userID:  1,
			testGrants: []*db.Grant{
				{Section: "news", Action: "see", Active: true},
			},
			want: true,
		},
		{
			name:    "Top-level privateforum behaves generic",
			section: "privateforum",
			itemID:  0, // Top-level
			userID:  1,
			testGrants: []*db.Grant{
				{Section: "privateforum", Action: "see", Active: true},
			},
			want: true,
		},
		{
			name:    "Specific privateforum exact user + exact item succeeds",
			section: "privateforum",
			itemID:  42,
			userID:  1,
			testGrants: []*db.Grant{
				{
					Section: "privateforum",
					Action:  "see",
					Active:  true,
					ItemID:  sql.NullInt32{Int32: 42, Valid: true},
					UserID:  sql.NullInt32{Int32: 1, Valid: true},
				},
			},
			want: true,
		},
		{
			name:    "Specific privateforum NULL item does not succeed",
			section: "privateforum",
			itemID:  42,
			userID:  1,
			testGrants: []*db.Grant{
				{
					Section: "privateforum",
					Action:  "see",
					Active:  true,
					UserID:  sql.NullInt32{Int32: 1, Valid: true},
				},
			},
			want: false,
		},
		{
			name:    "Specific privateforum item ID 0 does not succeed",
			section: "privateforum",
			itemID:  42,
			userID:  1,
			testGrants: []*db.Grant{
				{
					Section: "privateforum",
					Action:  "see",
					Active:  true,
					ItemID:  sql.NullInt32{Int32: 0, Valid: true},
					UserID:  sql.NullInt32{Int32: 1, Valid: true},
				},
			},
			want: false,
		},
		{
			name:    "Specific privateforum grant for item A does not authorize item B",
			section: "privateforum",
			itemID:  42,
			userID:  1,
			testGrants: []*db.Grant{
				{
					Section: "privateforum",
					Action:  "see",
					Active:  true,
					ItemID:  sql.NullInt32{Int32: 43, Valid: true},
					UserID:  sql.NullInt32{Int32: 1, Valid: true},
				},
			},
			want: false,
		},
		{
			name:    "Specific privateforum grant for user A does not authorize user B",
			section: "privateforum",
			itemID:  42,
			userID:  2,
			testGrants: []*db.Grant{
				{
					Section: "privateforum",
					Action:  "see",
					Active:  true,
					ItemID:  sql.NullInt32{Int32: 42, Valid: true},
					UserID:  sql.NullInt32{Int32: 1, Valid: true},
				},
			},
			want: false,
		},
		{
			name:    "Admin does not bypass specific privateforum resource",
			section: "privateforum",
			itemID:  42,
			userID:  1,
			isAdmin: true,
			want:    false,
		},
		{
			name:    "Admin bypasses ordinary specific resource",
			section: "news",
			itemID:  42,
			userID:  1,
			isAdmin: true,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, _, _ := sqlmock.New()
			queries := db.New(conn)
			cd := NewTestCoreData(t, queries)
			cd.UserID = tt.userID
			cd.cache.testGrants = tt.testGrants
			cd.cache.allRoles.Store([]*db.Role{})
			if tt.isAdmin {
				WithUserRoles([]string{})(cd) // reset default user role if any
				cd.cache.perms.Store([]*db.GetPermissionsByUserIDRow{
					{IsAdmin: true},
				})
				cd.AdminMode = true // Enable IsAdminMode() which cd.IsAdmin() depends on
			} else {
				cd.cache.perms.Store([]*db.GetPermissionsByUserIDRow{})
			}
			if got := cd.HasGrant(tt.section, "", "see", tt.itemID); got != tt.want {
				t.Errorf("HasGrant() = %v, want %v", got, tt.want)
			}
			_ = conn.Close()
		})
	}
}
