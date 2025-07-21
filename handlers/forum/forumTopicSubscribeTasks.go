package forum

import (
	"fmt"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// subscribeTopicTask subscribes a user to new threads within a topic.
type subscribeTopicTask struct{ tasks.TaskString }

var subscribeTopicTaskAction = &subscribeTopicTask{TaskString: TaskSubscribeToTopic}

var _ tasks.Task = (*subscribeTopicTask)(nil)

func (subscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	pattern := topicSubscriptionPattern(int32(topicID))
	if err := queries.InsertSubscription(r.Context(), db.InsertSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: "internal"}); err != nil {
		log.Printf("insert subscription: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/forum/topic/%d", topicID), http.StatusSeeOther)
}

// unsubscribeTopicTask removes a topic subscription.
type unsubscribeTopicTask struct{ tasks.TaskString }

var unsubscribeTopicTaskAction = &unsubscribeTopicTask{TaskString: TaskUnsubscribeFromTopic}

var _ tasks.Task = (*unsubscribeTopicTask)(nil)

func (unsubscribeTopicTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	pattern := topicSubscriptionPattern(int32(topicID))
	if err := queries.DeleteSubscription(r.Context(), db.DeleteSubscriptionParams{UsersIdusers: uid, Pattern: pattern, Method: "internal"}); err != nil {
		log.Printf("delete subscription: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/forum/topic/%d", topicID), http.StatusSeeOther)
}
