package admin

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathBuildUserGrantGroups(t *testing.T) {
	t.Run("Includes Available Actions Without Grants", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.ListGrantsByUserIDReturns = []*db.Grant{}
		q.GetAllForumCategoriesReturns = []*db.Forumcategory{}
		q.SystemListLanguagesReturns = []*db.Language{}

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
	})
}
