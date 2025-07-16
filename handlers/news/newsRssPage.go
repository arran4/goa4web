package news

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/feeds"

	"github.com/arran4/goa4web/a4code/a4code2html"
	hcommon "github.com/arran4/goa4web/handlers/common"
	imageshandler "github.com/arran4/goa4web/handlers/images"
)

func NewsRssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	posts, err := cd.LatestNews(r)
	if err != nil {
		log.Printf("latestNews: %v", err)
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
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		text := row.News.String
		conv := a4code2html.New(imageshandler.MapURL)
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(text)
		out, _ := io.ReadAll(conv.Process())
		i := len(text)
		if i > 255 {
			i = 255
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title: text[:i],
			Link:  &feeds.Link{Href: fmt.Sprintf("/news/news/%d", row.Idsitenews)},
			Created: func() time.Time {
				if row.Occurred.Valid {
					return row.Occurred.Time
				}
				return time.Now()
			}(),
			Description: fmt.Sprintf("%s\n-\n%s", string(out), row.Writername.String),
		})
	}

	if err := feed.WriteRss(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
