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

// SetTopicLabelsTask replaces public and private labels on a topic.
type SetTopicLabelsTask struct{ tasks.TaskString }

var (
	addTopicPublicLabelTask    = &AddTopicPublicLabelTask{TaskString: TaskAddTopicPublicLabel}
	removeTopicPublicLabelTask = &RemoveTopicPublicLabelTask{TaskString: TaskRemoveTopicPublicLabel}
	setTopicLabelsTask         = &SetTopicLabelsTask{TaskString: TaskSetTopicLabels}
)

var (
	_ tasks.Task = (*AddTopicPublicLabelTask)(nil)
	_ tasks.Task = (*RemoveTopicPublicLabelTask)(nil)
	_ tasks.Task = (*SetTopicLabelsTask)(nil)
)

// topicLabelPermissionSection returns consts.PermissionSectionPrivateForum when under /private,
// or consts.PermissionSectionForum otherwise.
func topicLabelPermissionSection(cd *common.CoreData, r *http.Request) consts.PermissionSection {
	if strings.HasPrefix(forumBasePath(cd, r), "/private") {
		return consts.PermissionSectionPrivateForum
	}
	return consts.PermissionSectionForum
}

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

	label := r.PostFormValue("label")
	if label != "" {
		isPrivate := topicLabelPermissionSection(cd, r) == consts.PermissionSectionPrivateForum
		if err := cd.AddTopicPublicLabelAction(r.Context(), common.TopicLabelParams{
			ActorID:        int32(cd.UserID),
			TopicID:        int32(topicID),
			Label:          label,
			Private:        isPrivate,
			EnforceHandler: true,
		}); err != nil {
			if _, ok := err.(common.ForumOperationForbiddenError); ok {
				return fmt.Errorf("permission denied")
			}
			if _, ok := err.(common.ForumResourceNotFoundError); ok {
				return fmt.Errorf("permission denied")
			}
			if _, ok := err.(common.ForumHandlerMismatchError); ok {
				return fmt.Errorf("permission denied")
			}
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

	label := r.PostFormValue("label")
	if label != "" {
		isPrivate := topicLabelPermissionSection(cd, r) == consts.PermissionSectionPrivateForum
		if err := cd.RemoveTopicPublicLabelAction(r.Context(), common.TopicLabelParams{
			ActorID:        int32(cd.UserID),
			TopicID:        int32(topicID),
			Label:          label,
			Private:        isPrivate,
			EnforceHandler: true,
		}); err != nil {
			if _, ok := err.(common.ForumOperationForbiddenError); ok {
				return fmt.Errorf("permission denied")
			}
			if _, ok := err.(common.ForumResourceNotFoundError); ok {
				return fmt.Errorf("permission denied")
			}
			if _, ok := err.(common.ForumHandlerMismatchError); ok {
				return fmt.Errorf("permission denied")
			}
			log.Printf("remove topic public label: %v", err)
			return fmt.Errorf("remove topic public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return topicLabelsRedirect(r)
}

func (SetTopicLabelsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, err := strconv.Atoi(vars["topic"])
	if err != nil {
		return fmt.Errorf("invalid topic id %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	pub := r.PostForm["public"]
	priv := r.PostForm["private"]

	isPrivate := topicLabelPermissionSection(cd, r) == consts.PermissionSectionPrivateForum
	if err := cd.SetTopicLabelsAction(r.Context(), common.SetTopicLabelsParams{
		ActorID:        int32(cd.UserID),
		TopicID:        int32(topicID),
		PublicLabels:   pub,
		PrivateLabels:  priv,
		Private:        isPrivate,
		EnforceHandler: true,
	}); err != nil {
		if _, ok := err.(common.ForumOperationForbiddenError); ok {
			return fmt.Errorf("permission denied")
		}
		if _, ok := err.(common.ForumResourceNotFoundError); ok {
			return fmt.Errorf("permission denied")
		}
		if _, ok := err.(common.ForumHandlerMismatchError); ok {
			return fmt.Errorf("permission denied")
		}
		log.Printf("set topic labels: %v", err)
		return fmt.Errorf("set topic labels %w", handlers.ErrRedirectOnSamePageHandler(err))
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
