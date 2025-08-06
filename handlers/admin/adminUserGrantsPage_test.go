package admin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminUserGrantsPage_UserIDInjected(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	userID := 3
	mock.ExpectQuery("SELECT").WithArgs(int32(userID)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(userID, "", "u", nil))
	mock.ExpectQuery("SELECT").WithArgs(int32(userID)).WillReturnRows(sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "role_id", "name"}))
	mock.ExpectQuery("SELECT").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "title", "description"}))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}))

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/user/%d/grants", userID), nil)
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminUserGrantsPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
