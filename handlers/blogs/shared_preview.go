package blogs

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"fmt"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"net/url"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/gorilla/mux"
)

// SharedPreviewPage renders an OpenGraph preview for a blog entry.
func SharedPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	shareSignExpiry, err := time.ParseDuration(cd.Config.ShareSignExpiry)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("parsing share sign expiry: %w", err))
		return
	}
	signer := sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret, shareSignExpiry)

	// Verify signature
	if share.VerifyAndGetPath(r, signer) == "" {
		// If user is logged in, redirect to actual content URL
		if cd.UserID != 0 {
			vars := mux.Vars(r)
			blogID := vars["blog"]
			actualURL := fmt.Sprintf("/blogs/blog/%s", blogID)
			http.Redirect(w, r, actualURL, http.StatusFound)
			return
		}

		// If user is not logged in, redirect to login page with a new short-lived signed URL
		newSignedURL := signer.SignedURL(r.URL.Path)
		loginURL := fmt.Sprintf("/login?redirect_to=%s", url.QueryEscape(newSignedURL))
		http.Redirect(w, r, loginURL, http.StatusFound)
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

	tsStr := r.URL.Query().Get("ts")
	ts, _ := strconv.ParseInt(tsStr, 10, 64)
	exp := time.Now().Add(24 * time.Hour)
	if ts > 0 {
		exp = time.Unix(ts, 0)
	}

	ogData := share.OpenGraphData{
		Title:       ogTitle,
		Description: ogDescription,
		ImageURL:    template.URL(share.MakeImageURL(cd.AbsoluteURL(""), ogTitle, signer, exp)),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
