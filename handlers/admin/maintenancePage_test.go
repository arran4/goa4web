package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminMaintenancePageListsTopics(t *testing.T) {
	q := &common.QuerierFake{
		AdminListTopicsWithUserGrantsNoRolesRows: []*db.AdminListTopicsWithUserGrantsNoRolesRow{
			{Idforumtopic: 1, Title: sql.NullString{String: "Topic One", Valid: true}},
			{Idforumtopic: 2, Title: sql.NullString{String: "Topic Two", Valid: true}},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/maintenance", nil)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	(&AdminMaintenancePage{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Topic One (ID 1)") || !strings.Contains(body, "Topic Two (ID 2)") {
		t.Fatalf("expected topic titles in response, got %s", body)
	}

	if len(q.AdminListTopicsWithUserGrantsNoRolesCalls) != 1 {
		t.Fatalf("expected 1 topic listing call, got %d", len(q.AdminListTopicsWithUserGrantsNoRolesCalls))
	}
	if include, ok := q.AdminListTopicsWithUserGrantsNoRolesCalls[0].(bool); !ok || !include {
		t.Fatalf("expected includeAdmin=true, got %#v", q.AdminListTopicsWithUserGrantsNoRolesCalls[0])
	}
}
