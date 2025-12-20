package admin

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

type roleGrantGroupQueries struct {
	db.Querier
	grants []*db.Grant
}

func (q *roleGrantGroupQueries) AdminListGrantsByRoleID(context.Context, sql.NullInt32) ([]*db.Grant, error) {
	return q.grants, nil
}

func (q *roleGrantGroupQueries) GetAllForumCategories(context.Context, db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
	return []*db.Forumcategory{}, nil
}

func (q *roleGrantGroupQueries) SystemListLanguages(context.Context) ([]*db.Language, error) {
	return []*db.Language{}, nil
}

// TestBuildGrantGroupsIncludesAvailableActionsWithoutGrants ensures that when a role has
// no grants, buildGrantGroups still returns groups for section/item pairs that do not
// require an item ID.
func TestBuildGrantGroupsIncludesAvailableActionsWithoutGrants(t *testing.T) {
	q := &roleGrantGroupQueries{}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

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
}

// TestBuildGrantGroupsSkipsInvalidItemIDGrants ensures that grants requiring an
// item ID are ignored when the item ID is missing.
func TestBuildGrantGroupsSkipsInvalidItemIDGrants(t *testing.T) {
	q := &roleGrantGroupQueries{
		grants: []*db.Grant{{
			ID:      1,
			RoleID:  sql.NullInt32{Int32: 1, Valid: true},
			Section: "forum",
			Item:    sql.NullString{String: "topic", Valid: true},
			Action:  "view",
			Active:  true,
		}},
	}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

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
}
