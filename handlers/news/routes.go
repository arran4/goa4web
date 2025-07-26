package news

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	htemplate "html/template"
	"net/http"
	"sync"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers/forum/comments"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

var (
	siteTemplates     *htemplate.Template
	loadTemplatesOnce sync.Once
)

func runTemplate(name string) http.HandlerFunc {
	loadTemplatesOnce.Do(func() {
		siteTemplates = templates.GetCompiledSiteTemplates((&common.CoreData{}).Funcs(nil))
	})
	if siteTemplates.Lookup(name) == nil {
		panic(fmt.Sprintf("missing template %s", name))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		handlers.TemplateHandler(w, r, name, r.Context().Value(consts.KeyCoreData))
	}
}

// RegisterRoutes attaches the public news endpoints to r.
func RegisterRoutes(r *mux.Router, navReg *nav.Registry) {
	navReg.RegisterIndexLink("News", "/", SectionWeight)
	navReg.RegisterAdminControlCenter("News", "/admin/news/users/roles", SectionWeight)
	r.Use(handlers.IndexMiddleware(CustomNewsIndex))
	r.HandleFunc("/", runTemplate("newsPage")).Methods("GET")
	r.HandleFunc("/", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	r.HandleFunc("/news.rss", NewsRssPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(handlers.IndexMiddleware(CustomNewsIndex))
	nr.HandleFunc("", runTemplate("newsPage")).Methods("GET")
	nr.HandleFunc("", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{post}", handlers.TaskHandler(editTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(editTask.Matcher())
	nr.HandleFunc("/news/{post}", handlers.TaskHandler(newPostTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(newPostTask.Matcher())
	nr.HandleFunc("/news/{post}/announcement", handlers.TaskHandler(announcementAddTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementAddTask.Matcher())
	nr.HandleFunc("/news/{post}/announcement", handlers.TaskHandler(announcementDeleteTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(announcementDeleteTask.Matcher())
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(cancelTask.Matcher())
	nr.HandleFunc("/news/{post}", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", NewsUserPermissionsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", handlers.TaskHandler(userAllowTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Matcher())
	nr.HandleFunc("/users/permissions", handlers.TaskHandler(userDisallowTask)).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Matcher())
}

// Register registers the news router module.
func Register(reg *router.Registry, navReg *nav.Registry) {
	reg.RegisterModule("news", nil, func(r *mux.Router) { RegisterRoutes(r, navReg) })
}
