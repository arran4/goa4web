package imagebbs

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the public image board endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLinkWithViewPermission("ImageBBS", "/imagebbs", SectionWeight, "imagebbs", "board")
	navReg.RegisterAdminControlCenter("ImageBBS", "ImageBBS", "/admin/imagebbs", SectionWeight)
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	ibr.Use(handlers.IndexMiddleware(CustomImageBBSIndex), handlers.SectionMiddleware("imagebbs"))
	ibr.HandleFunc("/rss", RssPage).Methods("GET")
	ibr.HandleFunc("/u/{username}/rss", RssPage).Methods("GET")
	ibr.HandleFunc("/atom", AtomPage).Methods("GET")
	ibr.HandleFunc("/u/{username}/atom", AtomPage).Methods("GET")
	bbsDir := cfg.ImageUploadDir
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(bbsDir))))
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", BoardRssPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", BoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", ImagebbsBoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", handlers.TaskHandler(uploadImageTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(uploadImageTask.Matcher())
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	ibr.HandleFunc("", ImagebbsPage).Methods("GET")
	ibr.HandleFunc("/", ImagebbsPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}", PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", PosterPage).Methods("GET")

}

// Register registers the imagebbs router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("imagebbs", nil, RegisterRoutes)
}
