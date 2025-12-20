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
)

type requestPageQueries struct {
	db.Querier
	requestID int32
	request   *db.AdminRequestQueue
}

func (q *requestPageQueries) AdminGetRequestByID(_ context.Context, id int32) (*db.AdminRequestQueue, error) {
	if id != q.requestID {
		return nil, fmt.Errorf("unexpected request id: %d", id)
	}
	return q.request, nil
}

func TestAdminRequestPage_RequestFound(t *testing.T) {
	requestID := 5
	queries := &requestPageQueries{
		requestID: int32(requestID),
		request: &db.AdminRequestQueue{
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
		},
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/request/%d", requestID), nil)
	req = mux.SetURLVars(req, map[string]string{"request": strconv.Itoa(requestID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminRequestPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if cr := cd.CurrentRequest(); cr == nil || cr.ID != int32(requestID) {
		t.Fatalf("current request id=%v", cr)
	}
}
