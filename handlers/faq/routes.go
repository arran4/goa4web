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
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []navpkg.RouterOptions {
	opts := []navpkg.RouterOptions{
		navpkg.NewIndexLinkWithViewPermission("Help", "/faq", SectionWeight, "faq", "question/answer"),
		navpkg.NewAdminControlCenterLink(navpkg.AdminCCCategory("Help"), "Help Questions", "/admin/faq/questions", SectionWeight),
		navpkg.NewAdminControlCenterLink(navpkg.AdminCCCategory("Help"), "FAQ Templates", "/admin/faq/templates", SectionWeight+1),
		navpkg.NewAdminControlCenterLink(navpkg.AdminCCCategory("Help"), "Help Categories", "/admin/faq/categories", SectionWeight+2),
	}
	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.Use(handlers.IndexMiddleware(CustomFAQIndex))
	faqr.HandleFunc("/preview", handlers.PreviewPage).Methods("POST")
	faqr.HandleFunc("", Page).Methods("GET", "POST")
	faqr.HandleFunc("/ask", askTask.Page).Methods("GET")
	faqr.HandleFunc("/ask", handlers.TaskHandler(askTask)).Methods("POST").MatcherFunc(askTask.Matcher())
	return opts
}

// RegisterAdminRoutes attaches the admin FAQ endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	farq := ar.PathPrefix("/faq").Subrouter()
	farq.Use(handlers.IndexMiddleware(CustomFAQIndex))
	farq.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	farq.HandleFunc("/categories/new", AdminNewCategoryPage).Methods("GET")
	farq.HandleFunc("/categories/category/{id:[0-9]+}", AdminCategoryPage).Methods("GET")
	farq.HandleFunc("/categories/category/{id:[0-9]+}", handlers.TaskHandler(addCategoryGrantTask)).Methods("POST").MatcherFunc(addCategoryGrantTask.Matcher())
	farq.HandleFunc("/categories/category/{id:[0-9]+}", handlers.TaskHandler(removeCategoryGrantTask)).Methods("POST").MatcherFunc(removeCategoryGrantTask.Matcher())
	farq.HandleFunc("/categories/category/{id:[0-9]+}/edit", AdminCategoryEditPage).Methods("GET")
	farq.HandleFunc("/categories/category/{id:[0-9]+}/questions", AdminCategoryQuestionsPage).Methods("GET")
	farq.HandleFunc("/categories", handlers.TaskHandler(updateCategoryTask)).Methods("POST").MatcherFunc(updateCategoryTask.Matcher())
	farq.HandleFunc("/categories", handlers.TaskHandler(deleteCategoryTask)).Methods("POST").MatcherFunc(deleteCategoryTask.Matcher())
	farq.HandleFunc("/categories", handlers.TaskHandler(createCategoryTask)).Methods("POST").MatcherFunc(createCategoryTask.Matcher())
	farq.HandleFunc("/questions", AdminQuestionsPage).Methods("GET").MatcherFunc(noTask())
	farq.HandleFunc("/questions", handlers.TaskHandler(editQuestionTask)).Methods("POST").MatcherFunc(editQuestionTask.Matcher())
	farq.HandleFunc("/questions", handlers.TaskHandler(deleteQuestionTask)).Methods("POST").MatcherFunc(deleteQuestionTask.Matcher())
	farq.HandleFunc("/questions", handlers.TaskHandler(createQuestionTask)).Methods("POST").MatcherFunc(createQuestionTask.Matcher())
	farq.HandleFunc("/question/create", AdminCreateQuestionPage).Methods("GET")
	farq.HandleFunc("/question/{id:[0-9]+}", AdminQuestionPage).Methods("GET")
	farq.HandleFunc("/question/{id:[0-9]+}/edit", AdminEditQuestionPage).Methods("GET")
	farq.HandleFunc("/revisions/{id:[0-9]+}", AdminRevisionHistoryPage).Methods("GET")
	farq.HandleFunc("/templates", AdminTemplatesPage).Methods("GET")
	farq.HandleFunc("/templates", handlers.TaskHandler(createTemplateTask)).Methods("POST").MatcherFunc(createTemplateTask.Matcher())
}

// Register registers the faq router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("faq", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []navpkg.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
