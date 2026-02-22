package news

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/gorilla/mux"
)

// SharedPreviewPage renders an OpenGraph preview for a news post.
func SharedPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Verify signature
	if share.VerifyAndGetPath(r, cd.ShareSignKey) == "" {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	newsID, _ := strconv.Atoi(vars["news"])

	foundPost, err := cd.SystemGetNewsPost(int32(newsID))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	titleLine := strings.Split(foundPost.News.String, "\n")[0]
	ogTitle := a4code.SnipText(titleLine, 100)
	ogDescription := a4code.SnipText(foundPost.News.String, 128)

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	imageURL, _ := share.MakeImageURL(cd.AbsoluteURL(), ogTitle, ogDescription, cd.ShareSignKey, false)
	ogData := share.OpenGraphData{
		Title:       ogTitle,
		Description: ogDescription,
		ImageURL:    template.URL(imageURL),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		JSONLD: map[string]interface{}{
			"@context":    "https://schema.org",
			"@type":       "NewsArticle",
			"headline":    ogTitle,
			"description": ogDescription,
			"author": map[string]interface{}{
				"@type": "Organization",
				"name":  cd.SiteTitle,
			},
		},
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
