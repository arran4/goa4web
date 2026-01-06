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
)

type userLangPageQueries struct {
	db.Querier
	languages []*db.Language
	userLangs []*db.UserLanguage
	pref      *db.Preference
	prefErr   error
}

func (q *userLangPageQueries) GetPreferenceForLister(_ context.Context, _ int32) (*db.Preference, error) {
	if q.prefErr != nil {
		return nil, q.prefErr
	}
	return q.pref, nil
}

func (q *userLangPageQueries) GetUserLanguages(_ context.Context, _ int32) ([]*db.UserLanguage, error) {
	return q.userLangs, nil
}

func (q *userLangPageQueries) SystemListLanguages(_ context.Context) ([]*db.Language, error) {
	return q.languages, nil
}

func (q *userLangPageQueries) GetPermissionsByUserID(_ context.Context, _ int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return []*db.GetPermissionsByUserIDRow{}, nil
}

func TestUserLangPage(t *testing.T) {
	queries := &userLangPageQueries{
		languages: []*db.Language{
			{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}},
		},
		prefErr: sql.ErrNoRows,
	}

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
}
