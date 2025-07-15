package imagebbs

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// AddImageBBSIndex injects image board index links into CoreData.
func AddImageBBSIndex(h http.Handler) http.Handler {
	return hcommon.IndexMiddleware(CustomImageBBSIndex)(h)
}

// RegisterRoutes attaches the public image board endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("ImageBBS", "/imagebbs", SectionWeight)
	nav.RegisterAdminControlCenter("ImageBBS", "/admin/imagebbs", SectionWeight)
	r.HandleFunc("/imagebbs.rss", RssPage).Methods("GET")
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.Use(AddImageBBSIndex)
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(config.AppRuntimeConfig.ImageUploadDir))))
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", BoardRssPage).Methods("GET")
	r.HandleFunc("/imagebbs.atom", AtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", BoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", BoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", UploadImageTask.Action).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(UploadImageTask.Match)
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", ReplyTask.Action).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(ReplyTask.Match)
	ibr.HandleFunc("", Page).Methods("GET")
	ibr.HandleFunc("/", Page).Methods("GET")
	ibr.HandleFunc("/poster/{username}", PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", PosterPage).Methods("GET")
}

// Register registers the imagebbs router module.
func Register() {
	router.RegisterModule("imagebbs", nil, RegisterRoutes)
}
