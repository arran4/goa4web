package forumcommon

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/internal/tasks"
)

const RedirectBackPageTmpl tasks.Template = "redirectBackPage.gohtml"

// SubscribeTopicPage renders a confirmation form for subscribing to a topic.
func (f *ForumContext) SubscribeTopicPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]

	type Data struct {
		BackURL string
		Method  string
		Values  url.Values
	}
	data := Data{
		BackURL: fmt.Sprintf("%s/topic/%s/subscribe", f.BasePath, topic),
		Method:  http.MethodPost,
		Values:  url.Values{},
	}
	RedirectBackPageTmpl.Handle(w, r, data)
}

// UnsubscribeTopicPage renders a confirmation form for unsubscribing from a topic.
func (f *ForumContext) UnsubscribeTopicPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]

	type Data struct {
		BackURL string
		Method  string
		Values  url.Values
	}
	data := Data{
		BackURL: fmt.Sprintf("%s/topic/%s/unsubscribe", f.BasePath, topic),
		Method:  http.MethodPost,
		Values:  url.Values{},
	}
	RedirectBackPageTmpl.Handle(w, r, data)
}
