package blogs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// BloggerPostsPage shows the posts written by a specific blogger.
func BloggerPostsPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	username := mux.Vars(r)["username"]
	cd.PageTitle = fmt.Sprintf("Posts by %s", username)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cd.SetBlogList(0, username, offset)
	handlers.TemplateHandler(w, r, "blogsPage", struct{}{})
}
