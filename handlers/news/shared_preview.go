package news

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/gorilla/mux"
)

// SharedPreviewPage renders an OpenGraph preview for a news post.
func SharedPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Create signer from config
	signer := sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret)

	// Verify signature
	if share.VerifyAndGetPath(r, signer) == "" {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	newsID, _ := strconv.Atoi(vars["news"])

	posts, err := cd.LatestNewsList(0, 1000)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var foundPost *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	for _, p := range posts {
		if p.Idsitenews == int32(newsID) {
			foundPost = p
			break
		}
	}

	if foundPost == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ogTitle := a4code.Snip(foundPost.News.String, 100)
	ogDescription := a4code.Snip(foundPost.News.String, 128)

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	ogData := share.OpenGraphData{
		Title:       ogTitle,
		Description: ogDescription,
		ImageURL:    template.URL(share.MakeImageURL(cd.AbsoluteURL(""), ogTitle, signer)),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.Path)),
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
