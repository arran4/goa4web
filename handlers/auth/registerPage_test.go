package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func TestRegisterActionPageValidation(t *testing.T) {
	cases := []struct {
		name string
		form url.Values
	}{
		{"no username", url.Values{"password": {"p"}, "email": {"e@example.com"}}},
		{"no password", url.Values{"username": {"u"}, "email": {"e@example.com"}}},
		{"no email", url.Values{"username": {"u"}, "password": {"p"}}},
		{"invalid email", url.Values{"username": {"u"}, "password": {"p"}, "email": {"foo@bar..com"}}},
	}
	for _, c := range cases {
		req := httptest.NewRequest("POST", "/register", strings.NewReader(c.form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handlers.TaskHandler(registerTask)(rr, req)
		want := http.StatusOK
		if c.name == "invalid email" {
			want = http.StatusSeeOther
		}
		if rr.Result().StatusCode != want {
			t.Errorf("%s: status=%d", c.name, rr.Result().StatusCode)
		}
	}
}

func TestRegisterActionRedirectsToLogin(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers,\n       username,\n       public_profile_enabled_at\nFROM users\nWHERE username = ?\n")).
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username\nFROM users u JOIN user_emails ue ON ue.user_id = u.idusers\nWHERE ue.email = ?\nLIMIT 1\n")).
		WithArgs("alice@example.com").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (username)\nVALUES (?)\n")).
		WithArgs(sql.NullString{String: "alice", Valid: true}).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_emails (user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority)\nVALUES (?, ?, ?, ?, ?, ?)\n")).
		WithArgs(int32(1), "alice@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int32(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)\nVALUES (?, ?, ?)\n")).
		WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{"username": {"alice"}, "password": {"pw"}, "email": {"alice@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(registerTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "/login?notice=") {
		t.Fatalf("expected body to reference login notice got %q", body)
	}
}

func TestRegisterActionPreservesBackParams(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	// Expect strict order of queries
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers,\n       username,\n       public_profile_enabled_at\nFROM users\nWHERE username = ?\n")).
		WithArgs(sql.NullString{String: "bob", Valid: true}).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username\nFROM users u JOIN user_emails ue ON ue.user_id = u.idusers\nWHERE ue.email = ?\nLIMIT 1\n")).
		WithArgs("bob@example.com").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (username)\nVALUES (?)\n")).
		WithArgs(sql.NullString{String: "bob", Valid: true}).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_emails (user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority)\nVALUES (?, ?, ?, ?, ?, ?)\n")).
		WithArgs(int32(2), "bob@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int32(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)\nVALUES (?, ?, ?)\n")).
		WithArgs(int32(2), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	form := url.Values{
		"username": {"bob"},
		"password": {"pw"},
		"email":    {"bob@example.com"},
		"back":     {"/protected"},
		"method":   {"POST"},
		"data":     {"foo=bar"},
	}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(registerTask)(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Logf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	// Verify parameters are in the redirect URL
	expected := "back=%2Fprotected"
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
	expected = "method=POST"
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
	expected = "data=foo%3Dbar"
	if !strings.Contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}
