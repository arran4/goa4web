package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/arran4/goa4web/a4code2html"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/feeds"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*GetBlogEntriesForUserDescendingLanguagesRow
		EditUrl string
	}
	type Data struct {
		*corecommon.CoreData
		Rows     []*BlogRow
		IsOffset bool
		UID      string
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	buid := r.URL.Query().Get("uid")
	userId, _ := strconv.Atoi(buid)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.GetBlogEntriesForUserDescendingLanguages(r.Context(), GetBlogEntriesForUserDescendingLanguagesParams{
		UsersIdusers:  int32(userId),
		ViewerIdusers: uid,
		Limit:         15,
		Offset:        int32(offset),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Query Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		IsOffset: offset != 0,
		UID:      buid,
	}

	for _, row := range rows {
		editUrl := ""
		if uid == row.UsersIdusers {
			editUrl = fmt.Sprintf("/blogs/blog/%d/edit", row.Idblogs)
		}
		data.Rows = append(data.Rows, &BlogRow{
			GetBlogEntriesForUserDescendingLanguagesRow: row,
			EditUrl: editUrl,
		})
	}

	CustomBlogIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "page.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomBlogIndex(data *corecommon.CoreData, r *http.Request) {
	user := r.URL.Query().Get("user")
	if data.FeedsEnabled {
		if user == "" {
			data.CustomIndexItems = append(data.CustomIndexItems,
				IndexItem{
					Name: "Everyones Atom Feed",
					Link: "/blogs/atom",
				},
				IndexItem{
					Name: "Everyones RSS Feed",
					Link: "/blogs/rss",
				},
			)
		} else {
			data.CustomIndexItems = append(data.CustomIndexItems,
				IndexItem{
					Name: fmt.Sprintf("%s Atom Feed", user),
					Link: fmt.Sprintf("/blogs/atom?user=%s", url.QueryEscape(user)),
				},
				IndexItem{
					Name: fmt.Sprintf("%s RSS Feed", user),
					Link: fmt.Sprintf("/blogs/rss?user=%s", url.QueryEscape(user)),
				},
			)
		}
		data.RSSFeedUrl = "/blogs/rss"
		data.AtomFeedUrl = "/blogs/atom"
	}

	userHasAdmin := data.HasRole("administrator")
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/admin/blogs/user/permissions",
		})
	}
	userHasWriter := data.HasRole("writer")
	if userHasWriter {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Write blog",
			Link: "/blogs/add",
		})

	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "List bloggers",
		Link: "/blogs/bloggers",
	})
	if user == "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/blogs?offset=%d", offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/blogs?offset=%d", offset-15),
			})
		}
	} else {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/blogs?user=%s&offset=%d", url.QueryEscape(user), offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/blogs?user=%s&offset=%d", url.QueryEscape(user), offset-15),
			})
		}
	}
}

func RssPage(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("rss")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	if err := feed.WriteRss(w); err != nil {
		log.Printf("Feed write Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AtomPage(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("rss")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
	if err != nil {
		log.Printf("Username to uid error: %s", err)
	}
	feed, err := FeedGen(r, queries, int(u.Idusers), username)
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

func FeedGen(r *http.Request, queries *db.Queries, uid int, username string) (*feeds.Feed, error) {

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

	rows, err := queries.GetBlogEntriesForUserDescendingLanguages(r.Context(), GetBlogEntriesForUserDescendingLanguagesParams{
		UsersIdusers:  int32(uid),
		ViewerIdusers: int32(uid),
		Limit:         15,
		Offset:        0,
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
		conv := a4code2html.NewA4Code2HTML()
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(row.Blog.String)
		conv.Process()
		i := len(row.Blog.String)
		if i > 255 {
			i = 255
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title: row.Blog.String[:i],
			Link: &feeds.Link{
				Href: u.String(),
			},
			Description: fmt.Sprintf("%s\n-\n%s", conv.Output(), row.Username.String),
		})
	}
	return feed, nil
}
