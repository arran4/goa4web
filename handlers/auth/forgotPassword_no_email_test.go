package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestForgotPasswordNoEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("GetUserByUsername").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "", "u"))

	cd := common.NewCoreData(context.Background(), q)
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"u"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	forgotPasswordTask.Action(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEmailAssociationRequestTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("GetUserByUsername").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "", "u"))
	mock.ExpectExec("INSERT INTO admin_request_queue").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO admin_request_comments").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO admin_user_comments").WillReturnResult(sqlmock.NewResult(1, 1))

	cd := common.NewCoreData(context.Background(), q)
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"u"}, "email": {"a@test.com"}, "reason": {"help"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	emailAssociationRequestTask.Action(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
