package writings

import (
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", tasks.Action(userAllowTask)).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/permissions", tasks.Action(userDisallowTask)).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", tasks.Action(userAllowTask)).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/levels", tasks.Action(userDisallowTask)).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", tasks.Action(writingCategoryChangeTask)).Methods("POST").MatcherFunc(writingCategoryChangeTask.Matcher())
	war.HandleFunc("/categories", tasks.Action(writingCategoryCreateTask)).Methods("POST").MatcherFunc(writingCategoryCreateTask.Matcher())
}
