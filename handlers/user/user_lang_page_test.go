package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserLangPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemListLanguagesReturns = []*db.Language{
			{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}},
		}

		queries.GetUserLanguagesFn = func(ctx context.Context, id int32) ([]*db.UserLanguage, error) {
			return nil, nil // Or mocked langs
		}

		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return nil, sql.ErrNoRows
		}

		queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}

		req := httptest.NewRequest("GET", "/usr/lang", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		userLangPage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		// simple check for content
		if !strings.Contains(rr.Body.String(), "Save languages") {
			t.Errorf("Expected body to contain 'Save languages', got %s", rr.Body.String())
		}
	})
}
