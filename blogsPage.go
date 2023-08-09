package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/feeds"
	"log"
	"net/http"
	"time"
)

func blogsHandler(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		CustomIndexItems []IndexItem
		RSSFeedUrl       string
		AtomFeedUrl      string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		CustomIndexItems: []IndexItem{
			{
				Name: "Atom Feed",
				Link: "/blogs/atom",
			},
			{
				Name: "RSS Feed",
				Link: "/blogs/rss",
			},
		},
		RSSFeedUrl:  "/blogs/rss",
		AtomFeedUrl: "/blogs/atom",
	}

	if err := compiledTemplates.ExecuteTemplate(w, "blogsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsRssHandler(w http.ResponseWriter, r *http.Request) {
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

func blogsAtomHandler(w http.ResponseWriter, r *http.Request) {
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
		var text = &a4code2html{}
		text.codeType = ct_tagstrip
		text.input = row.Blog.String
		text.process()
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
