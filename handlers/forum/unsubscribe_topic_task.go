package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// unsubscribeTopicTask removes a topic subscription.
type unsubscribeTopicTask struct{ tasks.TaskString }

var unsubscribeTopicTaskAction = &unsubscribeTopicTask{TaskString: TaskUnsubscribeFromTopic}

// UnsubscribeTopicTaskHandler removes a topic subscription. Exported for reuse.
var UnsubscribeTopicTaskHandler = unsubscribeTopicTaskAction

var _ tasks.Task = (*unsubscribeTopicTask)(nil)

func (unsubscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := cd.UnsubscribeTopic(cd.UserID, int32(topicID)); err != nil {
		log.Printf("delete subscription: %v", err)
		return fmt.Errorf("delete subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	return handlers.RedirectHandler(fmt.Sprintf("%s/topic/%d", base, topicID))
}
