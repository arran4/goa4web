package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func TestMarkThreadReadTaskRedirect(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/2")
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/2" {
		t.Fatalf("redirect %q, want /private/topic/1/thread/2", rdh.TargetURL)
	}
}

func TestMarkThreadReadTaskRefererFallback(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/private/topic/1/thread/1")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/1" {
		t.Fatalf("redirect %q, want /private/topic/1/thread/1", rdh.TargetURL)
	}
}

func TestSetLabelsTaskAddsInverseLabels(t *testing.T) {
	q := &db.QuerierStub{
		ContentPublicLabelsRows: map[string][]*db.ListContentPublicLabelsRow{
			"thread:1": {},
		},
		ContentLabelStatusRows: map[string][]*db.ListContentLabelStatusRow{
			"thread:1": {},
		},
		ContentPrivateLabelsRows: map[string][]*db.ListContentPrivateLabelsRow{
			"thread:1:1": {},
		},
	}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	form := url.Values{}
	form.Set("task", string(TaskSetLabels))
	req := httptest.NewRequest(http.MethodPost, "/forum/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	setLabelsTask.Action(rr, req)

	if got := q.AddContentPrivateLabelCalls; len(got) != 2 {
		t.Fatalf("inverse private label inserts %+v, want two entries", got)
	} else {
		if got[0].Label != "new" || !got[0].Invert || got[1].Label != "unread" || !got[1].Invert {
			t.Fatalf("inverse private label args %+v, want new/unread inverted", got)
		}
	}
}

func TestSetLabelsTaskUpdatesSpecialLabels(t *testing.T) {
	q := &db.QuerierStub{}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	form.Set("task", string(TaskMarkThreadRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	// Execute the mark-as-read task, which should upsert the inverse labels.
	_ = MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)

	if got := q.AddContentPrivateLabelCalls; len(got) != 2 {
		t.Fatalf("inverse private label inserts %+v, want new and unread", got)
	} else {
		if got[0].Label != "new" || !got[0].Invert || got[1].Label != "unread" || !got[1].Invert {
			t.Fatalf("inverse private label args %+v, want inverted new/unread", got)
		}
	}
}

func TestMarkThreadReadTaskRedirectWithThread(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	form := url.Values{}
	form.Set("redirect", "/private/topic/1/thread/3")
	form.Set("task", string(TaskMarkThreadRead))
	req := httptest.NewRequest(http.MethodPost, "/private/topic/1/thread/1/labels", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})

	res := MarkThreadReadTask{}.Action(httptest.NewRecorder(), req)
	rdh, ok := res.(handlers.RefreshDirectHandler)
	if !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if rdh.TargetURL != "/private/topic/1/thread/3" {
		t.Fatalf("expected redirect to /private/topic/1/thread/3 got %s", rdh.TargetURL)
	}
}
