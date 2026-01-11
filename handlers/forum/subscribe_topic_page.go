package forum

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// SubscribeTopicPage renders a simple confirmation form that POSTS to the
// existing subscribe task endpoint. This allows exposing the action as a link
// in the Custom Index while keeping the state change as a POST.
func SubscribeTopicPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topic := vars["topic"]
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	type Data struct {
		BackURL string
		Method  string
		Values  url.Values
	}
	data := Data{
		BackURL: fmt.Sprintf("%s/topic/%s/subscribe", base, topic),
		Method:  http.MethodPost,
		Values:  url.Values{},
	}
	RedirectBackPageTmpl.Handle(w, r, data)
}

const RedirectBackPageTmpl handlers.Page = "redirectBackPage.gohtml"
