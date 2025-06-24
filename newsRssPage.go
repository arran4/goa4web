package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/feeds"
	"log"
	"net/http"
	"time"
)

func newsRssPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	posts, err := queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(r.Context(), GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
		Limit:  15,
		Offset: 0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	feed := &feeds.Feed{
		Title:       "News feed",
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: "Latest news posts",
		Created:     time.Now(),
	}

	for _, row := range posts {
		text := row.News.String
		conv := &A4code2html{}
		conv.codeType = ct_tagstrip
		conv.input = text
		conv.Process()
		i := len(text)
		if i > 255 {
			i = 255
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title: text[:i],
			Link:  &feeds.Link{Href: fmt.Sprintf("/news/news/%d", row.Idsitenews)},
			Created: func() time.Time {
				if row.Occured.Valid {
					return row.Occured.Time
				}
				return time.Now()
			}(),
			Description: fmt.Sprintf("%s\n-\n%s", conv.output.String(), row.Writername.String),
		})
	}

	if err := feed.WriteRss(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
