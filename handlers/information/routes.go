package information

import "github.com/gorilla/mux"

// RegisterRoutes attaches the information endpoints to r.
func RegisterRoutes(r *mux.Router) {
	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", Page).Methods("GET")
}
