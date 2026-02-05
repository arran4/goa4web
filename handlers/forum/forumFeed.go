package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

func TopicFeed(r *http.Request, title string, topicID int, rows []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, basePath string) *feeds.Feed {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	feedTitle := title
	if cd.SiteTitle != "" {
		feedTitle = fmt.Sprintf("%s - %s", cd.SiteTitle, title)
	}
	feed := &feeds.Feed{
		Title:       feedTitle,
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: fmt.Sprintf("Latest threads for %s", title),
		Created:     time.Now(),
	}
	for _, row := range rows {
		if !row.Firstposttext.Valid {
			continue
		}
		text := row.Firstposttext.String
		conv := a4code2html.New(cd.ImageURLMapper)
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(text)
		out, _ := io.ReadAll(conv.Process())
		i := len(text)
		if i > 255 {
			i = 255
		}
		item := &feeds.Item{
			Title:   text[:i],
			Link:    &feeds.Link{Href: fmt.Sprintf("%s/topic/%d/thread/%d", basePath, topicID, row.Idforumthread)},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", string(out), func() string {
				if row.Firstpostusername.Valid {
					return row.Firstpostusername.String
				}
				return ""
			}()),
		}
		if row.Firstpostwritten.Valid {
			item.Created = row.Firstpostwritten.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}

func TopicRssPage(w http.ResponseWriter, r *http.Request) {
	handleTopicFeed(w, r, "rss")
}

func TopicAtomPage(w http.ResponseWriter, r *http.Request) {
	handleTopicFeed(w, r, "atom")
}

func handleTopicFeed(w http.ResponseWriter, r *http.Request, feedType string) {
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	basePath := fmt.Sprintf("/forum/topic/%d.%s", topicID, feedType)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if _, ok := vars["username"]; ok {
		user, err := handlers.VerifyFeedRequest(r, basePath)
		if err == nil && user != nil {
			cd.UserID = user.Idusers
		}
	} else if _, ok := core.GetSessionOrFail(w, r); ok {
		// Session loaded in CoreData (via IndexMiddleware / WithSession)
		// No op, cd.UserID already set
	}

	topic, err := cd.ForumTopicByID(int32(topicID))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetForumTopicByIdForUser error: %s", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - %s Feed", topic.Title.String)
	rows, err := cd.ForumThreads(int32(topicID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	feed := TopicFeed(r, topic.Title.String, topicID, rows, "/forum")
	if feedType == "rss" {
		if err := feed.WriteRss(w); err != nil {
			log.Printf("feed write error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	} else {
		if err := feed.WriteAtom(w); err != nil {
			log.Printf("feed write error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}
}
