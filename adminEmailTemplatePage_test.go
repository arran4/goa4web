package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	os.Unsetenv("EMAIL_PROVIDER")
	appRuntimeConfig.EmailProvider = ""

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{UserID: 1})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminEmailTemplateTestActionPage(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(errMailNotConfigured)
	if loc := rr.Header().Get("Location"); !strings.Contains(loc, want) {
		t.Fatalf("location=%q", loc)
	}
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "log")
	appRuntimeConfig.EmailProvider = "log"
	defer os.Unsetenv("EMAIL_PROVIDER")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	rows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "username"}).AddRow(1, "u@example.com", "", "u")
	mock.ExpectQuery("SELECT idusers, email, passwd, username FROM users WHERE idusers = ?").WithArgs(int32(1)).WillReturnRows(rows)

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{UserID: 1})
	ctx = context.WithValue(ctx, ContextValues("queries"), q)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminEmailTemplateTestActionPage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/email/template" {
		t.Fatalf("location=%q", loc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
