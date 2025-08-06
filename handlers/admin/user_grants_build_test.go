package admin

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// TestBuildUserGrantGroupsIncludesAvailableActionsWithoutGrants ensures groups exist when user has no grants.
func TestBuildUserGrantGroupsIncludesAvailableActionsWithoutGrants(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants WHERE user_id = ? ORDER BY id\n")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}))
	qs := regexp.QuoteMeta("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.language_idlanguage, f.title, f.description\nFROM forumcategory f\nWHERE (")
	mock.ExpectQuery(qs).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_idlanguage", "title", "description"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language\n")).
		WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}))

	groups, err := buildGrantGroupsForUser(context.Background(), cd, 1)
	if err != nil {
		t.Fatalf("buildGrantGroupsForUser: %v", err)
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
