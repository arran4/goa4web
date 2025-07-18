package faq

import (
	"github.com/gorilla/mux"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

func noTask() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		return r.PostFormValue("task") == ""
	}
}

// RegisterRoutes attaches the public FAQ endpoints to the router.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("FAQ", "/faq", SectionWeight)
	nav.RegisterAdminControlCenter("FAQ", "/admin/faq/categories", SectionWeight)
	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.Use(handlers.IndexMiddleware(CustomFAQIndex))
	faqr.HandleFunc("", Page).Methods("GET", "POST")
	faqr.HandleFunc("/ask", askTask.Page).Methods("GET")
	faqr.HandleFunc("/ask", askTask.Action).Methods("POST").MatcherFunc(askTask.Matcher())
}

// RegisterAdminRoutes attaches the admin FAQ endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	farq := ar.PathPrefix("/faq").Subrouter()
	farq.Use(handlers.IndexMiddleware(CustomFAQIndex))
	farq.HandleFunc("/answer", AdminAnswerPage).Methods("GET", "POST").MatcherFunc(noTask())
	farq.HandleFunc("/answer", answerTask.Action).Methods("POST").MatcherFunc(answerTask.Matcher())
	farq.HandleFunc("/answer", removeQuestionTask.Action).Methods("POST").MatcherFunc(removeQuestionTask.Matcher())
	farq.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	farq.HandleFunc("/categories", renameCategoryTask.Action).Methods("POST").MatcherFunc(renameCategoryTask.Matcher())
	farq.HandleFunc("/categories", deleteCategoryTask.Action).Methods("POST").MatcherFunc(deleteCategoryTask.Matcher())
	farq.HandleFunc("/categories", createCategoryTask.Action).Methods("POST").MatcherFunc(createCategoryTask.Matcher())
	farq.HandleFunc("/questions", AdminQuestionsPage).Methods("GET", "POST").MatcherFunc(noTask())
	farq.HandleFunc("/questions", editQuestionTask.Action).Methods("POST").MatcherFunc(editQuestionTask.Matcher())
	farq.HandleFunc("/questions", deleteQuestionTask.Action).Methods("POST").MatcherFunc(deleteQuestionTask.Matcher())
	farq.HandleFunc("/questions", createQuestionTask.Action).Methods("POST").MatcherFunc(createQuestionTask.Matcher())
}

// Register registers the faq router module.
func Register() {
	router.RegisterModule("faq", nil, RegisterRoutes)
}
