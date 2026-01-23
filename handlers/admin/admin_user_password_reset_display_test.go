package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserPasswordResetIncludesPasswordInMessages(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
		Idusers:  42,
		Username: sql.NullString{String: "target", Valid: true},
	}
	qs.SystemDeletePasswordResetsByUserResult = db.FakeSQLResult{RowsAffectedValue: 1}

	req := httptest.NewRequest(http.MethodPost, "/admin/user/42/reset", nil)
	req = mux.SetURLVars(req, map[string]string{"user": "42"})
	// Mock host for AbsoluteURL
	req.Host = "example.com"

	cfg := config.NewRuntimeConfig()
	cfg.TemplatesDir = "../../core/templates"

	cd := common.NewCoreData(req.Context(), qs, cfg)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := userPasswordResetTask.Action(rr, req)

	handler, ok := result.(http.Handler)
	if !ok {
		t.Fatalf("Result is not http.Handler")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Recovered from panic during template rendering: %v", r)
		}
	}()

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	if !strings.Contains(body, "Password reset to:") {
		t.Fatalf("Did not find password in output. Body snippet: %s", body[strings.Index(body, "<main"):strings.Index(body, "</main>")+7])
	}
}
