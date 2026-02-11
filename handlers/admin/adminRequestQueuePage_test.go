package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathAdminRequestPage_RequestFound(t *testing.T) {
	requestID := 5
	queries := testhelpers.NewQuerierStub()
	queries.AdminGetRequestByIDFn = func(_ context.Context, id int32) (*db.AdminRequestQueue, error) {
		if id != int32(requestID) {
			return nil, fmt.Errorf("unexpected request id: %d", id)
		}
		return &db.AdminRequestQueue{
			ID:             int32(requestID),
			UsersIdusers:   7,
			ChangeTable:    "tbl",
			ChangeField:    "fld",
			ChangeRowID:    0,
			ChangeValue:    sql.NullString{},
			ContactOptions: sql.NullString{},
			Status:         "pending",
			CreatedAt:      time.Now(),
			ActedAt:        sql.NullTime{},
		}, nil
	}
	queries.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
		Idusers:  7,
		Username: sql.NullString{String: "testuser", Valid: true},
	}
	queries.AdminListUserEmailsReturns = []*db.UserEmail{}
	queries.AdminListRequestCommentsReturns = []*db.AdminRequestComment{}

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/request/%d", requestID), nil)
	req = mux.SetURLVars(req, map[string]string{"request": strconv.Itoa(requestID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	(&AdminRequestPage{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if cr := cd.CurrentRequest(); cr == nil || cr.ID != int32(requestID) {
		t.Fatalf("current request id=%v", cr)
	}
}

func TestHappyPathAdminRequestPage_UserEmailsRequest(t *testing.T) {
	requestID := 5
	userID := int32(7)
	queries := testhelpers.NewQuerierStub()
	queries.AdminGetRequestByIDFn = func(_ context.Context, id int32) (*db.AdminRequestQueue, error) {
		if id != int32(requestID) {
			return nil, fmt.Errorf("unexpected request id: %d", id)
		}
		return &db.AdminRequestQueue{
			ID:             int32(requestID),
			UsersIdusers:   userID,
			ChangeTable:    "user_emails",
			ChangeField:    "email",
			ChangeRowID:    userID,
			ChangeValue:    sql.NullString{String: "new@example.com", Valid: true},
			ContactOptions: sql.NullString{String: "new@example.com", Valid: true},
			Status:         "pending",
			CreatedAt:      time.Now(),
			ActedAt:        sql.NullTime{},
		}, nil
	}
	queries.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
		Idusers:  userID,
		Username: sql.NullString{String: "testuser", Valid: true},
	}
	queries.AdminListUserEmailsReturns = []*db.UserEmail{}
	queries.AdminListRequestCommentsReturns = []*db.AdminRequestComment{}

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/request/%d", requestID), nil)
	req = mux.SetURLVars(req, map[string]string{"request": strconv.Itoa(requestID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	(&AdminRequestPage{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if cr := cd.CurrentRequest(); cr == nil || cr.ID != int32(requestID) {
		t.Fatalf("current request id=%v", cr)
	}
	// Check if CurrentProfileUserID was set correctly
	if pid := cd.CurrentProfileUserID(); pid != userID {
		t.Fatalf("CurrentProfileUserID not set: got %d, want %d", pid, userID)
	}
}
