package news

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/handlers"
)

func NewsRssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if _, ok := mux.Vars(r)["username"]; ok {
		u, err := handlers.VerifyFeedRequest(r, "/news/rss")
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		cd.UserID = u.Idusers
	}

	posts, err := cd.LatestNews()
	if err != nil {
		log.Printf("latestNews: %v", err)
		handlers.RenderErrorPage(w, r, err)
		return
	}

	title := "News feed"
	if cd.SiteTitle != "" {
		title = fmt.Sprintf("%s - %s", cd.SiteTitle, title)
	}
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: "Latest news posts",
		Created:     time.Now(),
	}

	for _, row := range posts {
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		text := row.News.String
		conv := a4code2html.New(cd.ImageURLMapper)
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
		handlers.RenderErrorPage(w, r, err)
		return
	}
}
