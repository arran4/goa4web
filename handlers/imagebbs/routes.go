package imagebbs

import (
	"net/http"

	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/runtimeconfig"

	"github.com/arran4/goa4web/internal/sections"
)

// RegisterRoutes attaches the public image board endpoints to r.
func RegisterRoutes(r *mux.Router) {
	sections.RegisterIndexLink("ImageBBS", "/imagebbs", SectionWeight)
	sections.RegisterAdminControlCenter("ImageBBS", "/admin/imagebbs", SectionWeight)
	r.HandleFunc("/imagebbs.rss", RssPage).Methods("GET")
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(runtimeconfig.AppRuntimeConfig.ImageUploadDir))))
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", BoardRssPage).Methods("GET")
	r.HandleFunc("/imagebbs.atom", AtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", BoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", BoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", BoardPostImageActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUploadImage))
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", BoardThreadReplyActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	ibr.HandleFunc("", Page).Methods("GET")
	ibr.HandleFunc("/", Page).Methods("GET")
	ibr.HandleFunc("/poster/{username}", PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", PosterPage).Methods("GET")
}

// Register registers the imagebbs router module.
func Register() {
	router.RegisterModule("imagebbs", nil, RegisterRoutes)
}
