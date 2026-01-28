package blogs

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

// SetLabelsTask replaces private labels on a blog post.
type SetLabelsTask struct{ tasks.TaskString }

var setLabelsTask = &SetLabelsTask{TaskString: forumcommon.TaskSetLabels}

var _ tasks.Task = (*SetLabelsTask)(nil)

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
	blogID, _ := strconv.Atoi(vars["blog"])
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
	if err := cd.SetBlogPrivateLabels(int32(blogID), filtered); err != nil {
		return fmt.Errorf("set private labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetPrivateLabelStatus("blog", int32(blogID), inverse["new"], inverse["unread"]); err != nil {
		return fmt.Errorf("set private label status %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return labelsRedirect(r)
}
