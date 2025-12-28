package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/handlers/handlertest"
)

func TestRedirectBackPageHandlerGETAlt(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req, cd, _, cleanup := handlertest.RequestWithCoreData(t, req)
	defer cleanup()
	rr := httptest.NewRecorder()

	h := redirectBackPageHandler{BackURL: "/foo", Method: http.MethodGet}
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/foo") {
		t.Fatalf("auto refresh=%q", cd.AutoRefresh)
	}
}
