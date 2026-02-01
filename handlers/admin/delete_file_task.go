package admin

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type DeleteUnmanagedFileTask struct{}

var _ tasks.Task = (*DeleteUnmanagedFileTask)(nil)

// Action deletes the specified file from the image upload directory.
func (t *DeleteUnmanagedFileTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	reqPath := r.FormValue("path")
	if reqPath == "" {
		return fmt.Errorf("path is required")
	}

	base := cd.Config.ImageUploadDir
	cleaned := filepath.Clean("/" + reqPath)
	abs := filepath.Join(base, cleaned)
	if rel, err := filepath.Rel(base, abs); err != nil || rel == ".." || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("invalid path")
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("cannot delete directory")
	}

	if err := os.Remove(abs); err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	// Redirect back to listing, parent dir of deleted file
	parent := filepath.Dir(cleaned)
	return handlers.RedirectHandler("/admin/files/unmanaged?path=" + parent)
}

// Matcher returns a matcher that checks if the request is a POST to /admin/files/delete.
func (t *DeleteUnmanagedFileTask) Matcher() mux.MatcherFunc {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.URL.Path == "/admin/files/delete" && r.Method == "POST"
	}
}

func (h *Handlers) NewDeleteUnmanagedFileTask() *DeleteUnmanagedFileTask {
	return &DeleteUnmanagedFileTask{}
}
