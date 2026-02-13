package bookmarks

import (
	"net/http"

	"github.com/arran4/goa4web/handlers/bookmarks/routes"
	"github.com/gorilla/mux"

	"github.com/arran4/go-consume"
	"github.com/arran4/go-consume/strconsume"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

const (
	// SectionWeight controls the order of the Bookmarks section in navigation.
	SectionWeight = 70
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	r.PathPrefix("/bookmarks").Handler(NewRouter())
}

func NewRouter() *Router {
	return &Router{
		consumeUntilSlash: strconsume.NewUntilConsumer("/"),
	}
}

// Register registers the bookmarks router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("bookmarks", nil, RegisterRoutes)
}

type Router struct {
	consumeUntilSlash strconsume.UntilConsumer
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	var ok bool
	_, _, p, ok = r.consumeUntilSlash.Consume(p, consume.Inclusive(true), consume.ConsumeRemainingIfNotFound(true))
	if !ok {
		handlers.RenderNotFoundOrLogin(w, req)
		return
	}
	cd, ok := req.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok {
		handlers.RenderNotFoundOrLogin(w, req)
		return
	}
	cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
		Name: "Show",
		Link: "/bookmarks/mine",
	})
	cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
		Name: "Edit",
		Link: "/bookmarks/edit",
	})
	switch p {
	case "", "/":
		r.Serve(w, req, cd)
		return
	case "/mine":
		r.ServeMine(w, req, cd)
		return
	case "/edit":
		r.ServeEdit(w, req, cd)
		return
	}
	handlers.RenderNotFoundOrLogin(w, req)
}

func (r *Router) Serve(w http.ResponseWriter, req *http.Request, cd *common.CoreData) {
	routes.BookmarksPage(w, req)
}

func (r *Router) ServeMine(w http.ResponseWriter, req *http.Request, cd *common.CoreData) {
	if !cd.IsUserLoggedIn() {
		handlers.RenderNotFoundOrLogin(w, req)
		return
	}
	switch req.Method {
	case "GET":
		routes.MinePage(w, req)
		return
	}
	handlers.RenderNotFoundOrLogin(w, req)
}

func (r *Router) ServeEdit(w http.ResponseWriter, req *http.Request, cd *common.CoreData) {
	if !cd.IsUserLoggedIn() {
		handlers.RenderNotFoundOrLogin(w, req)
		return
	}
	switch req.Method {
	case "GET":
		routes.EditPage(w, req)
		return
	case "POST":
		task := req.PostFormValue("task")
		switch task {
		case string(routes.TaskSave):
			handlers.TaskHandler(routes.SaveTask)(w, req)
			return
		case string(routes.TaskCreate):
			handlers.TaskHandler(routes.CreateTask)(w, req)
			return
		}
	}
	handlers.RenderNotFoundOrLogin(w, req)
}
