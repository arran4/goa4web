package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRegisterTask_Action(t *testing.T) {
	t.Run("Unhappy Path - Validation Errors", func(t *testing.T) {
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
			t.Run(c.name, func(t *testing.T) {
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
			})
		}
	})

	t.Run("Happy Path - Redirects To Login", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetUserByUsernameErr = sql.ErrNoRows
		q.SystemGetUserByEmailErr = sql.ErrNoRows
		q.SystemInsertUserReturns = 1

		form := url.Values{"username": {"alice"}, "password": {"pw"}, "email": {"alice@example.com"}}
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(registerTask)(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("status=%d", rr.Code)
		}
		location, err := url.Parse(rr.Header().Get("Location"))
		if err != nil {
			t.Fatalf("parse Location: %v", err)
		}
		if location.Path != "/login" || location.Query().Get("notice") != "approval is pending" {
			t.Fatalf("Location=%q; want login approval notice", location.String())
		}
		if strings.Contains(strings.ToLower(rr.Body.String()), "http-equiv=\"refresh\"") {
			t.Fatalf("registration rendered a meta-refresh transition: %q", rr.Body.String())
		}
		if got := rr.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
			t.Errorf("Cache-Control=%q; want no-store", got)
		}
		if got := rr.Header().Get("Cloudflare-CDN-Cache-Control"); got != "no-store" {
			t.Errorf("Cloudflare-CDN-Cache-Control=%q; want no-store", got)
		}
		if len(q.SystemInsertUserCalls) != 1 {
			t.Errorf("expected user insert")
		}
		if len(q.InsertUserEmailCalls) != 1 {
			t.Errorf("expected email insert")
		}
		if len(q.InsertPasswordCalls) != 1 {
			t.Errorf("expected password insert")
		}
	})

	t.Run("Happy Path - Preserves Safe Back Only", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetUserByUsernameErr = sql.ErrNoRows
		q.SystemGetUserByEmailErr = sql.ErrNoRows
		q.SystemInsertUserReturns = 2

		form := url.Values{
			"username": {"bob"},
			"password": {"pw"},
			"email":    {"bob@example.com"},
			"back":     {"/protected?view=full"},
		}
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(registerTask)(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		location, err := url.Parse(rr.Header().Get("Location"))
		if err != nil {
			t.Fatalf("parse Location: %v", err)
		}
		if got := location.Query().Get("back"); got != "/protected?view=full" {
			t.Errorf("back=%q; want safe local continuation with query", got)
		}
		if location.Query().Has("method") || location.Query().Has("data") {
			t.Errorf("registration redirect retained replay fields: %q", location.String())
		}
	})

	t.Run("Unhappy Path - Rejects Legacy Replay Fields", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetUserByUsernameErr = sql.ErrNoRows
		q.SystemGetUserByEmailErr = sql.ErrNoRows
		q.SystemInsertUserReturns = 3

		form := url.Values{
			"username": {"carol"},
			"password": {"pw"},
			"email":    {"carol@example.com"},
			"back":     {"/protected"},
			"method":   {"POST"},
			"data":     {"secret=must-not-leak"},
		}
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		rr := httptest.NewRecorder()
		handlers.TaskHandler(registerTask)(rr, req)

		if location := rr.Header().Get("Location"); strings.Contains(location, "method=") || strings.Contains(location, "data=") || strings.Contains(location, "secret") {
			t.Errorf("registration propagated rejected replay data in Location: %q", location)
		}
		if len(q.SystemInsertUserCalls) != 0 {
			t.Errorf("registration with legacy replay fields created %d users", len(q.SystemInsertUserCalls))
		}
	})

	for _, back := range []string{"https://evil.example/steal", "//evil.example/steal"} {
		t.Run("Unhappy Path - Sanitizes Back "+back, func(t *testing.T) {
			q := testhelpers.NewQuerierStub()
			q.SystemGetUserByUsernameErr = sql.ErrNoRows
			q.SystemGetUserByEmailErr = sql.ErrNoRows
			q.SystemInsertUserReturns = 4

			form := url.Values{
				"username": {"dave"},
				"password": {"pw"},
				"email":    {"dave@example.com"},
				"back":     {back},
			}
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			rr := httptest.NewRecorder()
			handlers.TaskHandler(registerTask)(rr, req)

			location, err := url.Parse(rr.Header().Get("Location"))
			if err != nil {
				t.Fatalf("parse Location: %v", err)
			}
			if location.Path != "/login" || location.Query().Has("back") {
				t.Errorf("Location=%q; want sanitized login redirect without back", location.String())
			}
		})
	}
}
