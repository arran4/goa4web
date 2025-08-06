package admin

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// TestBuildGrantGroupsIncludesAvailableActionsWithoutGrants ensures that when a role has
// no grants, buildGrantGroups still returns groups for section/item pairs that do not
// require an item ID.
func TestBuildGrantGroupsIncludesAvailableActionsWithoutGrants(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants WHERE role_id = ? ORDER BY id\n")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.title, f.description\nFROM forumcategory f\n")).
		WillReturnRows(sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "title", "description"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language\n")).
		WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}))

	groups, err := buildGrantGroups(context.Background(), cd, 1)
	if err != nil {
		t.Fatalf("buildGrantGroups: %v", err)
	}
	expected := 0
	for _, def := range GrantActionMap {
		if !def.RequireItemID {
			expected++
		}
	}
	if len(groups) != expected {
		t.Fatalf("expected %d groups, got %d", expected, len(groups))
	}
	var topicFound bool
	var searchFound bool
	for _, g := range groups {
		if g.Section == "forum" && g.Item == "topic" {
			topicFound = true
		}
		if g.Section == "forum" && g.Item == "" && len(g.Available) == len(GrantActionMap["forum|"].Actions) {
			searchFound = true
		}
	}
	if topicFound {
		t.Fatalf("unexpected forum|topic group")
	}
	if !searchFound {
		t.Fatalf("missing forum| search group")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestBuildGrantGroupsSkipsInvalidItemIDGrants ensures that grants requiring an
// item ID are ignored when the item ID is missing.
func TestBuildGrantGroupsSkipsInvalidItemIDGrants(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}).
		AddRow(1, time.Now(), time.Now(), 0, 1, "forum", "topic", "", nil, "", "view", "", true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants WHERE role_id = ? ORDER BY id\n")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.title, f.description\nFROM forumcategory f\n")).
		WillReturnRows(sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "title", "description"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language\n")).
		WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}))

	groups, err := buildGrantGroups(context.Background(), cd, 1)
	if err != nil {
		t.Fatalf("buildGrantGroups: %v", err)
	}
	expected := 0
	for _, def := range GrantActionMap {
		if !def.RequireItemID {
			expected++
		}
	}
	if len(groups) != expected {
		t.Fatalf("expected %d groups, got %d", expected, len(groups))
	}
	for _, g := range groups {
		if g.Section == "forum" && g.Item == "topic" {
			t.Fatalf("unexpected forum|topic group")
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
