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

// AddPublicLabelTask adds a public label to a topic.
type AddPublicLabelTask struct{ tasks.TaskString }

// RemovePublicLabelTask removes a public label from a topic.
type RemovePublicLabelTask struct{ tasks.TaskString }

// AddPrivateLabelTask adds a private label for the current user.
type AddPrivateLabelTask struct{ tasks.TaskString }

// RemovePrivateLabelTask removes a private label for the current user.
type RemovePrivateLabelTask struct{ tasks.TaskString }

var (
	addPublicLabelTask     = &AddPublicLabelTask{TaskString: TaskAddPublicLabel}
	removePublicLabelTask  = &RemovePublicLabelTask{TaskString: TaskRemovePublicLabel}
	addPrivateLabelTask    = &AddPrivateLabelTask{TaskString: TaskAddPrivateLabel}
	removePrivateLabelTask = &RemovePrivateLabelTask{TaskString: TaskRemovePrivateLabel}
	addAuthorLabelTask     = &AddAuthorLabelTask{TaskString: TaskAddAuthorLabel}
	removeAuthorLabelTask  = &RemoveAuthorLabelTask{TaskString: TaskRemoveAuthorLabel}
	markTopicReadTask      = &MarkTopicReadTask{TaskString: TaskMarkTopicRead}
	setLabelsTask          = &SetLabelsTask{TaskString: TaskSetLabels}
)

// Exported task handlers for reuse outside the forum package.
var (
	AddPublicLabelTaskHandler     = addPublicLabelTask
	RemovePublicLabelTaskHandler  = removePublicLabelTask
	AddPrivateLabelTaskHandler    = addPrivateLabelTask
	RemovePrivateLabelTaskHandler = removePrivateLabelTask
	AddAuthorLabelTaskHandler     = addAuthorLabelTask
	RemoveAuthorLabelTaskHandler  = removeAuthorLabelTask
	MarkTopicReadTaskHandler      = markTopicReadTask
	SetLabelsTaskHandler          = setLabelsTask
)

// labelsRedirect determines the page to return to after processing a label task.
// It first checks the "back" form value, then falls back to the Referer header
// before defaulting to the topic page when neither is present.
func labelsRedirect(r *http.Request) handlers.RefreshDirectHandler {
	tgt := r.PostFormValue("back")
	if tgt == "" {
		tgt = r.Header.Get("Referer")
	}
	if tgt == "" {
		tgt = strings.TrimSuffix(r.URL.Path, "/labels")
	}
	return handlers.RefreshDirectHandler{TargetURL: tgt}
}

var (
	_ tasks.Task = (*AddPublicLabelTask)(nil)
	_ tasks.Task = (*RemovePublicLabelTask)(nil)
	_ tasks.Task = (*AddPrivateLabelTask)(nil)
	_ tasks.Task = (*RemovePrivateLabelTask)(nil)
	_ tasks.Task = (*AddAuthorLabelTask)(nil)
	_ tasks.Task = (*RemoveAuthorLabelTask)(nil)
	_ tasks.Task = (*MarkTopicReadTask)(nil)
	_ tasks.Task = (*SetLabelsTask)(nil)
)

func (AddPublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddTopicPublicLabel(int32(topicID), label); err != nil {
			log.Printf("add public label: %v", err)
			return fmt.Errorf("add public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return labelsRedirect(r)
}

func (RemovePublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveTopicPublicLabel(int32(topicID), label); err != nil {
			log.Printf("remove public label: %v", err)
			return fmt.Errorf("remove public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return labelsRedirect(r)
}

func (AddPrivateLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddTopicPrivateLabel(int32(topicID), label); err != nil {
			log.Printf("add private label: %v", err)
			return fmt.Errorf("add private label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return labelsRedirect(r)
}

func (RemovePrivateLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveTopicPrivateLabel(int32(topicID), label); err != nil {
			log.Printf("remove private label: %v", err)
			return fmt.Errorf("remove private label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return labelsRedirect(r)
}

// AddAuthorLabelTask adds an author-only label to a topic.
type AddAuthorLabelTask struct{ tasks.TaskString }

// RemoveAuthorLabelTask removes an author-only label from a topic.
type RemoveAuthorLabelTask struct{ tasks.TaskString }

func (AddAuthorLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddTopicAuthorLabel(int32(topicID), label); err != nil {
			log.Printf("add author label: %v", err)
			return fmt.Errorf("add author label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return labelsRedirect(r)
}

func (RemoveAuthorLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveTopicAuthorLabel(int32(topicID), label); err != nil {
			log.Printf("remove author label: %v", err)
			return fmt.Errorf("remove author label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return labelsRedirect(r)
}

// MarkTopicReadTask clears the special new/unread flags for a topic.
type MarkTopicReadTask struct{ tasks.TaskString }

func (MarkTopicReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetTopicPrivateLabelStatus(int32(topicID), false, false); err != nil {
		log.Printf("mark read: %v", err)
		return fmt.Errorf("mark read %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

  target := r.PostFormValue("redirect")
	if target == "" {
		target = r.Header.Get("Referer")
	}
	return handlers.RefreshDirectHandler{TargetURL: target}
}

// SetLabelsTask replaces public and private labels on a topic.
type SetLabelsTask struct{ tasks.TaskString }

func (SetLabelsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	pub := r.PostForm["public"]
	priv := r.PostForm["private"]
	// Special inverse private labels: show unless stored in the database.
	inverse := map[string]bool{"new": false, "unread": false}
	filteredPriv := make([]string, 0, len(priv))
	for _, l := range priv {
		if _, ok := inverse[l]; ok {
			inverse[l] = true
			continue
		}
		filteredPriv = append(filteredPriv, l)
	}
	if err := cd.SetTopicPublicLabels(int32(topicID), pub); err != nil {
		log.Printf("set public labels: %v", err)
		return fmt.Errorf("set public labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetTopicPrivateLabels(int32(topicID), filteredPriv); err != nil {
		log.Printf("set private labels: %v", err)
		return fmt.Errorf("set private labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

  if err := cd.SetTopicPrivateLabelStatus(int32(topicID), inverse["new"], inverse["unread"]); err != nil {
		log.Printf("set private label status: %v", err)
		return fmt.Errorf("set private label status %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return labelsRedirect(r)
}
