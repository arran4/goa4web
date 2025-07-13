package writings

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
	"net/http"

	auth "github.com/arran4/goa4web/handlers/auth"
	comments "github.com/arran4/goa4web/handlers/comments"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
	pkghandlers "github.com/arran4/goa4web/pkg/handlers"
)

var legacyRedirectsEnabled = true

// RegisterRoutes attaches the public writings endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Writings", "/writings", SectionWeight)
	nav.RegisterAdminControlCenter("Writings", "/admin/writings/categories", SectionWeight)
	wr := r.PathPrefix("/writings").Subrouter()
	wr.HandleFunc("/rss", RssPage).Methods("GET")
	wr.HandleFunc("/atom", AtomPage).Methods("GET")
	wr.HandleFunc("", Page).Methods("GET")
	wr.HandleFunc("/", Page).Methods("GET")
	wr.HandleFunc("/writer/{username}", WriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", WriterPage).Methods("GET")
	wr.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserAllow))
	wr.HandleFunc("/users/permissions", UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserDisallow))
	wr.HandleFunc("/article/{article}", ArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", ArticleReplyActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	wr.HandleFunc("/article/{article}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(ArticleCommentEditActionPage)).ServeHTTP).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEditReply))
	wr.HandleFunc("/article/{article}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(ArticleCommentEditActionCancelPage)).ServeHTTP).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	wr.Handle("/article/{article}/edit", RequireWritingAuthor(http.HandlerFunc(ArticleEditPage))).Methods("GET").MatcherFunc(auth.RequiredAccess("content writer", "administrator"))
	wr.Handle("/article/{article}/edit", RequireWritingAuthor(http.HandlerFunc(ArticleEditActionPage))).Methods("POST").MatcherFunc(auth.RequiredAccess("content writer", "administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUpdateWriting))
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", ArticleAddPage).Methods("GET").MatcherFunc(Or(auth.RequiredAccess("content writer"), auth.RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", ArticleAddActionPage).Methods("POST").MatcherFunc(Or(auth.RequiredAccess("content writer"), auth.RequiredAccess("administrator"))).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSubmitWriting))

	if legacyRedirectsEnabled {
		// legacy redirects
		r.Path("/writing").HandlerFunc(pkghandlers.RedirectPermanentPrefix("/writing", "/writings"))
		r.PathPrefix("/writing/").HandlerFunc(pkghandlers.RedirectPermanentPrefix("/writing", "/writings"))
	}
}

// Register registers the writings router module.
func Register() {
	router.RegisterModule("writings", nil, RegisterRoutes)
}
