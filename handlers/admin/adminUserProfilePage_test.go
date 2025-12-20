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

type userProfileQueries struct {
	db.Querier
	userID int32
	user   *db.SystemGetUserByIDRow
}

func (q *userProfileQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func TestAdminUserProfilePage_UserFound(t *testing.T) {
	t.Skip("templates not available")
	userID := 22
	queries := &userProfileQueries{
		userID: int32(userID),
		user: &db.SystemGetUserByIDRow{
			Idusers:                int32(userID),
			Email:                  sql.NullString{String: "u@example.com", Valid: true},
			Username:               sql.NullString{String: "u", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/user/%d", userID), nil)
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminUserProfilePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
