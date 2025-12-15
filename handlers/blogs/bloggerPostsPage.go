package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// BloggerPostsPage shows the posts written by a specific blogger.
func BloggerPostsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	username := mux.Vars(r)["username"]
	if _, err := cd.BloggerProfile(username); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("BloggerProfile: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	cd.PageTitle = fmt.Sprintf("Posts by %s", username)

	handlers.TemplateHandler(w, r, "blogs/bloggerPostsPage.gohtml", struct{}{})
}
