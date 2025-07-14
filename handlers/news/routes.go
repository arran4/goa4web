package news

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	comments "github.com/arran4/goa4web/handlers/comments"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	nav "github.com/arran4/goa4web/internal/navigation"
)

func runTemplate(tmpl string) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type Data struct {
			*corecommon.CoreData
		}

		data := Data{
			CoreData: r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData),
		}

		CustomNewsIndex(data.CoreData, r)

		if err := templates.RenderTemplate(w, tmpl, data, corecommon.NewFuncs(r)); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}

func AddNewsIndex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData)
		CustomNewsIndex(cd, r)
		handler.ServeHTTP(w, r)
	})
}

// RegisterRoutes attaches the public news endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("News", "/", SectionWeight)
	nav.RegisterAdminControlCenter("News", "/admin/news/users/levels", SectionWeight)
	r.Handle("/", AddNewsIndex(http.HandlerFunc(runTemplate("newsPage")))).Methods("GET")
	r.HandleFunc("/", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	r.HandleFunc("/news.rss", NewsRssPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(AddNewsIndex)
	nr.HandleFunc("", runTemplate("newsPage")).Methods("GET")
	nr.HandleFunc("", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", ReplyTask.Action).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(ReplyTask.Match)
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(NewsPostCommentEditActionPage))).Methods("POST").MatcherFunc(hcommon.EditReplyTask.Match)
	nr.Handle("/news/{post}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(NewsPostCommentEditActionCancelPage))).Methods("POST").MatcherFunc(hcommon.CancelTask.Match)
	nr.HandleFunc("/news/{post}", EditTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("content writer", "administrator")).MatcherFunc(EditTask.Match)
	nr.HandleFunc("/news/{post}", NewPostTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("content writer", "administrator")).MatcherFunc(NewPostTask.Match)
	nr.HandleFunc("/news/{post}/announcement", AnnouncementAddTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(AnnouncementAddTask.Match)
	nr.HandleFunc("/news/{post}/announcement", AnnouncementDeleteTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(AnnouncementDeleteTask.Match)
	nr.HandleFunc("/news/{post}", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(hcommon.CancelTask.Match)
	nr.HandleFunc("/news/{post}", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", NewsUserPermissionsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", UserAllowTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UserAllowTask.Match)
	nr.HandleFunc("/users/permissions", UserDisallowTask.Action).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(UserDisallowTask.Match)
}

// Register registers the news router module.
func Register() {
	router.RegisterModule("news", nil, RegisterRoutes)
}
