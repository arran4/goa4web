package news

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
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

		log.Printf("rendering template %s", tmpl)

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
	r.Handle("/", AddNewsIndex(http.HandlerFunc(runTemplate("newsPage")))).Methods("GET")
	r.HandleFunc("/", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	r.HandleFunc("/news.rss", NewsRssPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(AddNewsIndex)
	nr.HandleFunc("", runTemplate("newsPage")).Methods("GET")
	nr.HandleFunc("", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", NewsPostReplyActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	nr.HandleFunc("/news/{post}", NewsPostEditActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskEdit))
	nr.HandleFunc("/news/{post}", NewsPostNewActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskNewPost))
	nr.HandleFunc("/news/{post}/announcement", NewsAnnouncementActivateActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskAdd))
	nr.HandleFunc("/news/{post}/announcement", NewsAnnouncementDeactivateActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDelete))
	nr.HandleFunc("/news/{post}", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCancel))
	nr.HandleFunc("/news/{post}", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", NewsUserPermissionsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", NewsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher("User Allow"))
	nr.HandleFunc("/users/permissions", NewsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(hcommon.TaskMatcher("User Disallow"))
}
