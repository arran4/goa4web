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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminRequestPage_RequestFound(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	requestID := 5
	reqRow := sqlmock.NewRows([]string{"id", "users_idusers", "change_table", "change_field", "change_row_id", "change_value", "contact_options", "status", "created_at", "acted_at"}).
		AddRow(requestID, 7, "tbl", "fld", 0, sql.NullString{}, sql.NullString{}, "pending", time.Now(), sql.NullTime{})
	mock.ExpectQuery("admin_request_queue").WillReturnRows(reqRow)
	userRow := sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).
		AddRow(7, "u@example.com", "user", sql.NullTime{})
	mock.ExpectQuery("FROM users").WillReturnRows(userRow)
	mock.ExpectQuery("admin_request_comments").WillReturnRows(sqlmock.NewRows([]string{"id", "request_id", "comment", "created_at"}))

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/request/%d", requestID), nil)
	req = mux.SetURLVars(req, map[string]string{"request": strconv.Itoa(requestID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
