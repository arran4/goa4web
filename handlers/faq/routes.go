package faq

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

func noTask() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		return r.PostFormValue("task") == ""
	}
}

// RegisterRoutes attaches the public FAQ endpoints to the router.
func RegisterRoutes(r *mux.Router, navReg *nav.Registry) {
	navReg.RegisterIndexLink("FAQ", "/faq", SectionWeight)
	navReg.RegisterAdminControlCenter("FAQ", "/admin/faq/categories", SectionWeight)
	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.Use(handlers.IndexMiddleware(CustomFAQIndex))
	faqr.HandleFunc("", Page).Methods("GET", "POST")
	faqr.HandleFunc("/ask", askTask.Page).Methods("GET")
	faqr.HandleFunc("/ask", handlers.TaskHandler(askTask)).Methods("POST").MatcherFunc(askTask.Matcher())
}

// RegisterAdminRoutes attaches the admin FAQ endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	farq := ar.PathPrefix("/faq").Subrouter()
	farq.Use(handlers.IndexMiddleware(CustomFAQIndex))
	farq.HandleFunc("/answer", AdminAnswerPage).Methods("GET", "POST").MatcherFunc(noTask())
	farq.HandleFunc("/answer", handlers.TaskHandler(answerTask)).Methods("POST").MatcherFunc(answerTask.Matcher())
	farq.HandleFunc("/answer", handlers.TaskHandler(removeQuestionTask)).Methods("POST").MatcherFunc(removeQuestionTask.Matcher())
	farq.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	farq.HandleFunc("/categories", handlers.TaskHandler(renameCategoryTask)).Methods("POST").MatcherFunc(renameCategoryTask.Matcher())
	farq.HandleFunc("/categories", handlers.TaskHandler(deleteCategoryTask)).Methods("POST").MatcherFunc(deleteCategoryTask.Matcher())
	farq.HandleFunc("/categories", handlers.TaskHandler(createCategoryTask)).Methods("POST").MatcherFunc(createCategoryTask.Matcher())
	farq.HandleFunc("/questions", AdminQuestionsPage).Methods("GET", "POST").MatcherFunc(noTask())
	farq.HandleFunc("/questions", handlers.TaskHandler(editQuestionTask)).Methods("POST").MatcherFunc(editQuestionTask.Matcher())
	farq.HandleFunc("/questions", handlers.TaskHandler(deleteQuestionTask)).Methods("POST").MatcherFunc(deleteQuestionTask.Matcher())
	farq.HandleFunc("/questions", handlers.TaskHandler(createQuestionTask)).Methods("POST").MatcherFunc(createQuestionTask.Matcher())
}

// Register registers the faq router module.
func Register(reg *router.Registry, navReg *nav.Registry) {
	reg.RegisterModule("faq", nil, func(r *mux.Router) { RegisterRoutes(r, navReg) })
}
