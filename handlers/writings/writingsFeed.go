package writings

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/a4code/a4code2html"
	common "github.com/arran4/goa4web/core/common"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/gorilla/feeds"
)

func feedGen(r *http.Request, cd *common.CoreData) (*feeds.Feed, error) {
	feed := &feeds.Feed{
		Title:       "Latest writings",
		Link:        &feeds.Link{Href: r.URL.String()},
		Description: "recent writings",
		Created:     time.Now(),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	rows, err := cd.LatestWritings(common.WithWritingsOffset(int32(offset)))
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		desc := row.Abstract.String
		if desc == "" {
			desc = row.Writing.String
		}
		conv := a4code2html.New(imagesign.MapURL)
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(desc)
		out, _ := io.ReadAll(conv.Process())
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
			Description: string(out),
		}
		if row.Published.Valid {
			item.Created = row.Published.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed, nil
}

func RssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	feed, err := feedGen(r, cd)
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
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	feed, err := feedGen(r, cd)
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
