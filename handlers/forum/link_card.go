package forum

import (
	"bytes"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

// LinkCardFragment renders current metadata for a block link in a visible comment.
func LinkCardFragment(w http.ResponseWriter, r *http.Request) {
	handlers.DisableCaching(w)
	cd := common.FromContext(r.Context())
	if cd == nil || cd.UserID == 0 {
		http.NotFound(w, r)
		return
	}
	topic, err := cd.CurrentTopic()
	if err != nil || topic == nil || (topic.Handler == "private") != (cd.Section() == "privateforum") {
		http.NotFound(w, r)
		return
	}
	key := r.URL.Query().Get("card")
	card, err := cd.ForumLinkCardForViewer(r.Context(), cd.SelectedThreadID(), key)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "unable to load link preview", http.StatusInternalServerError)
		return
	}
	var body bytes.Buffer
	if err := cd.RenderForumLinkCard(&body, r, key, strings.TrimSuffix(r.URL.Path, "/link-preview"), card); err != nil {
		log.Printf("render external link card fragment: %v", err)
		http.Error(w, "unable to render link preview", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = body.WriteTo(w)
}
