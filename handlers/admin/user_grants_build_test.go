package admin

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

type userGrantGroupQueries struct {
	db.Querier
	grants []*db.Grant
}

func (q *userGrantGroupQueries) ListGrantsByUserID(context.Context, sql.NullInt32) ([]*db.Grant, error) {
	return q.grants, nil
}

func (q *userGrantGroupQueries) GetAllForumCategories(context.Context, db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
	return []*db.Forumcategory{}, nil
}

func (q *userGrantGroupQueries) SystemListLanguages(context.Context) ([]*db.Language, error) {
	return []*db.Language{}, nil
}

// TestBuildUserGrantGroupsIncludesAvailableActionsWithoutGrants ensures groups exist when user has no grants.
func TestBuildUserGrantGroupsIncludesAvailableActionsWithoutGrants(t *testing.T) {
	q := &userGrantGroupQueries{}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())

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
}
