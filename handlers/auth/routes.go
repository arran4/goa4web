package auth

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router) {
	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", RegisterPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", RegisterActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskRegister))

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", LoginUserPassPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	lr.HandleFunc("", LoginActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskLogin))
}

// Register registers the auth router module.
func Register() {
	router.RegisterModule("auth", nil, RegisterRoutes)
}
