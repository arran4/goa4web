package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TestAdminGrantsAvailablePage(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/grants/available", nil)
	ctx := req.Context()
	cfg := config.NewRuntimeConfig()
	// Templates are in ../../core/templates relative to handlers/admin
	cfg.TemplatesDir = "../../core/templates"

	queries := &grantsPageQueries{}

	cd := common.NewCoreData(ctx, queries, cfg, common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminGrantsAvailablePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()

	// The current template implementation uses define "title" and define "body",
	// which presumably results in empty output when executed directly without a layout.
	// We expect "Available Grants" to be in the output if it was working correctly.

	expectedTitle := "Available Grants"
	if !strings.Contains(body, expectedTitle) {
		t.Errorf("Response body does not contain title '%s'", expectedTitle)
	}

	expectedContent := "Allows posting new blog entries"
	if !strings.Contains(body, expectedContent) {
		t.Errorf("Response body does not contain expected content '%s'", expectedContent)
	}

	if len(strings.TrimSpace(body)) == 0 {
		t.Errorf("Response body is empty")
	}
}
