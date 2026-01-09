package privateforum

import (
	"fmt"
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

// SharedPreviewPage renders an OpenGraph preview for a private forum thread.
// This endpoint bypasses access control to allow social media bots to see metadata.
func SharedPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
	topicID, _ := strconv.Atoi(vars["topic"])

	// If user is logged in, redirect to actual content URL
	if cd.UserID != 0 {
		actualURL := fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)
		http.Redirect(w, r, actualURL, http.StatusFound)
		return
	}

	// For non-authenticated users, fetch metadata and show login page with OG tags

	// Use admin queries to bypass access control for OpenGraph previews
	queries := cd.Queries()
	thread, err := queries.AdminGetForumThreadById(r.Context(), int32(threadID))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	topic, err := queries.GetForumTopicById(r.Context(), thread.Idforumtopic)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Get first comment for description
	comments, err := queries.SystemListCommentsByThreadID(r.Context(), int32(threadID))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	ogTitle := topic.Title.String
	ogDescription := ""
	if len(comments) > 0 {
		ogDescription = a4code.Snip(comments[0].Text.String, 128)
	}

	// Redirect URL after successful login
	redirectURL := fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	cd.OpenGraph = &common.OpenGraph{
		Title:       ogTitle,
		Description: ogDescription,
		Image:       share.MakeImageURL(cd.AbsoluteURL(""), ogTitle, cd.ShareSigner),
		URL:         cd.AbsoluteURL(redirectURL),
	}

	// Render login page with OpenGraph metadata
	handlers.TemplateHandler(w, r, "sharedPreviewLogin.gohtml", struct {
		RedirectURL string
	}{
		RedirectURL: url.QueryEscape(redirectURL),
	})
}
