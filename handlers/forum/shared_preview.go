package forum

import (
	"fmt"
	"html/template"
	"log"
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

// SharedThreadPreviewPage renders an OpenGraph preview for a forum thread.
func SharedThreadPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Create signer from config
	signer := sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret)

	// Verify signature
	if share.VerifyAndGetPath(r, signer) == "" {
		log.Printf("[Forum Share] Invalid signature for URL: %s", r.URL.String())
		handlers.RenderErrorPage(w, r, handlers.WrapForbidden(fmt.Errorf("invalid signature")))
		return
	}

	vars := mux.Vars(r)
	threadID, _ := strconv.Atoi(vars["thread"])
	topicID, _ := strconv.Atoi(vars["topic"])

	// If user is logged in, redirect to actual content URL
	if cd.UserID != 0 {
		actualURL := fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, threadID)
		http.Redirect(w, r, actualURL, http.StatusFound)
		return
	}

	// Use admin queries to bypass access control for OpenGraph previews
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

	renderPublicSharedPreview(w, r, cd, signer, ogTitle, ogDescription)
}

// SharedTopicPreviewPage renders an OpenGraph preview for a forum topic.
func SharedTopicPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	signer := sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret)

	if share.VerifyAndGetPath(r, signer) == "" {
		log.Printf("[Forum Share] Invalid signature for URL: %s", r.URL.String())
		handlers.RenderErrorPage(w, r, handlers.WrapForbidden(fmt.Errorf("invalid signature")))
		return
	}

	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])

	// If user is logged in, redirect to actual content URL
	if cd.UserID != 0 {
		actualURL := fmt.Sprintf("/forum/topic/%d", topicID)
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

	renderPublicSharedPreview(w, r, cd, signer, ogTitle, ogDescription)
}

func renderPublicSharedPreview(w http.ResponseWriter, r *http.Request, cd *common.CoreData, signer *sharesign.Signer, title, desc string) {
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	tsStr := r.URL.Query().Get("ts")
	tsVal, _ := strconv.ParseInt(tsStr, 10, 64)
	if tsVal == 0 {
		vars := mux.Vars(r)
		if t, err := strconv.ParseInt(vars["ts"], 10, 64); err == nil {
			tsVal = t
		}
	}

	usePathAuth := mux.Vars(r)["ts"] != ""

	ogData := share.OpenGraphData{
		Title:       title,
		Description: desc,
		ImageURL:    template.URL(share.MakeImageURL(cd.AbsoluteURL(""), title, signer, usePathAuth)),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
