package admin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/admin"
)

// TestHappyPathRoleGrantsEditorJSRoute ensures the role grants editor script is served.
func TestHappyPathRoleGrantsEditorJSRoute(t *testing.T) {
	h := admin.New()
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	h.RegisterRoutes(ar, &config.RuntimeConfig{})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/admin/role-grants-editor.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/javascript" {
		t.Fatalf("content-type=%q", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("empty body")
	}

	headReq := httptest.NewRequest(http.MethodHead, "http://example.com/admin/role-grants-editor.js", nil)
	headRec := httptest.NewRecorder()
	r.ServeHTTP(headRec, headReq)
	if headRec.Code != http.StatusOK {
		t.Fatalf("head status=%d", headRec.Code)
	}
	if headRec.Body.Len() != 0 {
		t.Fatalf("head body length=%d", headRec.Body.Len())
	}

	optReq := httptest.NewRequest(http.MethodOptions, "http://example.com/admin/role-grants-editor.js", nil)
	optRec := httptest.NewRecorder()
	r.ServeHTTP(optRec, optReq)
	if optRec.Code != http.StatusOK {
		t.Fatalf("options status=%d", optRec.Code)
	}
}
