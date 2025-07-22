package auth

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestLoginAction_NoSuchUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO login_attempts (username, ip_address) VALUES (?, ?)")).WithArgs("bob", "1.2.3.4").WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{"username": {"bob"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	loginTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "No such user") {
		t.Fatalf("body=%q", body)
	}
}

func TestLoginAction_InvalidPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "e", "7c4f29407893c334a6cb7a87bf045c0d", "md5", "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers,")).WithArgs(sql.NullString{String: "bob", Valid: true}).WillReturnRows(rows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(int32(1), sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO login_attempts (username, ip_address) VALUES (?, ?)")).WithArgs("bob", "1.2.3.4").WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{"username": {"bob"}, "password": {"wrong"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	loginTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid password") {
		t.Fatalf("body=%q", body)
	}
}

func TestLoginPageHiddenFields(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	q := dbpkg.New(db)

	req := httptest.NewRequest(http.MethodGet, "/login?code=abc&back=%2Ffoo&method=POST&data=x", nil)
	cd := common.NewCoreData(context.Background(), q)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	loginTask.Page(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "name=\"code\" value=\"abc\"") {
		t.Fatalf("missing code field: %q", body)
	}
	if !strings.Contains(body, "name=\"back\" value=\"/foo\"") {
		t.Fatalf("missing back field: %q", body)
	}
	if !strings.Contains(body, "name=\"method\" value=\"POST\"") {
		t.Fatalf("missing method field: %q", body)
	}
	if !strings.Contains(body, "name=\"data\" value=\"x\"") {
		t.Fatalf("missing data field: %q", body)
	}
}
