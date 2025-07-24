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
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// subscribeTopicTask subscribes a user to new threads within a topic.
type subscribeTopicTask struct{ tasks.TaskString }

var subscribeTopicTaskAction = &subscribeTopicTask{TaskString: TaskSubscribeToTopic}

var _ tasks.Task = (*subscribeTopicTask)(nil)

func (subscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData).GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	pattern := topicSubscriptionPattern(int32(topicID))
	if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: "internal"}); err != nil {
		log.Printf("insert subscription: %v", err)
		return fmt.Errorf("insert subscription %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d", topicID))
}
