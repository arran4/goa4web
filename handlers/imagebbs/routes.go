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
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig) []navpkg.RouterOptions {
	opts := []navpkg.RouterOptions{
		navpkg.NewIndexLinkWithViewPermission("ImageBBS", "/imagebbs", `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-image"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><circle cx="8.5" cy="8.5" r="1.5"></circle><polyline points="21 15 16 10 5 21"></polyline></svg>`, SectionWeight, "imagebbs", "board"),
		navpkg.NewAdminControlCenterLink(navpkg.AdminCCCategory("ImageBBS"), "ImageBBS", "/admin/imagebbs", SectionWeight),
	}
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
	ibr.HandleFunc("/board/{boardno}", CheckBoardViewGrant(ImagebbsBoardPage)).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", handlers.TaskHandler(uploadImageTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(uploadImageTask.Matcher())
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", CheckBoardViewGrant(BoardThreadPage)).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(replyTask.Matcher())
	ibr.HandleFunc("", ImagebbsPage).Methods("GET")
	ibr.HandleFunc("/", ImagebbsPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}", PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", PosterPage).Methods("GET")
	return opts
}

// Register registers the imagebbs router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("imagebbs", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []navpkg.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
