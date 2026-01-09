package share

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sharesign"
)

var (
	newsRegex     = regexp.MustCompile(`/(news)/news/(\d+)`)
	blogsRegex    = regexp.MustCompile(`/(blogs)/blog/(\d+)`)
	writingsRegex = regexp.MustCompile(`/(writings)/article/(\d+)`)
	forumRegex    = regexp.MustCompile(`/(forum|private)/thread/(\d+)`)
)

type SharedContentHandler struct {
	signer *sharesign.Signer
	server *server.Server
}

func NewSharedContentHandler(signer *sharesign.Signer, server *server.Server) *SharedContentHandler {
	return &SharedContentHandler{signer: signer, server: server}
}

func (h *SharedContentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := strings.TrimPrefix(r.URL.Path, "/shared")
	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")

	if !h.signer.Verify(u, ts, sig) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	cd, _ := h.server.GetCoreData(w, r)
	if cd == nil {
		return
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	var contentType string
	var contentID int

	if matches := newsRegex.FindStringSubmatch(parsedURL.Path); len(matches) == 3 {
		contentType = matches[1]
		contentID, _ = strconv.Atoi(matches[2])
	} else if matches := blogsRegex.FindStringSubmatch(parsedURL.Path); len(matches) == 3 {
		contentType = matches[1]
		contentID, _ = strconv.Atoi(matches[2])
	} else if matches := writingsRegex.FindStringSubmatch(parsedURL.Path); len(matches) == 3 {
		contentType = matches[1]
		contentID, _ = strconv.Atoi(matches[2])
	} else if matches := forumRegex.FindStringSubmatch(parsedURL.Path); len(matches) == 3 {
		contentType = matches[1]
		contentID, _ = strconv.Atoi(matches[2])
	} else {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	if contentID == 0 {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	var og_title, og_description string

	switch contentType {
	case "news":
		news, err := cd.Queries().GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
			ViewerID: cd.UserID,
			ID:       int32(contentID),
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		og_title = strings.Split(news.News.String, "\n")[0]
		og_description = a4code.Snip(news.News.String, 128)
	case "blogs":
		blog, err := cd.BlogEntryByID(int32(contentID))
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		og_title = blog.Title
		if og_title == "" {
			og_title = fmt.Sprintf("Blog by %s", blog.Username.String)
		}
		og_description = a4code.Snip(blog.Blog.String, 128)
	case "writings":
		writing, err := cd.WritingByID(int32(contentID))
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		og_title = writing.Title.String
		og_description = a4code.Snip(writing.Abstract.String, 128)
	case "forum", "private":
		thread, err := cd.ForumThreadByID(int32(contentID))
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		topic, err := cd.ForumTopicByID(thread.ForumtopicIdforumtopic)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		og_title = topic.Title.String
		comments, err := cd.ThreadComments(int32(contentID))
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if len(comments) > 0 {
			og_description = a4code.Snip(comments[0].Text.String, 128)
		}
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.New("og").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="{{.Title}}" />
	<meta property="og:description" content="{{.Description}}" />
	<meta property="og:image" content="{{.Image}}" />
	<meta property="og:url" content="{{.URL}}" />
</head>
<body>
	<h1>Content Not Available</h1>
	<p>Please log in to view this content.</p>
</body>
</html>
`)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title       string
		Description string
		Image       string
		URL         string
	}{
		Title:       og_title,
		Description: og_description,
		Image:       cd.AbsoluteURL(fmt.Sprintf("/api/og-image?title=%s", url.QueryEscape(og_title))),
		URL:         cd.AbsoluteURL(u),
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}
