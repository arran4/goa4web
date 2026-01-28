package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
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

// subscribeTopicTask subscribes a user to new threads within a topic.
type subscribeTopicTask struct{ tasks.TaskString }

var subscribeTopicTaskAction = &subscribeTopicTask{TaskString: forumcommon.TaskSubscribeToTopic}

// SubscribeTopicTaskHandler subscribes a user to a topic. Exported for reuse.
var SubscribeTopicTaskHandler = subscribeTopicTaskAction

var _ tasks.Task = (*subscribeTopicTask)(nil)

func (subscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	return handlers.RedirectHandler(fmt.Sprintf("%s/topic/%d", base, topicID))
}
