package auth

import (
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
		rr := httptest.NewRecorder()
		RegisterActionPage(rr, req)
		if rr.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status=%d", c.name, rr.Result().StatusCode)
		}
	}
}
