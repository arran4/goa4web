package forum

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

// unsubscribeTopicTask removes a topic subscription.
type unsubscribeTopicTask struct{ tasks.TaskString }

var unsubscribeTopicTaskAction = &unsubscribeTopicTask{TaskString: TaskUnsubscribeFromTopic}

var _ tasks.Task = (*unsubscribeTopicTask)(nil)

func (unsubscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	_, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.UnsubscribeTopic(int32(topicID)); err != nil {
		log.Printf("delete subscription: %v", err)
		return fmt.Errorf("delete subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d", topicID))
}
