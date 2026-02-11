package admin

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathSaveTemplateTaskRecordsOverride(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), q, cfg)

	req := httptest.NewRequest("POST", "/admin/email/template", strings.NewReader("name=reply.gotxt&body=custom"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := saveTemplateTask.Action(httptest.NewRecorder(), req)
	if len(q.AdminSetTemplateOverrideCalls) != 1 {
		t.Fatalf("expected template override call, got %d", len(q.AdminSetTemplateOverrideCalls))
	}
	call := q.AdminSetTemplateOverrideCalls[0]
	if call.Name != "reply.gotxt" || call.Body != "custom" {
		t.Fatalf("unexpected params: %#v", call)
	}
	if _, ok := res.(handlers.RefreshDirectHandler); !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
}
