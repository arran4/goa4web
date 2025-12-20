package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type userGrantsQueries struct {
	db.Querier
	userID int32
	user   *db.SystemGetUserByIDRow
}

func (q *userGrantsQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *userGrantsQueries) GetPermissionsByUserID(_ context.Context, id int32) ([]*db.GetPermissionsByUserIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected permissions user id: %d", id)
	}
	return []*db.GetPermissionsByUserIDRow{}, nil
}

func (q *userGrantsQueries) ListGrantsByUserID(_ context.Context, id sql.NullInt32) ([]*db.Grant, error) {
	if !id.Valid || id.Int32 != q.userID {
		return nil, fmt.Errorf("unexpected grant user id: %v", id)
	}
	return []*db.Grant{}, nil
}

func (q *userGrantsQueries) GetAllForumCategories(context.Context, db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
	return []*db.Forumcategory{}, nil
}

func (q *userGrantsQueries) SystemListLanguages(context.Context) ([]*db.Language, error) {
	return []*db.Language{}, nil
}

func TestAdminUserGrantsPage_UserIDInjected(t *testing.T) {
	userID := 3
	queries := &userGrantsQueries{
		userID: int32(userID),
		user: &db.SystemGetUserByIDRow{
			Idusers:                int32(userID),
			Username:               sql.NullString{String: "u", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/user/%d/grants", userID), nil)
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminUserGrantsPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
