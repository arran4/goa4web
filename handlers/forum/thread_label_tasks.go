package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AddPublicLabelTask adds a public label to a thread.
type AddPublicLabelTask struct{ tasks.TaskString }

// RemovePublicLabelTask removes a public label from a thread.
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
	markThreadReadTask     = &MarkThreadReadTask{TaskString: TaskMarkThreadRead}
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
	MarkThreadReadTaskHandler     = markThreadReadTask
	SetLabelsTaskHandler          = setLabelsTask
)

func (AddPublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddThreadPublicLabel(int32(threadID), label); err != nil {
			log.Printf("add public label: %v", err)
			return fmt.Errorf("add public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

func (RemovePublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveThreadPublicLabel(int32(threadID), label); err != nil {
			log.Printf("remove public label: %v", err)
			return fmt.Errorf("remove public label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

func (AddPrivateLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddThreadPrivateLabel(int32(threadID), label); err != nil {
			log.Printf("add private label: %v", err)
			return fmt.Errorf("add private label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

func (RemovePrivateLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveThreadPrivateLabel(int32(threadID), label); err != nil {
			log.Printf("remove private label: %v", err)
			return fmt.Errorf("remove private label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

// AddAuthorLabelTask adds an author-only label to a thread.
type AddAuthorLabelTask struct{ tasks.TaskString }

// RemoveAuthorLabelTask removes an author-only label from a thread.
type RemoveAuthorLabelTask struct{ tasks.TaskString }

func (AddAuthorLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.AddThreadAuthorLabel(int32(threadID), label); err != nil {
			log.Printf("add author label: %v", err)
			return fmt.Errorf("add author label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

func (RemoveAuthorLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	label := r.PostFormValue("label")
	if label != "" {
		if err := cd.RemoveThreadAuthorLabel(int32(threadID), label); err != nil {
			log.Printf("remove author label: %v", err)
			return fmt.Errorf("remove author label %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

// MarkThreadReadTask clears the special new/unread flags for a thread.
type MarkThreadReadTask struct{ tasks.TaskString }

func (MarkThreadReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := cd.SetThreadPrivateLabelStatus(int32(threadID), false, false); err != nil {
		log.Printf("mark read: %v", err)
		return fmt.Errorf("mark read %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}

// SetLabelsTask replaces public and private labels on a thread.
type SetLabelsTask struct{ tasks.TaskString }

func (SetLabelsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["topic"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	pub := r.PostForm["public"]
	priv := r.PostForm["private"]
	filteredPriv := make([]string, 0, len(priv))
	for _, l := range priv {
		if l != "new" && l != "unread" {
			filteredPriv = append(filteredPriv, l)
		}
	}
	if err := cd.SetThreadPublicLabels(int32(threadID), pub); err != nil {
		log.Printf("set public labels: %v", err)
		return fmt.Errorf("set public labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetThreadPrivateLabels(int32(threadID), filteredPriv); err != nil {
		log.Printf("set private labels: %v", err)
		return fmt.Errorf("set private labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: r.Header.Get("Referer")}
}
