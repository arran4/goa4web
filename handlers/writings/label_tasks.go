package writings

import (
	"fmt"
	"log"
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

// SetLabelsTask replaces private labels on a writing.
type SetLabelsTask struct{ tasks.TaskString }

var setLabelsTask = &SetLabelsTask{TaskString: forumcommon.TaskSetLabels}

var _ tasks.Task = (*SetLabelsTask)(nil)

// MarkWritingReadTask clears the special new/unread flags for a writing and its thread.
type MarkWritingReadTask struct{ tasks.TaskString }

var markWritingReadTask = &MarkWritingReadTask{TaskString: forumcommon.TaskMarkThreadRead}

var _ tasks.Task = (*MarkWritingReadTask)(nil)

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

func (MarkWritingReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	writingID, _ := strconv.Atoi(vars["writing"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	threadID, err := cd.Queries().SystemGetWritingByID(r.Context(), int32(writingID))
	if err != nil {
		log.Printf("get writing thread: %v", err)
	} else {
		if err := cd.SetThreadPrivateLabelStatus(threadID, false, false); err != nil {
			log.Printf("mark thread read: %v", err)
			return fmt.Errorf("mark thread read %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if last := r.PostFormValue("last_comment"); last != "" {
			if cid, err := strconv.Atoi(last); err == nil {
				if err := cd.SetThreadReadMarker(threadID, int32(cid)); err != nil {
					log.Printf("set read marker: %v", err)
					return fmt.Errorf("set read marker %w", handlers.ErrRedirectOnSamePageHandler(err))
				}
			}
		}
	}

	if err := cd.SetPrivateLabelStatus("writing", int32(writingID), false, false); err != nil {
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
