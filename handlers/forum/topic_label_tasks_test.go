package forum

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
	"github.com/arran4/goa4web/handlers"
)

func TestMarkTopicReadTaskRedirect(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	form := url.Values{}
	form.Set("redirect", "/private/topic/1")
	form.Set("task", string(TaskMarkTopicRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/private/topic/1/thread/51")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()

	res := MarkTopicReadTaskHandler.Action(rr, req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1" {
		t.Fatalf("expected redirect to /private/topic/1 got %s", rdh.TargetURL)
	}
}
