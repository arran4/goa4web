package externallink

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ReloadExternalLinkTask handles HTTP request for reloading a link.
type ReloadExternalLinkTask struct {
	tasks.TaskString
}

var reloadExternalLinkTask = &ReloadExternalLinkTask{TaskString: "admin:externallink:reload"}

var _ tasks.Task = (*ReloadExternalLinkTask)(nil)

func (ReloadExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL, _, _, _, err := cd.ResolveExternalLink(r)
	if err != nil {
		return fmt.Errorf("invalid link: %w", handlers.ErrForbidden)
	}

	if evt := cd.Event(); evt != nil {
		evt.Task = &tasks.ReloadExternalLinkTask{
			TaskString: "admin:externallink:reload",
			URL:        rawURL,
			Config:     cd.Config,
		}
	}

	redirectURI := r.RequestURI
	if r.URL.Query().Get("msg") == "" {
		redirectURI += "&msg=Reloading+Open+Graph+data+in+the+background..."
	}
	return handlers.RedirectHandler(redirectURI)
}

func (t *ReloadExternalLinkTask) Matcher() func(*http.Request, *mux.RouteMatch) bool {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}
}

// PrefetchExternalLinkTask handles HTTP request for prefetching a link.
type PrefetchExternalLinkTask struct {
	tasks.TaskString
}

var prefetchExternalLinkTask = &PrefetchExternalLinkTask{TaskString: "externallink:prefetch"}

var _ tasks.Task = (*PrefetchExternalLinkTask)(nil)

func (PrefetchExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL := r.FormValue("url")
	if rawURL == "" {
		return handlers.TextByteWriter("missing url")
	}

	if evt := cd.Event(); evt != nil {
		evt.Task = &tasks.PrefetchExternalLinkTask{
			TaskString: "externallink:prefetch",
			URL:        rawURL,
			Config:     cd.Config,
		}
	}

	return handlers.TextByteWriter("ok")
}

func GetReloadExternalLinkTask() *ReloadExternalLinkTask {
	return reloadExternalLinkTask
}

func GetPrefetchExternalLinkTask() *PrefetchExternalLinkTask {
	return prefetchExternalLinkTask
}
