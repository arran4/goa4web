package imagebbs

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/tasks"

	nav "github.com/arran4/goa4web/internal/navigation"
)

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
	ibr.HandleFunc("/board/{boardno}", tasks.Action(uploadImageTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(uploadImageTask.Matcher())
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", tasks.Action(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	ibr.HandleFunc("", Page).Methods("GET")
	ibr.HandleFunc("/", Page).Methods("GET")
	ibr.HandleFunc("/poster/{username}", PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", PosterPage).Methods("GET")
}

// Register registers the imagebbs router module.
func Register() {
	router.RegisterModule("imagebbs", nil, RegisterRoutes)
}
