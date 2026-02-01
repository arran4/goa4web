package tasks

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Task describes an application action that may trigger notifications.
// It provides handlers for performing the action and rendering the result as
// well as a factory for creating an EventNotification.
type Task interface {
	// Action results are either:
	// * http.Handler / http.HandlerFunc
	// * type Template string
	// * error
	Action(w http.ResponseWriter, r *http.Request) any
}

// TemplatesRequired can be implemented by Task types that render templates
// to declare the exact site template paths they depend on. Tests can use this
// to verify that the templates exist.
type TemplatesRequired interface {
	RequiredTemplates() []Template
}

type EmailTemplatesRequired interface {
	RequiredTemplates() []Template
}

type TaskMatcher interface {
	Matcher() mux.MatcherFunc
}

type Name interface {
	Name() string
}

type TaskString string

func (t TaskString) Name() string {
	return string(t)
}

func (t TaskString) Action(http.ResponseWriter, *http.Request) any { return nil }

func (t TaskString) Matcher() mux.MatcherFunc {
	return HasTask(string(t))
}

var _ TaskMatcher = (TaskString)("")
var _ Name = (TaskString)("")
var _ Task = (TaskString)("")

type HasBreadcrumb interface {
	Breadcrumb() (name, link string, parent HasBreadcrumb)
}

type HasPageTitle interface {
	PageTitle() string
}
