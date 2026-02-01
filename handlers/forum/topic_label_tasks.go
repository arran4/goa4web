package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AddTopicPublicLabelTask adds a public label to a topic.
type AddTopicPublicLabelTask struct{ tasks.TaskString }

// RemoveTopicPublicLabelTask removes a public label from a topic.
type RemoveTopicPublicLabelTask struct{ tasks.TaskString }

var (
	addTopicPublicLabelTask    = &AddTopicPublicLabelTask{TaskString: TaskAddTopicPublicLabel}
	removeTopicPublicLabelTask = &RemoveTopicPublicLabelTask{TaskString: TaskRemoveTopicPublicLabel}
)

var (
	_ tasks.Task = (*AddTopicPublicLabelTask)(nil)
	_ tasks.Task = (*RemoveTopicPublicLabelTask)(nil)
)

func (AddTopicPublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		return fmt.Errorf("invalid topic id %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if !cd.IsAdmin() {
		canLabel, err := UserCanLabelTopic(r.Context(), cd.Queries(), "forum", int32(topicID), int32(cd.UserID))
		if err != nil {
			log.Printf("UserCanLabelTopic error: %v", err)
			return fmt.Errorf("check permission fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if !canLabel {
			return fmt.Errorf("permission denied")
		}
	}

	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddTopicPublicLabel(int32(topicID), label); err != nil {
			log.Printf("add topic public label: %v", err)
			return fmt.Errorf("add topic public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return topicLabelsRedirect(r)
}

func (RemoveTopicPublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		return fmt.Errorf("invalid topic id %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if !cd.IsAdmin() {
		canLabel, err := UserCanLabelTopic(r.Context(), cd.Queries(), "forum", int32(topicID), int32(cd.UserID))
		if err != nil {
			log.Printf("UserCanLabelTopic error: %v", err)
			return fmt.Errorf("check permission fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if !canLabel {
			return fmt.Errorf("permission denied")
		}
	}

	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveTopicPublicLabel(int32(topicID), label); err != nil {
			log.Printf("remove topic public label: %v", err)
			return fmt.Errorf("remove topic public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return topicLabelsRedirect(r)
}

func topicLabelsRedirect(r *http.Request) handlers.RefreshDirectHandler {
	tgt := r.PostFormValue("back")
	if tgt == "" {
		tgt = r.Header.Get("Referer")
	}
	if tgt == "" {
		// Fallback
		tgt = strings.TrimSuffix(r.URL.Path, "/labels")
	}
	return handlers.RefreshDirectHandler{TargetURL: tgt}
}
