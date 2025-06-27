package information

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/internal/sections"
)

// RegisterRoutes attaches the information endpoints to r.
func RegisterRoutes(r *mux.Router) {
	sections.RegisterIndexLink("Information", "/information", SectionWeight)
	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", Page).Methods("GET")
}
