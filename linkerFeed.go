package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/feeds"
	"net/http"
	"strconv"
	"time"
)

func linkerFeed(r *http.Request, rows []*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       "Latest links",
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: "Latest submitted links",
		Created:     time.Now(),
	}
	for _, row := range rows {
		if !row.Title.Valid {
			continue
		}
		desc := ""
		if row.Description.Valid {
			var conv = &A4code2html{}
			conv.codeType = ct_tagstrip
			conv.input = row.Description.String
			conv.Process()
			desc = conv.output.String()
		}
		href := fmt.Sprintf("/linker/show/%d", row.Idlinker)
		if row.Url.Valid && row.Url.String != "" {
			href = row.Url.String
		}
		item := &feeds.Item{
			Title:   row.Title.String,
			Link:    &feeds.Link{Href: href},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", desc, func() string {
				if row.Posterusername.Valid {
					return row.Posterusername.String
				}
				return ""
			}()),
		}
		if row.Listed.Valid {
			item.Created = row.Listed.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}

func linkerRssPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	catID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	rows, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), int32(catID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	feed := linkerFeed(r, rows)
	if err := feed.WriteRss(w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAtomPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	catID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	rows, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), int32(catID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	feed := linkerFeed(r, rows)
	if err := feed.WriteAtom(w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
