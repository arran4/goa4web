package information

import (
	"github.com/gorilla/mux"

	router "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/sections"
)

// RegisterRoutes attaches the information endpoints to r.
func RegisterRoutes(r *mux.Router) {
	sections.RegisterIndexLink("Information", "/information", SectionWeight)
	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", Page).Methods("GET")
}

// Register registers the information router module.
func Register() {
	router.RegisterModule("information", nil, RegisterRoutes)
}
