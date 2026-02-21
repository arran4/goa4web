package forum

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

// TestBuildTopicGrantGroupsIncludesAllRoles ensures groups are returned for each action when no grants exist.
func TestBuildTopicGrantGroupsIncludesAllRoles(t *testing.T) {
	common.ResetGlobalRolesCache()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles ORDER BY id\n")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}).
			AddRow(1, "member", true, false, true, time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants ORDER BY id\n")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}))

	groups, err := buildTopicGrantGroups(context.Background(), cd, 1)
	if err != nil {
		t.Fatalf("buildTopicGrantGroups: %v", err)
	}
	if len(groups) != 5 {
		t.Fatalf("expected 5 groups, got %d", len(groups))
	}
	for _, g := range groups {
		if len(g.Have) != 0 || len(g.Disabled) != 0 || len(g.Available) != 2 {
			t.Fatalf("unexpected group %+v", g)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
