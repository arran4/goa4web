package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/a4code2html"
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/feeds"
	"log"
	"net/http"
	"time"
)

func feedGen(r *http.Request, queries *db.Queries) (*feeds.Feed, error) {
	feed := &feeds.Feed{
		Title:       "Latest writings",
		Link:        &feeds.Link{Href: r.URL.String()},
		Description: "recent writings",
		Created:     time.Now(),
	}

	rows, err := queries.GetPublicWritings(r.Context(), db.GetPublicWritingsParams{Limit: 15, Offset: 0})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	for _, row := range rows {
		desc := row.Abstract.String
		if desc == "" {
			desc = row.Writing.String
		}
		conv := a4code2html.NewA4Code2HTML()
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(desc)
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
			Description: conv.Output(),
		}
		if row.Published.Valid {
			item.Created = row.Published.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed, nil
}

func RssPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	feed, err := feedGen(r, queries)
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

func AtomPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	feed, err := feedGen(r, queries)
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
