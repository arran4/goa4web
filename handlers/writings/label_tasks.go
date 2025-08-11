package writings

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// SetLabelsTask replaces private labels on a writing.
type SetLabelsTask struct{ tasks.TaskString }

var setLabelsTask = &SetLabelsTask{TaskString: forumhandlers.TaskSetLabels}

var _ tasks.Task = (*SetLabelsTask)(nil)

// labelsRedirect determines the page to return to after processing a label task.
// It first checks the "back" form value, then falls back to the Referer header
// before defaulting to the current page when neither is present.
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
	writingID, _ := strconv.Atoi(vars["writing"])
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
	if err := cd.SetWritingPrivateLabels(int32(writingID), filtered); err != nil {
		return fmt.Errorf("set private labels %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetPrivateLabelStatus("writing", int32(writingID), inverse["new"], inverse["unread"]); err != nil {
		return fmt.Errorf("set private label status %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return labelsRedirect(r)
}
