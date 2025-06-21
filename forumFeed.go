package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
	"time"
)

func forumTopicFeed(r *http.Request, title string, topicID int, rows []*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow) *feeds.Feed {
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
		var conv = &A4code2html{}
		conv.codeType = ct_tagstrip
		conv.input = text
		conv.Process()
		i := len(text)
		if i > 255 {
			i = 255
		}
		item := &feeds.Item{
			Title:   text[:i],
			Link:    &feeds.Link{Href: fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, row.Idforumthread)},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", conv.output.String(), func() string {
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

func forumTopicRssPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

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

	feed := forumTopicFeed(r, topic.Title.String, topicID, rows)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumTopicAtomPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicID, _ := strconv.Atoi(vars["topic"])
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

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

	feed := forumTopicFeed(r, topic.Title.String, topicID, rows)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
