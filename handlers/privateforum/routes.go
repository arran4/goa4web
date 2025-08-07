package privateforum

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the private forum endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Private", "/private", SectionWeight)
	pr := r.PathPrefix("/private").Subrouter()
	pr.Use(handlers.IndexMiddleware(CustomIndex), handlers.SectionMiddleware("privateforum"))
	pr.HandleFunc("", Page).Methods(http.MethodGet)
	pr.HandleFunc("", handlers.TaskHandler(privateTopicCreateTask)).Methods(http.MethodPost).MatcherFunc(privateTopicCreateTask.Matcher())
	pr.HandleFunc("/private_forum.js", handlers.PrivateForumJS).Methods(http.MethodGet)
}

// Register registers the private forum router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("privateforum", []string{"private"}, RegisterRoutes)
}
