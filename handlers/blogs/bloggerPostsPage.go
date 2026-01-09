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

	"time"

	"github.com/arran4/goa4web/handlers/share"
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
	cd.OpenGraph = &common.OpenGraph{
		Title:       cd.PageTitle,
		Description: fmt.Sprintf("View posts by %s", username),
		Image:       share.MakeImageURL(cd.AbsoluteURL(""), cd.PageTitle, cd.ShareSigner, time.Now().Add(24*time.Hour)),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
	}

	handlers.TemplateHandler(w, r, "blogs/bloggerPostsPage.gohtml", struct{}{})
}
