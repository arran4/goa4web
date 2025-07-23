package auth

import (
	"context"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRegisterActionPageValidation(t *testing.T) {
	cases := []struct {
		name string
		form url.Values
	}{
		{"no username", url.Values{"password": {"p"}, "email": {"e@example.com"}}},
		{"no password", url.Values{"username": {"u"}, "email": {"e@example.com"}}},
		{"no email", url.Values{"username": {"u"}, "password": {"p"}}},
		{"invalid email", url.Values{"username": {"u"}, "password": {"p"}, "email": {"bad"}}},
	}
	for _, c := range cases {
		req := httptest.NewRequest("POST", "/register", strings.NewReader(c.form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, &common.CoreData{})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handlers.TaskHandler(registerTask)(rr, req)
		want := http.StatusOK
		if c.name == "invalid email" {
			want = http.StatusTemporaryRedirect
		}
		if rr.Result().StatusCode != want {
			t.Errorf("%s: status=%d", c.name, rr.Result().StatusCode)
		}
	}
}
