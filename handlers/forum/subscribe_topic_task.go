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

// subscribeTopicTask subscribes a user to new threads within a topic.
type subscribeTopicTask struct{ tasks.TaskString }

var subscribeTopicTaskAction = &subscribeTopicTask{TaskString: TaskSubscribeToTopic}

var _ tasks.Task = (*subscribeTopicTask)(nil)

func (subscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	_, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.SubscribeTopic(int32(topicID)); err != nil {
		log.Printf("insert subscription: %v", err)
		return fmt.Errorf("insert subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d", topicID))
}
