package blogs

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/gorilla/mux"
)

// SharedPreviewPage renders an OpenGraph preview for a blog entry.
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
	blogID, _ := strconv.Atoi(vars["blog"])

	blog, err := cd.BlogEntryByID(int32(blogID))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ogTitle := blog.Title
	if ogTitle == "" {
		ogTitle = "Blog by " + blog.Username.String
	}
	ogDescription := a4code.Snip(blog.Blog.String, 128)

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	ogData := share.OpenGraphData{
		Title:       ogTitle,
		Description: ogDescription,
		ImageURL:    template.URL(share.MakeImageURL(cd.AbsoluteURL(""), ogTitle, cd.ShareSigner, false)),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
