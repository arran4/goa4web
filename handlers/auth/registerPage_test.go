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

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "/login?notice=") {
			t.Fatalf("expected body to reference login notice got %q", body)
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

	t.Run("Happy Path - Preserves Back Params", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetUserByUsernameErr = sql.ErrNoRows
		q.SystemGetUserByEmailErr = sql.ErrNoRows
		q.SystemInsertUserReturns = 2

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
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(registerTask)(rr, req)

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
	})
}
