package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/a4code2html"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

func TopicFeed(r *http.Request, title string, topicID int, rows []*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: fmt.Sprintf("Latest threads for %s", title),
		Created:     time.Now(),
	}
	for _, row := range rows {
		if !row.Firstposttext.Valid {
			continue
		}
		text := row.Firstposttext.String
		conv := a4code2html.NewA4Code2HTML()
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(text)
		conv.Process()
		i := len(text)
		if i > 255 {
			i = 255
		}
		item := &feeds.Item{
			Title:   text[:i],
			Link:    &feeds.Link{Href: fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, row.Idforumthread)},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", conv.Output(), func() string {
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
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	topic, err := queries.GetForumTopicByIdForUser(r.Context(), GetForumTopicByIdForUserParams{
		UsersIdusers: uid,
		Idforumtopic: int32(topicID),
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetForumTopicByIdForUser error: %s", err)
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rows, err := queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(r.Context(), GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: int32(topicID),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	feed := TopicFeed(r, topic.Title.String, topicID, rows)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func TopicAtomPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	topic, err := queries.GetForumTopicByIdForUser(r.Context(), GetForumTopicByIdForUserParams{
		UsersIdusers: uid,
		Idforumtopic: int32(topicID),
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetForumTopicByIdForUser error: %s", err)
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rows, err := queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(r.Context(), GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: int32(topicID),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	feed := TopicFeed(r, topic.Title.String, topicID, rows)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
