package privateforum

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// UnreadThreadsPageTmpl is the template for unread private threads.
const UnreadThreadsPageTmpl tasks.Template = "privateforum/unread.gohtml"

// UnreadThreadsPage serves the page listing all unread private threads.
func UnreadThreadsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	cd.PageTitle = "Unread Private Threads"
	img, err := share.MakeImageURL(cd.AbsoluteURL(), "Unread Private Threads", "Unread Private discussion forums", cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}
	cd.OpenGraph = &common.OpenGraph{
		Title:       "Unread Private Threads",
		Description: "Unread Private discussion forums",
		Image:       img,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.RequestURI()),
		Type:        "website",
	}

	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		SharedPreviewLoginPageTmpl.Handle(w, r, struct {
			RedirectURL string
		}{
			RedirectURL: url.QueryEscape(r.URL.RequestURI()),
		})
		return
	}

	var currentError string
	rows, err := cd.Queries().ListUnreadPrivateThreadsForUser(r.Context(), db.ListUnreadPrivateThreadsForUserParams{
		UserIDNull: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		UserIDVal:  cd.UserID,
	})
	if err != nil {
		log.Printf("Error ListUnreadPrivateThreadsForUser: %v", err)
		currentError = "Error loading unread threads."
	}

	UnreadThreadsPageTmpl.Handle(w, r, struct {
		Threads []*db.ListUnreadPrivateThreadsForUserRow
		CurrentError string
	}{
		Threads: rows,
		CurrentError: currentError,
	})
}
