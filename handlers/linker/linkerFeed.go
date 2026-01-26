package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

func linkerFeed(r *http.Request, rows []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow) *feeds.Feed {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	title := "Latest links"
	if cd.SiteTitle != "" {
		title = fmt.Sprintf("%s - %s", cd.SiteTitle, title)
	}
	feed := &feeds.Feed{
		Title:       title,
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
			conv := a4code2html.New(cd.ImageURLMapper)
			conv.CodeType = a4code2html.CTTagStrip
			conv.SetInput(row.Description.String)
			out, _ := io.ReadAll(conv.Process())
			desc = string(out)
		}
		href := fmt.Sprintf("/linker/show/%d", row.ID)
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

func RssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if _, ok := mux.Vars(r)["username"]; ok {
		u, err := handlers.VerifyFeedRequest(r, "/linker/rss")
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		cd.UserID = u.Idusers
	}

	queries := cd.Queries()
	catID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	rows, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{CategoryID: int32(catID)})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	feed := linkerFeed(r, rows)
	if err := feed.WriteRss(w); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func AtomPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if _, ok := mux.Vars(r)["username"]; ok {
		u, err := handlers.VerifyFeedRequest(r, "/linker/atom")
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		cd.UserID = u.Idusers
	}

	queries := cd.Queries()
	catID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	rows, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{CategoryID: int32(catID)})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	feed := linkerFeed(r, rows)
	if err := feed.WriteAtom(w); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}
