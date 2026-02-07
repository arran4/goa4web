package admin

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathBuildGrantGroups(t *testing.T) {
	t.Run("Includes Available Actions Without Grants", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.AdminListGrantsByRoleIDReturns = []*db.Grant{}
		q.GetAllForumCategoriesReturns = []*db.Forumcategory{}
		q.SystemListLanguagesReturns = []*db.Language{}

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
	})

	t.Run("Skips Invalid Item ID Grants", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.AdminListGrantsByRoleIDReturns = []*db.Grant{{
			ID:      1,
			RoleID:  sql.NullInt32{Int32: 1, Valid: true},
			Section: "forum",
			Item:    sql.NullString{String: "topic", Valid: true},
			Action:  "view",
			Active:  true,
		}}
		q.GetAllForumCategoriesReturns = []*db.Forumcategory{}
		q.SystemListLanguagesReturns = []*db.Language{}

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
	})
}
