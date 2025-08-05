package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"io"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/gorilla/feeds"
)

func Page(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	buid := r.URL.Query().Get("uid")
	userID, _ := strconv.Atoi(buid)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Blogs"
	cd.SetBlogListParams(int32(userID), offset)

	ps := cd.PageSize()
	qv := r.URL.Query()
	qv.Set("offset", strconv.Itoa(offset+ps))
	cd.NextLink = "/blogs?" + qv.Encode()
	if offset > 0 {
		qv.Set("offset", strconv.Itoa(offset-ps))
		cd.PrevLink = "/blogs?" + qv.Encode()
	}

	handlers.TemplateHandler(w, r, "blogsPage", struct{}{})
}

func CustomBlogIndex(data *common.CoreData, r *http.Request) {
	user := r.URL.Query().Get("user")
	data.CustomIndexItems = []common.IndexItem{}
	if data.FeedsEnabled {
		suffix := ""
		if user != "" {
			suffix = "?user=" + url.QueryEscape(user)
		}
		data.RSSFeedURL = "/blogs/rss" + suffix
		data.AtomFeedURL = "/blogs/atom" + suffix
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{Name: "Atom Feed", Link: data.AtomFeedURL},
			common.IndexItem{Name: "RSS Feed", Link: data.RSSFeedURL},
		)
	}

	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "User Roles",
			Link: "/admin/blogs/users/roles",
		})
	}
	userHasWriter := data.HasRole("content writer")
	if userHasWriter {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Write blog",
			Link: "/blogs/add",
		})

	}
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "List bloggers",
		Link: "/blogs/bloggers",
	})
}

func RssPage(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("rss")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
	if err != nil {
		log.Printf("Username to uid error: %s", err)
	}
	uid := u.Idusers
	feed, err := FeedGen(r, queries, int(uid), username)
	if err != nil {
		log.Printf("FeedGen Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	if err := feed.WriteRss(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func AtomPage(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("rss")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
	if err != nil {
		log.Printf("Username to uid error: %s", err)
	}
	feed, err := FeedGen(r, queries, int(u.Idusers), username)
	if err != nil {
		log.Printf("FeedGen Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func FeedGen(r *http.Request, queries db.Querier, uid int, username string) (*feeds.Feed, error) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	title := "Everyone's blog"
	if uid > 0 {
		title = fmt.Sprintf("%s blog", username)
	}
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: r.URL.String()},
		Description: "discussion about tech, footie, photos",
		Author:      &feeds.Author{Name: username, Email: "n@a"},
		Created:     time.Date(2005, 6, 25, 0, 0, 0, 0, time.UTC),
	}

	rows, err := queries.ListBlogEntriesByAuthorForLister(r.Context(), db.ListBlogEntriesByAuthorForListerParams{
		AuthorID: int32(uid),
		ListerID: int32(uid),
		UserID:   sql.NullInt32{Int32: int32(uid), Valid: uid != 0},
		Limit:    15,
		Offset:   0,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			return nil, err
		}
	}

	for _, row := range rows {
		u := r.URL
		u.Query().Set("show", fmt.Sprintf("%d", row.Idblogs))
		conv := a4code2html.New(cd.ImageURLMapper)
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(row.Blog.String)
		out, _ := io.ReadAll(conv.Process())
		i := len(row.Blog.String)
		if i > 255 {
			i = 255
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title: row.Blog.String[:i],
			Link: &feeds.Link{
				Href: u.String(),
			},
			Description: fmt.Sprintf("%s\n-\n%s", string(out), row.Username.String),
		})
	}
	return feed, nil
}
