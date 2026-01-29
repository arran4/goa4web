package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestBuildRoleGrantsExport(t *testing.T) {
	t.Parallel()

	queries := &db.QuerierStub{
		GetRoleByNameReturns: &db.Role{
			ID:            42,
			Name:          "moderator",
			CanLogin:      true,
			IsAdmin:       false,
			PrivateLabels: true,
		},
		AdminListGrantsByRoleIDReturns: []*db.Grant{
			{
				ID:      1,
				Section: "forum",
				Item:    sql.NullString{Valid: false},
				Action:  "search",
				Active:  true,
			},
		},
		GetAllForumCategoriesReturns: []*db.Forumcategory{},
		SystemListLanguagesReturns:   []*db.Language{},
	}

	export, err := buildRoleGrantsExport(context.Background(), queries, "moderator", nil)
	if err != nil {
		t.Fatalf("buildRoleGrantsExport: %v", err)
	}

	if export.Role.ID != 42 {
		t.Fatalf("role id mismatch: got %d want 42", export.Role.ID)
	}
	if export.Role.Name != "moderator" {
		t.Fatalf("role name mismatch: got %q want %q", export.Role.Name, "moderator")
	}
	if !export.Role.CanLogin {
		t.Fatalf("expected CanLogin to be true")
	}
	if export.Role.IsAdmin {
		t.Fatalf("expected IsAdmin to be false")
	}
	if !export.Role.PrivateLabels {
		t.Fatalf("expected PrivateLabels to be true")
	}

	var group *roleGrantsExportGroup
	for i := range export.GrantGroups {
		if export.GrantGroups[i].Section == "forum" && export.GrantGroups[i].Item == "" {
			group = &export.GrantGroups[i]
			break
		}
	}
	if group == nil {
		t.Fatalf("expected forum group to be present")
	}
	if group.ItemID.Valid {
		t.Fatalf("expected item_id to be invalid")
	}
	if len(group.Have) != 1 || group.Have[0].Name != "search" {
		t.Fatalf("unexpected have actions: %+v", group.Have)
	}
	if group.Have[0].Unsupported {
		t.Fatalf("expected have action to be supported")
	}
	if len(group.Available) != 0 {
		t.Fatalf("expected no available actions, got %v", group.Available)
	}
}

func TestWriteRoleGrantsExportCSV(t *testing.T) {
	t.Parallel()

	export := roleGrantsExport{
		Role: roleGrantsExportRole{
			ID:            7,
			Name:          "user",
			CanLogin:      true,
			IsAdmin:       false,
			PrivateLabels: false,
		},
		GrantGroups: []roleGrantsExportGroup{
			{
				Section: "forum",
				Item:    "",
				ItemID:  roleGrantsExportItemID{Valid: false},
				Have: []roleGrantsExportAction{
					{Name: "search"},
				},
				Available: []string{},
			},
		},
	}

	var buf bytes.Buffer
	if err := writeRoleGrantsExportCSV(&buf, export); err != nil {
		t.Fatalf("writeRoleGrantsExportCSV: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("expected header and data row, got %d rows", len(records))
	}
	if records[0][0] != "role_id" {
		t.Fatalf("unexpected header: %v", records[0])
	}
	if records[1][1] != "user" {
		t.Fatalf("unexpected role name: %v", records[1])
	}
}
