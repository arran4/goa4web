package privateforum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/gorilla/mux"
)

func TestPrivateLabelRoutes(t *testing.T) {
	r := mux.NewRouter()
	nav := navpkg.NewRegistry()
	RegisterRoutes(r, config.NewRuntimeConfig(), nav)

	t.Run("uses redirect parameter", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
		cd.UserID = 1
		cd.ForumBasePath = "/private"

		body := "task=" + url.QueryEscape(string(forumcommon.TaskMarkThreadRead)) + "&redirect=" + url.QueryEscape("/private/topic/1/thread/2")
		req := httptest.NewRequest(http.MethodPost, "/private/thread/1/labels", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", rr.Code)
		}
		if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/private/topic/1/thread/2") {
			t.Fatalf("auto refresh=%q", cd.AutoRefresh)
		}
	})

	t.Run("falls back without redirect parameter", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
		cd.UserID = 1
		cd.ForumBasePath = "/private"

		req := httptest.NewRequest(http.MethodPost, "/private/thread/1/labels", strings.NewReader("task="+url.QueryEscape(string(forumcommon.TaskMarkThreadRead))))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", rr.Code)
		}
		if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/private/thread/1") {
			t.Fatalf("auto refresh=%q", cd.AutoRefresh)
		}
	})
}
