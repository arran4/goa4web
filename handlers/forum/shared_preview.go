package forum

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// SharedThreadPreviewPage renders an OpenGraph preview for a forum thread.
func SharedThreadPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Verify signature
	if share.VerifyAndGetPath(r, cd.ShareSignKey) == "" {
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
	if t, err := getPrivateTopicTitle(r.Context(), queries, topic); err == nil {
		ogTitle = t
	}

	ogDescription := ""
	if len(comments) > 0 {
		ogDescription = a4code.SnipText(comments[0].Text.String, 128)
	}

	user := cd.UserByID(thread.CreatedBy)
	authorName := "Unknown"
	if user != nil && user.Username.Valid {
		authorName = user.Username.String
	}

	renderPublicSharedPreview(w, r, cd,
		share.WithTitle(ogTitle),
		share.WithBody(ogDescription),
		share.WithSection("Public Forum Thread"),
		share.WithGeneratorType("forum"),
		share.WithJSONLD{Data: map[string]interface{}{
			"@context":      "https://schema.org",
			"@type":         "DiscussionForumPosting",
			"headline":      ogTitle,
			"description":   ogDescription,
			"datePublished": thread.CreatedAt.Time.Format(time.RFC3339),
			"author": map[string]interface{}{
				"@type": "Person",
				"name":  authorName,
			},
			"interactionStatistic": map[string]interface{}{
				"@type":                "InteractionCounter",
				"interactionType":      "https://schema.org/CommentAction",
				"userInteractionCount": thread.PostCount.Int32,
			},
		}},
	)
}

// SharedTopicPreviewPage renders an OpenGraph preview for a forum topic.
func SharedTopicPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if share.VerifyAndGetPath(r, cd.ShareSignKey) == "" {
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
	if t, err := getPrivateTopicTitle(r.Context(), queries, topic); err == nil {
		ogTitle = t
	}
	ogDescription := topic.Description.String

	renderPublicSharedPreview(w, r, cd,
		share.WithTitle(ogTitle),
		share.WithBody(ogDescription),
		share.WithSection("Public Forum Topic"),
		share.WithGeneratorType("forum"),
		share.WithJSONLD{Data: map[string]interface{}{
			"@context":    "https://schema.org",
			"@type":       "DiscussionForumPosting",
			"headline":    ogTitle,
			"description": ogDescription,
			"author": map[string]interface{}{
				"@type": "Organization",
				"name":  cd.SiteTitle,
			},
		}},
	)
}

func renderPublicSharedPreview(w http.ResponseWriter, r *http.Request, cd *common.CoreData, ops ...interface{}) {
	// Determine auth style: check if mux vars for ts/nonce are present
	vars := mux.Vars(r)
	usePathAuth := vars["ts"] != "" || vars["nonce"] != ""

	imageURL, _ := share.MakeImageURLWithOptions(cd.AbsoluteURL(), cd.ShareSignKey, usePathAuth, ops...)

	// If the user is viewing this, they are likely a guest (or the caller logic didn't redirect them).
	// We want to redirect guests to login, then back to here.
	redirectURL := "/login?return_url=" + url.QueryEscape(r.URL.RequestURI())

	var title, desc string
	var jsonLD interface{}
	for _, op := range ops {
		switch v := op.(type) {
		case share.WithTitle:
			title = string(v)
		case share.WithBody:
			desc = string(v)
		case share.WithJSONLD:
			jsonLD = v.Data
		}
	}

	ogData := share.OpenGraphData{
		Title:       title,
		Description: desc,
		JSONLD:      jsonLD,
		ImageURL:    template.URL(imageURL),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		RedirectURL: template.URL(redirectURL),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func getPrivateTopicTitle(ctx context.Context, queries db.Querier, topic *db.Forumtopic) (string, error) {
	if topic.ForumcategoryIdforumcategory != common.PrivateForumCategoryID {
		return topic.Title.String, nil
	}
	parts, err := queries.AdminListPrivateTopicParticipantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
	if err != nil {
		return "", err
	}
	var names []string
	for _, part := range parts {
		if part.Username.Valid {
			names = append(names, part.Username.String)
		}
	}
	sort.Strings(names)
	n := len(names)
	if n == 0 {
		return "Private forum", nil
	}
	if n == 1 {
		return fmt.Sprintf("Private forum with %s", names[0]), nil
	}
	last := names[n-1]
	rest := names[:n-1]
	return fmt.Sprintf("Private forum with %s & %s", strings.Join(rest, ", "), last), nil
}

const SharedPreviewLoginPageTmpl tasks.Template = "sharedPreviewLogin.gohtml"
