package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestUserEmailTestAction_NoProvider(t *testing.T) {
	os.Unsetenv("EMAIL_PROVIDER")
	req := httptest.NewRequest("POST", "/email", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Email: sql.NullString{String: "u@example.com", Valid: true}})
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailTestActionPage(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(errMailNotConfigured)
	if loc := rr.Header().Get("Location"); !strings.Contains(loc, want) {
		t.Fatalf("location=%q", loc)
	}
}

func TestUserEmailTestAction_WithProvider(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "log")
	defer os.Unsetenv("EMAIL_PROVIDER")

	req := httptest.NewRequest("POST", "/email", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Email: sql.NullString{String: "u@example.com", Valid: true}})
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailTestActionPage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/usr/email" {
		t.Fatalf("location=%q", loc)
	}
}

func TestUserEmailPage_ShowError(t *testing.T) {
	req := httptest.NewRequest("GET", "/usr/email?error=missing", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Email: sql.NullString{String: "u@example.com", Valid: true}})
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "missing") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}
