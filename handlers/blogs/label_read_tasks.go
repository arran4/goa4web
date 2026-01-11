package blogs

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

// MarkBlogReadTask clears the special new/unread flags for a blog.
type MarkBlogReadTask struct{ tasks.TaskString }

var markBlogReadTask = &MarkBlogReadTask{TaskString: TaskMarkBlogRead}

func (t *MarkBlogReadTask) Matcher() mux.MatcherFunc {
	return tasks.HasFormOrQueryTask(string(t.TaskString))
}

var _ tasks.Task = (*MarkBlogReadTask)(nil)

func (MarkBlogReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	blogID, _ := strconv.Atoi(vars["blog"])
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.SetPrivateLabelStatus("blog", int32(blogID), false, false); err != nil {
		log.Printf("mark read: %v", err)
		return fmt.Errorf("mark read %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	target := r.FormValue("redirect")
	if target == "" {
		target = r.Header.Get("Referer")
	}
	if target == "" || !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") {
		target = strings.TrimSuffix(r.URL.Path, "/labels")
	}
	return handlers.RefreshDirectHandler{TargetURL: target}
}
