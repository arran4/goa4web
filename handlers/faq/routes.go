package faq

import (
	"github.com/gorilla/mux"
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// AddFAQIndex injects FAQ index links into CoreData.
func AddFAQIndex(h http.Handler) http.Handler {
	return hcommon.IndexMiddleware(func(cd *hcommon.CoreData, r *http.Request) {
		CustomFAQIndex(cd, r)
	})(h)
}

// Task constants mirror the values used by the main package.
const (
	// TaskAsk submits a new question to the FAQ system.
	TaskAsk = "Ask"
	// TaskAnswer submits an answer in the FAQ admin interface.
	TaskAnswer = "Answer"
	// TaskRemoveRemove removes an item, typically from a list.
	TaskRemoveRemove = "Remove"
	// TaskRenameCategory renames a category.
	TaskRenameCategory = "Rename Category"
	// TaskDeleteCategory removes a category.
	TaskDeleteCategory = "Delete Category"
	// TaskCreateCategory creates a new category entry.
	TaskCreateCategory = "Create Category"
	// TaskEdit modifies an existing item.
	TaskEdit = "Edit"
	// TaskCreate indicates creation of an object.
	TaskCreate = "Create"
)

func taskMatcher(taskName string) mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		return r.PostFormValue("task") == taskName
	}
}

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
	faqr.Use(hcommon.IndexMiddleware(CustomFAQIndex))
	faqr.HandleFunc("", Page).Methods("GET", "POST")
	faqr.HandleFunc("/ask", AskPage).Methods("GET")
	faqr.HandleFunc("/ask", AskActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskAsk))
}

// RegisterAdminRoutes attaches the admin FAQ endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	farq := ar.PathPrefix("/faq").Subrouter()
	farq.Use(hcommon.IndexMiddleware(CustomFAQIndex))
	farq.HandleFunc("/answer", AdminAnswerPage).Methods("GET", "POST").MatcherFunc(noTask())
	farq.HandleFunc("/answer", AnswerAnswerActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskAnswer))
	farq.HandleFunc("/answer", AnswerRemoveActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskRemoveRemove))
	farq.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	farq.HandleFunc("/categories", CategoriesRenameActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskRenameCategory))
	farq.HandleFunc("/categories", CategoriesDeleteActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskDeleteCategory))
	farq.HandleFunc("/categories", CategoriesCreateActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskCreateCategory))
	farq.HandleFunc("/questions", AdminQuestionsPage).Methods("GET", "POST").MatcherFunc(noTask())
	farq.HandleFunc("/questions", QuestionsEditActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskEdit))
	farq.HandleFunc("/questions", QuestionsDeleteActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskRemoveRemove))
	farq.HandleFunc("/questions", QuestionsCreateActionPage).Methods("POST").MatcherFunc(taskMatcher(TaskCreate))
}

// Register registers the faq router module.
func Register() {
	router.RegisterModule("faq", nil, RegisterRoutes)
}
