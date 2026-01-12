package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/gorilla/feeds"
)

func Page(w http.ResponseWriter, r *http.Request) {
	buid := r.URL.Query().Get("uid")
	userID, _ := strconv.Atoi(buid)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Blogs"
	cd.SetCurrentProfileUserID(int32(userID))

	offset := cd.Offset()
	ps := cd.PageSize()
	qv := r.URL.Query()
	qv.Set("offset", strconv.Itoa(offset+ps))
	cd.NextLink = "/blogs?" + qv.Encode()
	if offset > 0 {
		qv.Set("offset", strconv.Itoa(offset-ps))
		cd.PrevLink = "/blogs?" + qv.Encode()
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       "Blogs",
		Description: "Read blogs from our community.",
		Image:       share.MakeImageURL(cd.AbsoluteURL(), "Blogs", cd.ShareSigner, false),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "website",
	}

	BlogsPageTmpl.Handle(w, r, struct{}{})
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

func RssPageSigned(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	u, err := handlers.VerifyFeedRequest(r, "/blogs/rss")
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	// Pretend to be the signed user
	cd.UserID = u.Idusers

	RssPage(w, r)
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

func AtomPageSigned(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	u, err := handlers.VerifyFeedRequest(r, "/blogs/atom")
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	// Pretend to be the signed user
	cd.UserID = u.Idusers

	AtomPage(w, r)
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
		ListerID: cd.UserID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
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
		u := *r.URL
		q := u.Query()
		q.Set("show", fmt.Sprintf("%d", row.Idblogs))
		u.RawQuery = q.Encode()
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
