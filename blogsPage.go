package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func blogsPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*show_latest_blogsRow
		EditUrl string
	}
	type Data struct {
		*CoreData
		Rows     []*BlogRow
		IsOffset bool
		UID      string
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	buid := r.URL.Query().Get("uid")
	userId, _ := strconv.Atoi(buid)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	userLanguagePref := 0

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.show_latest_blogs(r.Context(), show_latest_blogsParams{
		UsersIdusers:       int32(userId),
		LanguageIdlanguage: int32(userLanguagePref),
		Limit:              15,
		Offset:             int32(offset),
	})
	if err != nil {
		log.Printf("Query Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsOffset: offset != 0,
		UID:      buid,
	}

	for _, row := range rows {
		editUrl := ""
		if uid == row.UsersIdusers {
			editUrl = fmt.Sprintf("/blogs/blog/%d/edit", row.Idblogs)
		}
		data.Rows = append(data.Rows, &BlogRow{
			show_latest_blogsRow: row,
			EditUrl:              editUrl,
		})
	}

	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "blogsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomBlogIndex(data *CoreData, r *http.Request) {
	user := r.URL.Query().Get("user")
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

	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/blogs/user/permissions",
		})
	}
	userHasWriter := true // TODO
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

func blogsRssPage(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("rss")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid, err := queries.usernametouid(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
	if err != nil {
		log.Printf("Username to uid error: %s", err)
	}
	feed, err := FeedGen(r, queries, int(uid), username)
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

func blogsAtomPage(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("rss")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid, err := queries.usernametouid(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
	if err != nil {
		log.Printf("Username to uid error: %s", err)
	}
	feed, err := FeedGen(r, queries, int(uid), username)
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

func FeedGen(r *http.Request, queries *Queries, uid int, username string) (*feeds.Feed, error) {

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

	rows, err := queries.write_blog_rss(r.Context(), write_blog_rssParams{
		UsersIdusers: int32(uid),
		Limit:        15,
	})
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		u := r.URL
		u.Query().Set("show", fmt.Sprintf("%d", row.Idblogs))
		var text = &A4code2html{}
		text.codeType = ct_tagstrip
		text.input = row.Blog.String
		text.Process()
		feed.Items = append(feed.Items, &feeds.Item{
			Title: row.Left,
			Link: &feeds.Link{
				Href: u.String(),
			},
			Description: fmt.Sprintf("%s\n-\n%s", text.output.String(), row.Username.String),
		})
	}
	return feed, nil
}
