package faq

import (
	"github.com/gorilla/mux"
	"net/http"

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
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Help", "/faq", SectionWeight)
	navReg.RegisterAdminControlCenter("Help Questions", "/admin/faq/questions", SectionWeight)
	navReg.RegisterAdminControlCenter("Help Answers", "/admin/faq/answer", SectionWeight+1)
	navReg.RegisterAdminControlCenter("Help Categories", "/admin/faq/categories", SectionWeight+2)
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
func Register(reg *router.Registry) {
	reg.RegisterModule("faq", nil, RegisterRoutes)
}
