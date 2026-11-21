package forum

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func latestReplyPosition(comments []*db.GetCommentsByThreadIdForUserRow) int64 {
	var latest int64
	for _, comment := range comments {
		if !comment.DeletedAt.Valid {
			latest = max(latest, common.CalculateActivityPosition(comment.Idcomments, comment.LastAppendID))
		}
	}
	return latest
}

// ThreadActivityFragment checks for new visible replies since the rendered snapshot.
func ThreadActivityFragment(w http.ResponseWriter, r *http.Request) {
	handlers.DisableCaching(w)
	cd := common.FromContext(r.Context())
	if cd == nil || !cd.CanLogIn() {
		http.NotFound(w, r)
		return
	}
	topic, err := cd.CurrentTopic()
	if err != nil || topic == nil || (topic.Handler == "private") != (cd.Section() == "privateforum") {
		http.NotFound(w, r)
		return
	}
	after, err := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	if err != nil || after < 0 {
		http.Error(w, "invalid displayed activity position", http.StatusBadRequest)
		return
	}
	comments, err := cd.SelectedSectionThreadComments()
	if err != nil {
		http.Error(w, "unable to check thread activity", http.StatusInternalServerError)
		return
	}
	if latestReplyPosition(comments) <= after {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct{ ThreadURL string }{ThreadURL: strings.TrimSuffix(r.URL.Path, "/live")}
	if err := cd.ExecuteSiteTemplate(w, r, "domains/forum/threadActivityFragment.gohtml", data); err != nil {
		log.Printf("render thread activity fragment: %v", err)
	}
}
