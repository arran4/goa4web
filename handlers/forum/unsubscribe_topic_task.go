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
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// unsubscribeTopicTask removes a topic subscription.
type unsubscribeTopicTask struct{ tasks.TaskString }

var unsubscribeTopicTaskAction = &unsubscribeTopicTask{TaskString: TaskUnsubscribeFromTopic}

var _ tasks.Task = (*unsubscribeTopicTask)(nil)

func (unsubscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	pattern := topicSubscriptionPattern(int32(topicID))
	if err := queries.DeleteSubscriptionForSubscriber(r.Context(), db.DeleteSubscriptionForSubscriberParams{SubscriberID: uid, Pattern: pattern, Method: "internal"}); err != nil {
		log.Printf("delete subscription: %v", err)
		return fmt.Errorf("delete subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d", topicID))
}
