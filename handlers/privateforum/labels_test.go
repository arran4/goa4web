package privateforum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/gorilla/mux"
)

func TestPrivateLabelRoutes(t *testing.T) {
	r := mux.NewRouter()
	nav := navpkg.NewRegistry()
	RegisterRoutes(r, config.NewRuntimeConfig(), nav)

	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	cd.ForumBasePath = "/private"

	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/labels", strings.NewReader("task="+url.QueryEscape(string(forumhandlers.TaskMarkThreadRead))))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}
