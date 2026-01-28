package news

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// SetLabelsTask replaces private labels on a news item.
type SetLabelsTask struct{ tasks.TaskString }

// MarkReadTask clears the new and unread labels on a news post.
type MarkReadTask struct{ tasks.TaskString }

func (t *MarkReadTask) Matcher() mux.MatcherFunc {
	return tasks.HasFormOrQueryTask(string(t.TaskString))
}

var (
	setLabelsTask = &SetLabelsTask{TaskString: forumcommon.TaskSetLabels}
	markReadTask  = &MarkReadTask{TaskString: forumcommon.TaskMarkThreadRead}
)

var (
	_ tasks.Task = (*SetLabelsTask)(nil)
	_ tasks.Task = (*MarkReadTask)(nil)
)

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

func (SetLabelsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	postID, _ := strconv.Atoi(vars["news"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	priv := r.PostForm["private"]
	inverse := map[string]bool{"new": false, "unread": false}
	filtered := make([]string, 0, len(priv))
	for _, l := range priv {
		if _, ok := inverse[l]; ok {
			inverse[l] = true
			continue
		}
		filtered = append(filtered, l)
	}
	if err := cd.SetNewsPrivateLabels(int32(postID), filtered); err != nil {
		return fmt.Errorf("set private labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetPrivateLabelStatus("news", int32(postID), inverse["new"], inverse["unread"]); err != nil {
		return fmt.Errorf("set private label status %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return labelsRedirect(r)
}

// Action handles clearing new and unread labels on a news post.
func (MarkReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	postID, _ := strconv.Atoi(vars["news"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetPrivateLabelStatus("news", int32(postID), false, false); err != nil {
		return fmt.Errorf("mark read %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	target := r.PostFormValue("redirect")
	if target == "" {
		target = r.Header.Get("Referer")
	}
	if target == "" {
		target = strings.TrimSuffix(r.URL.Path, "/labels")
	}
	return handlers.RefreshDirectHandler{TargetURL: target}
}
