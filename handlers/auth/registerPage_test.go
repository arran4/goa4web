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
	queries := testhelpers.NewQuerierStub()
	queries.SystemGetUserByUsernameErr = sql.ErrNoRows
	queries.SystemGetUserByEmailErr = sql.ErrNoRows
	queries.SystemInsertUserReturns = 1
	queries.InsertUserEmailErr = nil
	queries.InsertPasswordErr = nil

	form := url.Values{"username": {"alice"}, "password": {"pw"}, "email": {"alice@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
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
}
