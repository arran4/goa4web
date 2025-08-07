package linker

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("", AdminDashboardPage).Methods("GET")
	lar.HandleFunc("/", AdminDashboardPage).Methods("GET")

	// categories
	car := lar.PathPrefix("/categories").Subrouter()
	car.HandleFunc("", AdminCategoriesPage).Methods("GET")
	car.HandleFunc("/", AdminCategoriesPage).Methods("GET")
	car.HandleFunc("", handlers.TaskHandler(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	car.HandleFunc("", handlers.TaskHandler(RenameCategoryTask)).Methods("POST").MatcherFunc(RenameCategoryTask.Matcher())
	car.HandleFunc("", handlers.TaskHandler(AdminDeleteCategoryTask)).Methods("POST").MatcherFunc(AdminDeleteCategoryTask.Matcher())
	car.HandleFunc("", handlers.TaskHandler(CreateCategoryTask)).Methods("POST").MatcherFunc(CreateCategoryTask.Matcher())

	cat := car.PathPrefix("/category/{category}").Subrouter()
	cat.HandleFunc("", AdminCategoryPage).Methods("GET")
	cat.HandleFunc("/edit", AdminCategoryEditPage).Methods("GET")
	cat.HandleFunc("/grants", AdminCategoryGrantsPage).Methods("GET")
	cat.HandleFunc("/links", AdminCategoryPage).Methods("GET")
	cat.HandleFunc("/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	cat.HandleFunc("/grant/delete", handlers.TaskHandler(AdminCategoryGrantDeleteTask)).Methods("POST").MatcherFunc(AdminCategoryGrantDeleteTask.Matcher())

	// link list and items
	links := lar.PathPrefix("/links").Subrouter()
	links.HandleFunc("", AdminLinksPage).Methods("GET")
	links.HandleFunc("/", AdminLinksPage).Methods("GET")
	link := links.PathPrefix("/link/{link}").Subrouter()
	link.HandleFunc("", adminLinkViewPage).Methods("GET")
	link.HandleFunc("/edit", adminLinkPage).Methods("GET")
	link.HandleFunc("/edit", handlers.TaskHandler(AdminEditLinkTask)).Methods("POST").MatcherFunc(AdminEditLinkTask.Matcher())
	link.HandleFunc("/grants", AdminLinkGrantsPage).Methods("GET")
	link.HandleFunc("/grant", handlers.TaskHandler(linkGrantCreateTask)).Methods("POST").MatcherFunc(linkGrantCreateTask.Matcher())
	link.HandleFunc("/grant/delete", handlers.TaskHandler(AdminLinkGrantDeleteTask)).Methods("POST").MatcherFunc(AdminLinkGrantDeleteTask.Matcher())
	link.HandleFunc("/comments", CommentsPage).Methods("GET")

	// misc
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", handlers.TaskHandler(AdminAddTask)).Methods("POST").MatcherFunc(AdminAddTask.Matcher())
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminDeleteTask)).Methods("POST").MatcherFunc(AdminDeleteTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminApproveTask)).Methods("POST").MatcherFunc(AdminApproveTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminBulkApproveTask)).Methods("POST").MatcherFunc(AdminBulkApproveTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminBulkDeleteTask)).Methods("POST").MatcherFunc(AdminBulkDeleteTask.Matcher())
}
