package forumcommon

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type ConfiguredTask interface {
	tasks.Task
	tasks.TaskMatcher
}

// subscribeTopicTask subscribes a user to new threads within a topic.
type subscribeTopicTask struct {
	tasks.TaskString
	ctx *ForumContext
}

// SubscribeTopicTask returns a configured task handler for subscribing to a topic.
func (f *ForumContext) SubscribeTopicTask() ConfiguredTask {
	return &subscribeTopicTask{
		TaskString: TaskSubscribeToTopic,
		ctx:        f,
	}
}

func (t *subscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	_, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.SubscribeTopic(cd.UserID, int32(topicID)); err != nil {
		log.Printf("insert subscription: %v", err)
		return fmt.Errorf("insert subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("%s/topic/%d", t.ctx.BasePath, topicID))
}

// unsubscribeTopicTask removes a topic subscription.
type unsubscribeTopicTask struct {
	tasks.TaskString
	ctx *ForumContext
}

// UnsubscribeTopicTask returns a configured task handler for unsubscribing from a topic.
func (f *ForumContext) UnsubscribeTopicTask() ConfiguredTask {
	return &unsubscribeTopicTask{
		TaskString: TaskUnsubscribeFromTopic,
		ctx:        f,
	}
}

func (t *unsubscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	_, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.UnsubscribeTopic(cd.UserID, int32(topicID)); err != nil {
		log.Printf("delete subscription: %v", err)
		return fmt.Errorf("delete subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("%s/topic/%d", t.ctx.BasePath, topicID))
}
