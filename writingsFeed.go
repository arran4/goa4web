package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/feeds"
	"log"
	"net/http"
	"time"
)

func writingsFeedGen(r *http.Request, queries *Queries) (*feeds.Feed, error) {
	feed := &feeds.Feed{
		Title:       "Latest writings",
		Link:        &feeds.Link{Href: r.URL.String()},
		Description: "recent writings",
		Created:     time.Now(),
	}

	rows, err := queries.GetPublicWritings(r.Context(), GetPublicWritingsParams{Limit: 15, Offset: 0})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	for _, row := range rows {
		desc := row.Abstract.String
		if desc == "" {
			desc = row.Writting.String
		}
		conv := &A4code2html{codeType: ct_tagstrip, input: desc}
		conv.Process()
		title := row.Title.String
		if title == "" {
			if len(desc) > 20 {
				title = desc[:20]
			} else {
				title = desc
			}
		}
		item := &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: fmt.Sprintf("/writings/article/%d", row.Idwriting)},
			Created:     time.Now(),
			Description: conv.output.String(),
		}
		if row.Published.Valid {
			item.Created = row.Published.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed, nil
}

func writingsRssPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	feed, err := writingsFeedGen(r, queries)
	if err != nil {
		log.Printf("FeedGen Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := feed.WriteRss(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAtomPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	feed, err := writingsFeedGen(r, queries)
	if err != nil {
		log.Printf("FeedGen Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
