package privateforum

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/gorilla/mux"
)

// SharedThreadPreviewPage renders an OpenGraph preview for a private forum thread.
// It verifies the signature before displaying any metadata.
func SharedThreadPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if share.VerifyAndGetPath(r, cd.ShareSignKey) == "" {
		log.Printf("[Share] Invalid signature for URL: %s", r.URL.String())
		// No valid signature? If user is logged in, redirect to actual content (they might have perm).
		// If not logged in, show 403.
		if cd.UserID != 0 {
			vars := mux.Vars(r)
			actualURL := fmt.Sprintf("/private/topic/%s/thread/%s", vars["topic"], vars["thread"])
			http.Redirect(w, r, actualURL, http.StatusFound)
			return
		}
		handlers.RenderErrorPage(w, r, handlers.WrapForbidden(fmt.Errorf("invalid signature")))
		return
	}

	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
	topicID, _ := strconv.Atoi(vars["topic"])

	// If user is logged in, redirect to actual content URL
	if cd.UserID != 0 {
		actualURL := fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)
		http.Redirect(w, r, actualURL, http.StatusFound)
		return
	}

	// For non-authenticated users with VALID SIGNATURE, fetch metadata and show login page with OG tags
	queries := cd.Queries()
	thread, err := queries.AdminGetForumThreadById(r.Context(), int32(threadID))
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.WrapNotFound(err))
		return
	}

	topic, err := queries.GetForumTopicById(r.Context(), thread.Idforumtopic)
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.WrapNotFound(err))
		return
	}

	// Get first comment for description
	comments, err := queries.SystemListCommentsByThreadID(r.Context(), int32(threadID))
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.WrapNotFound(err))
		return
	}

	ogTitle := topic.Title.String
	ogDescription := ""
	if len(comments) > 0 {
		ogDescription = a4code.Snip(comments[0].Text.String, 128)
	}

	renderSharedPreview(w, r, cd, ogTitle, ogDescription, fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID))
}

// SharedTopicPreviewPage renders an OpenGraph preview for a private forum topic.
func SharedTopicPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Verify signature
	if share.VerifyAndGetPath(r, cd.ShareSignKey) == "" {
		log.Printf("[Share] Invalid signature for URL: %s", r.URL.String())
		if cd.UserID != 0 {
			vars := mux.Vars(r)
			actualURL := fmt.Sprintf("/private/topic/%s", vars["topic"])
			http.Redirect(w, r, actualURL, http.StatusFound)
			return
		}
		handlers.RenderErrorPage(w, r, handlers.WrapForbidden(fmt.Errorf("invalid signature")))
		return
	}

	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])

	if cd.UserID != 0 {
		actualURL := fmt.Sprintf("/private/topic/%d", topicID)
		http.Redirect(w, r, actualURL, http.StatusFound)
		return
	}

	queries := cd.Queries()
	topic, err := queries.GetForumTopicById(r.Context(), int32(topicID))
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.WrapNotFound(err))
		return
	}

	ogTitle := topic.Title.String
	ogDescription := topic.Description.String

	renderSharedPreview(w, r, cd, ogTitle, ogDescription, fmt.Sprintf("/private/topic/%d", topicID))
}

func renderSharedPreview(w http.ResponseWriter, r *http.Request, cd *common.CoreData, title, desc, redirectPath string) {

	// Determine auth style: check if mux vars for ts/nonce are present
	vars := mux.Vars(r)
	usePathAuth := vars["ts"] != "" || vars["nonce"] != ""

	// tsVal is CREATION TIME of the share link (if ts used). Do not use as expiration.
	// Generate a fresh expiration for the image link.

	// Calculate image URL with error handling
	imgURL, err := share.MakeImageURL(cd.AbsoluteURL(), title, desc, cd.ShareSignKey, usePathAuth)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       title,
		Description: desc,
		Image:       imgURL,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.RequestURI()),
		Type:        "website",
	}

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	SharedPreviewLoginPageTmpl.Handle(w, r, struct {
		RedirectURL string
	}{
		RedirectURL: url.QueryEscape(redirectPath),
	})
}

const SharedPreviewLoginPageTmpl handlers.Page = "sharedPreviewLogin.gohtml"
