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
	rr.HandleFunc("", RegisterTask.Page).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", RegisterTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(RegisterTask.Match)

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", LoginTask.Page).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	lr.HandleFunc("", LoginTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(LoginTask.Match)
	lr.HandleFunc("/verify", LoginVerifyPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(hcommon.SaveAllTask.Match)

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", ForgotPasswordPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	fr.HandleFunc("", ForgotPasswordActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(hcommon.AddTask.Match)
}

// Register registers the auth router module.
func Register() {
	router.RegisterModule("auth", nil, RegisterRoutes)
}
