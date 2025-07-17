package imagebbs

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// AddImageBBSIndex injects image board index links into CoreData.
func AddImageBBSIndex(h http.Handler) http.Handler {
	return handlers.IndexMiddleware(CustomImageBBSIndex)(h)
}

// RegisterRoutes attaches the public image board endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("ImageBBS", "/imagebbs", SectionWeight)
	nav.RegisterAdminControlCenter("ImageBBS", "/admin/imagebbs", SectionWeight)
	r.HandleFunc("/imagebbs.rss", RssPage).Methods("GET")
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.Use(handlers.IndexMiddleware(CustomImageBBSIndex))
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(config.AppRuntimeConfig.ImageUploadDir))))
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", BoardRssPage).Methods("GET")
	r.HandleFunc("/imagebbs.atom", AtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", BoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", BoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", uploadImageTask.Action).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(uploadImageTask.Match)
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", replyTask.Action).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Match)
	ibr.HandleFunc("", Page).Methods("GET")
	ibr.HandleFunc("/", Page).Methods("GET")
	ibr.HandleFunc("/poster/{username}", PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", PosterPage).Methods("GET")
}

// Register registers the imagebbs router module.
func Register() {
	router.RegisterModule("imagebbs", nil, RegisterRoutes)
}
