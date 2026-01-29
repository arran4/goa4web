package faq

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

func noTask() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		return r.PostFormValue("task") == ""
	}
}

// RegisterRoutes attaches the public FAQ endpoints to the router.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry, _ *sql.DB, _ sessions.Store) {
	navReg.RegisterIndexLinkWithViewPermission("Help", "/faq", SectionWeight, "faq", "question/answer")
	navReg.RegisterAdminControlCenter("Help", "Help Questions", "/admin/faq/questions", SectionWeight)
	navReg.RegisterAdminControlCenter("Help", "Help Categories", "/admin/faq/categories", SectionWeight+2)
	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.Use(handlers.IndexMiddleware(CustomFAQIndex))
	faqr.HandleFunc("/preview", handlers.PreviewPage).Methods("POST")
	faqr.HandleFunc("", Page).Methods("GET", "POST")
	faqr.HandleFunc("/ask", askTask.Page).Methods("GET")
	faqr.HandleFunc("/ask", handlers.TaskHandler(askTask)).Methods("POST").MatcherFunc(askTask.Matcher())
}

// RegisterAdminRoutes attaches the admin FAQ endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	farq := ar.PathPrefix("/faq").Subrouter()
	farq.Use(handlers.IndexMiddleware(CustomFAQIndex))
	farq.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	farq.HandleFunc("/categories/new", AdminNewCategoryPage).Methods("GET")
	farq.HandleFunc("/categories/category/{id:[0-9]+}", AdminCategoryPage).Methods("GET")
	farq.HandleFunc("/categories/category/{id:[0-9]+}/edit", AdminCategoryEditPage).Methods("GET")
	farq.HandleFunc("/categories/category/{id:[0-9]+}/questions", AdminCategoryQuestionsPage).Methods("GET")
	farq.HandleFunc("/categories", handlers.TaskHandler(renameCategoryTask)).Methods("POST").MatcherFunc(renameCategoryTask.Matcher())
	farq.HandleFunc("/categories", handlers.TaskHandler(deleteCategoryTask)).Methods("POST").MatcherFunc(deleteCategoryTask.Matcher())
	farq.HandleFunc("/categories", handlers.TaskHandler(createCategoryTask)).Methods("POST").MatcherFunc(createCategoryTask.Matcher())
	farq.HandleFunc("/questions", AdminQuestionsPage).Methods("GET").MatcherFunc(noTask())
	farq.HandleFunc("/questions", handlers.TaskHandler(editQuestionTask)).Methods("POST").MatcherFunc(editQuestionTask.Matcher())
	farq.HandleFunc("/questions", handlers.TaskHandler(deleteQuestionTask)).Methods("POST").MatcherFunc(deleteQuestionTask.Matcher())
	farq.HandleFunc("/questions", handlers.TaskHandler(createQuestionTask)).Methods("POST").MatcherFunc(createQuestionTask.Matcher())
	farq.HandleFunc("/question/create", AdminCreateQuestionPage).Methods("GET")
	farq.HandleFunc("/question/{id:[0-9]+}/edit", AdminEditQuestionPage).Methods("GET")
	farq.HandleFunc("/revisions/{id:[0-9]+}", AdminRevisionHistoryPage).Methods("GET")
}

// Register registers the faq router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("faq", nil, RegisterRoutes)
}
