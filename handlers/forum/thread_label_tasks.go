package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
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

// AddPublicLabelTask adds a public label to a thread.
type AddPublicLabelTask struct{ tasks.TaskString }

// RemovePublicLabelTask removes a public label from a thread.
type RemovePublicLabelTask struct{ tasks.TaskString }

// AddPrivateLabelTask adds a private label for the current user.
type AddPrivateLabelTask struct{ tasks.TaskString }

// RemovePrivateLabelTask removes a private label for the current user.
type RemovePrivateLabelTask struct{ tasks.TaskString }

var (
	addPublicLabelTask     = &AddPublicLabelTask{TaskString: forumcommon.TaskAddPublicLabel}
	removePublicLabelTask  = &RemovePublicLabelTask{TaskString: forumcommon.TaskRemovePublicLabel}
	addPrivateLabelTask    = &AddPrivateLabelTask{TaskString: forumcommon.TaskAddPrivateLabel}
	removePrivateLabelTask = &RemovePrivateLabelTask{TaskString: forumcommon.TaskRemovePrivateLabel}
	addAuthorLabelTask     = &AddAuthorLabelTask{TaskString: forumcommon.TaskAddAuthorLabel}
	removeAuthorLabelTask  = &RemoveAuthorLabelTask{TaskString: forumcommon.TaskRemoveAuthorLabel}
	markThreadReadTask     = &MarkThreadReadTask{TaskString: forumcommon.TaskMarkThreadRead}
	setLabelsTask          = &SetLabelsTask{TaskString: forumcommon.TaskSetLabels}
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

// labelsRedirect determines the page to return to after processing a label task.
// It first checks the "back" form value, then falls back to the Referer header
// before defaulting to the thread page when neither is present.
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
	_ tasks.Task = (*MarkThreadReadTask)(nil)
	_ tasks.Task = (*SetLabelsTask)(nil)
)

func (AddPublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	return labelsRedirect(r)
}

func (RemovePublicLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	return labelsRedirect(r)
}

func (AddPrivateLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	return labelsRedirect(r)
}

func (RemovePrivateLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	return labelsRedirect(r)
}

// AddAuthorLabelTask adds an author-only label to a thread.
type AddAuthorLabelTask struct{ tasks.TaskString }

// RemoveAuthorLabelTask removes an author-only label from a thread.
type RemoveAuthorLabelTask struct{ tasks.TaskString }

func (AddAuthorLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	return labelsRedirect(r)
}

func (RemoveAuthorLabelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	return labelsRedirect(r)
}

// MarkThreadReadTask clears the special new/unread flags for a thread.
type MarkThreadReadTask struct{ tasks.TaskString }

func (t *MarkThreadReadTask) Matcher() mux.MatcherFunc {
	return tasks.HasFormOrQueryTask(string(t.TaskString))
}

func (MarkThreadReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetThreadPrivateLabelStatus(int32(threadID), false, false); err != nil {
		log.Printf("mark read: %v", err)
		return fmt.Errorf("mark read %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if last := r.FormValue("last_comment"); last != "" {
		if cid, err := strconv.Atoi(last); err == nil {
			if err := cd.SetThreadReadMarker(int32(threadID), int32(cid)); err != nil {
				log.Printf("set read marker: %v", err)
				return fmt.Errorf("set read marker %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}

	target := r.FormValue("redirect")
	if target == "" {
		target = r.Header.Get("Referer")
	}
	if target == "" {
		target = strings.TrimSuffix(r.URL.Path, "/labels")
	}
	return handlers.RefreshDirectHandler{TargetURL: target}
}

// SetLabelsTask replaces public and private labels on a thread.
type SetLabelsTask struct{ tasks.TaskString }

func (SetLabelsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
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
	if err := cd.SetThreadPublicLabels(int32(threadID), pub); err != nil {
		log.Printf("set public labels: %v", err)
		return fmt.Errorf("set public labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetThreadPrivateLabels(int32(threadID), filteredPriv); err != nil {
		log.Printf("set private labels: %v", err)
		return fmt.Errorf("set private labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := cd.SetThreadPrivateLabelStatus(int32(threadID), inverse["new"], inverse["unread"]); err != nil {
		log.Printf("set private label status: %v", err)
		return fmt.Errorf("set private label status %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return labelsRedirect(r)
}
