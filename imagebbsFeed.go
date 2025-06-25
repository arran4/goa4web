package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/a4code2html"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func imagebbsFeed(r *http.Request, title string, boardID int, rows []*GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountRow) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: fmt.Sprintf("Latest posts for %s", title),
		Created:     time.Now(),
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Posted.Time.After(rows[j].Posted.Time)
	})
	if len(rows) > 15 {
		rows = rows[:15]
	}
	for _, row := range rows {
		if !row.Description.Valid {
			continue
		}
		desc := row.Description.String
		conv := a4code2html.NewA4Code2HTML()
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(desc)
		conv.Process()
		i := len(desc)
		if i > 255 {
			i = 255
		}
		item := &feeds.Item{
			Title:   desc[:i],
			Link:    &feeds.Link{Href: fmt.Sprintf("/imagebbs/board/%d/thread/%d", boardID, row.ForumthreadIdforumthread)},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", conv.Output(), func() string {
				if row.Username.Valid {
					return row.Username.String
				}
				return ""
			}()),
		}
		if row.Posted.Valid {
			item.Created = row.Posted.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}

func imagebbsRssPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	boards, err := queries.GetAllImageBoards(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query boards error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var posts []*GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountRow
	for _, b := range boards {
		rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("feed query error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, rows...)
	}
	feed := imagebbsFeed(r, "ImageBBS", 0, posts)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsAtomPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	boards, err := queries.GetAllImageBoards(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query boards error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var posts []*GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountRow
	for _, b := range boards {
		rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("feed query error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, rows...)
	}
	feed := imagebbsFeed(r, "ImageBBS", 0, posts)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsBoardRssPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount(r.Context(), int32(bid))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	title := fmt.Sprintf("Board %d", bid)
	boards, err := queries.GetAllImageBoards(r.Context())
	if err == nil {
		for _, b := range boards {
			if int(b.Idimageboard) == bid {
				if b.Title.Valid {
					title = b.Title.String
				}
				break
			}
		}
	}
	feed := imagebbsFeed(r, title, bid, rows)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsBoardAtomPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount(r.Context(), int32(bid))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	title := fmt.Sprintf("Board %d", bid)
	boards, err := queries.GetAllImageBoards(r.Context())
	if err == nil {
		for _, b := range boards {
			if int(b.Idimageboard) == bid {
				if b.Title.Valid {
					title = b.Title.String
				}
				break
			}
		}
	}
	feed := imagebbsFeed(r, title, bid, rows)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
