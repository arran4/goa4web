package privateforum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// UnreadThreadsPageTmpl is the template for unread private threads.
const UnreadThreadsPageTmpl tasks.Template = "domains/privateforum/unread.gohtml"

// UnreadThreadsPage serves the page listing all unread private threads.
func UnreadThreadsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	topicIDNull := sql.NullInt32{}
	topicIDVal := int32(0)
	if t := mux.Vars(r)["topic"]; t != "" {
		if val, err := strconv.Atoi(t); err == nil && val > 0 {
			topicIDNull.Valid = true
			topicIDNull.Int32 = int32(val)
			topicIDVal = int32(val)
		}
	}

	topicTitle := ""
	// Set the topic ID for breadcrumbs
	if topicIDVal > 0 {
		if top, err := cd.ForumTopicByID(topicIDVal); err == nil {
			topicTitle = cd.GetPrivateTopicDisplayTitle(topicIDVal, top.Title.String)
		}
	}

	if topicTitle != "" {
		cd.PageTitle = fmt.Sprintf("Unread Private Threads in %s", topicTitle)
	} else {
		cd.PageTitle = "Unread Private Threads"
	}

	img, err := share.MakeImageURL(cd.AbsoluteURL(), cd.PageTitle, "Unread Private discussion forums", cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}
	cd.OpenGraph = &common.OpenGraph{
		Title:       cd.PageTitle,
		Description: "Unread Private discussion forums",
		Image:       img,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.RequestURI()),
		Type:        "website",
	}

	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		handlers.RenderNotFoundOrLogin(w, r)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	limit := int32(50)
	offset := int32(page-1) * limit

	var currentError string
	rows, err := cd.UnreadPrivateThreads(limit, offset, topicIDNull, topicIDVal)
	if err != nil {
		log.Printf("Error UnreadPrivateThreads: %v", err)
		currentError = "Error loading unread threads."
	}

	// Make a slice to hold the decorated threads (e.g. for display title)
	type DecorThread struct {
		*db.ListUnreadPrivateThreadsForUserRow
		DisplayTitle string
	}

	var threads []*DecorThread
	for _, row := range rows {
		displayTitle := row.TopicTitle.String
		if row.TopicTitle.Valid {
			displayTitle = cd.GetPrivateTopicDisplayTitle(row.TopicID, row.TopicTitle.String)
		}
		threads = append(threads, &DecorThread{
			ListUnreadPrivateThreadsForUserRow: row,
			DisplayTitle:                       displayTitle,
		})
	}

	_ = UnreadThreadsPageTmpl.Handle(w, r, struct {
		Threads      []*DecorThread
		CurrentError string
		Page         int
		PrevPage     int
		NextPage     int
		HasNextPage  bool
		CD           *common.CoreData
		TopicTitle   string
	}{
		Threads:      threads,
		CurrentError: currentError,
		Page:         page,
		PrevPage:     page - 1,
		NextPage:     page + 1,
		HasNextPage:  len(rows) == int(limit),
		CD:           cd,
		TopicTitle:   topicTitle,
	})
}
