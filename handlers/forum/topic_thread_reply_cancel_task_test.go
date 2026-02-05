package forum

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/go-be-lazy"
	"github.com/gorilla/mux"
)

func TestTopicThreadReplyCancel_BasePath(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	cd.ForumBasePath = "/private"
	thread := &db.GetThreadLastPosterAndPermsRow{Idforumthread: 2, ForumtopicIdforumtopic: 1}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: 1}
	if _, err := cd.ForumThreadByID(2, lazy.Set(thread)); err != nil {
		t.Fatalf("set thread: %v", err)
	}
	if _, err := cd.ForumTopicByID(1, lazy.Set(topic)); err != nil {
		t.Fatalf("set topic: %v", err)
	}
	cd.SetCurrentThreadAndTopic(2, 1)
	req := httptest.NewRequest("POST", "/private/topic/1/thread/2/reply", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	cd.LoadSelectionsFromRequest(req)

	res := TopicThreadReplyCancelHandler.Action(httptest.NewRecorder(), req)
	rh, ok := res.(handlers.RedirectHandler)
	if !ok {
		t.Fatalf("expected RedirectHandler, got %T", res)
	}
	if string(rh) != "/private/topic/1/thread/2#bottom" {
		t.Fatalf("redirect=%s", rh)
	}
}
