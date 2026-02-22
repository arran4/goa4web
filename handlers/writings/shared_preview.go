package writings

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/gorilla/mux"
)

// SharedPreviewPage renders an OpenGraph preview for a writing article.
func SharedPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Verify signature
	if share.VerifyAndGetPath(r, cd.ShareSignKey) == "" {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	writingID, _ := strconv.Atoi(vars["article"])

	writing, err := cd.WritingByID(int32(writingID))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ogTitle := writing.Title.String
	ogDescription := a4code.SnipText(writing.Abstract.String, 128)

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	imgURL, err := share.MakeImageURL(cd.AbsoluteURL(), ogTitle, ogDescription, cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}

	ogData := share.OpenGraphData{
		Title:       ogTitle,
		Description: ogDescription,
		ImageURL:    template.URL(imgURL),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		JSONLD: &common.JSONLD{
			Context:       "https://schema.org",
			Type:          "Article",
			Headline:      ogTitle,
			Description:   ogDescription,
			DatePublished: writing.Published.Time.Format(time.RFC3339),
			Author: &common.JSONLDAuthor{
				Type: "Person",
				Name: writing.Writerusername.String,
			},
		},
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
