package user

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

func setupRequest(t *testing.T, path string, userID int) (*http.Request, sqlmock.Sqlmock, *common.CoreData) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	mock.MatchExpectationsInOrder(false)
	req := httptest.NewRequest("GET", fmt.Sprintf(path, userID), nil)
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return req, mock, cd
}

func TestAdminUserPermissionsPage_UserIDInjected(t *testing.T) {
	req, mock, _ := setupRequest(t, "/admin/user/%d/permissions", 2)
	mock.ExpectQuery("SELECT").WithArgs(int32(2)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(2, "", "u", nil))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}))
	mock.ExpectQuery("SELECT").WithArgs(int32(2)).WillReturnRows(sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "role_id", "name"}))
	rr := httptest.NewRecorder()
	adminUserPermissionsPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdminUserDisableConfirmPage_UserIDInjected(t *testing.T) {
	req, mock, _ := setupRequest(t, "/admin/user/%d/disable", 5)
	mock.ExpectQuery("SELECT").WithArgs(int32(5)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(5, "", "u", nil))
	rr := httptest.NewRecorder()
	adminUserDisableConfirmPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdminUserEditFormPage_UserIDInjected(t *testing.T) {
	req, mock, _ := setupRequest(t, "/admin/user/%d/edit", 7)
	mock.ExpectQuery("SELECT").WithArgs(int32(7)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(7, "", "u", nil))
	rr := httptest.NewRecorder()
	adminUserEditFormPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
