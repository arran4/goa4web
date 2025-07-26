package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TestVerifyAccess(t *testing.T) {
	h := VerifyAccess(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "administrator")

	req := httptest.NewRequest("GET", "/", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	cd.SetRoles([]string{"anonymous"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Code)
	}

	cd.SetRoles([]string{"administrator"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr = httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
	}
}
