package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

func imagebbsFeed(r *http.Request, title string, boardID int, rows []*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow) *feeds.Feed {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: fmt.Sprintf("Latest posts for %s", title),
		Created:     time.Now(),
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Posted.Time.After(rows[j].Posted.Time)
	})
	if len(rows) > 15 {
		rows = rows[:15]
	}
	for _, row := range rows {
		if !row.Description.Valid {
			continue
		}
		desc := row.Description.String
		conv := a4code2html.New(cd.ImageSigner.MapURL)
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(desc)
		out, _ := io.ReadAll(conv.Process())
		i := len(desc)
		if i > 255 {
			i = 255
		}
		item := &feeds.Item{
			Title:   desc[:i],
			Link:    &feeds.Link{Href: fmt.Sprintf("/imagebbs/board/%d/thread/%d", boardID, row.ForumthreadID)},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", string(out), func() string {
				if row.Username.Valid {
					return row.Username.String
				}
				return ""
			}()),
		}
		if row.Posted.Valid {
			item.Created = row.Posted.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}

func RssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	boards, err := queries.GetAllImageBoardsForUser(r.Context(), db.GetAllImageBoardsForUserParams{
		ViewerID:     cd.UserID,
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query boards error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var posts []*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow
	for _, b := range boards {
		rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUser(r.Context(), db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserParams{
			ViewerID:     cd.UserID,
			BoardID:      b.Idimageboard,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("feed query error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, rows...)
	}
	feed := imagebbsFeed(r, "ImageBBS", 0, posts)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AtomPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	boards, err := queries.GetAllImageBoardsForUser(r.Context(), db.GetAllImageBoardsForUserParams{
		ViewerID:     cd.UserID,
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query boards error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var posts []*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow
	for _, b := range boards {
		rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUser(r.Context(), db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserParams{
			ViewerID:     cd.UserID,
			BoardID:      b.Idimageboard,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("feed query error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, rows...)
	}
	feed := imagebbsFeed(r, "ImageBBS", 0, posts)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func BoardRssPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if !cd.HasGrant("imagebbs", "board", "see", int32(bid)) {
		_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd)
		return
	}
	rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUser(r.Context(), db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserParams{
		ViewerID:     cd.UserID,
		BoardID:      int32(bid),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	title := fmt.Sprintf("Board %d", bid)
	boards, err := queries.GetAllImageBoardsForUser(r.Context(), db.GetAllImageBoardsForUserParams{
		ViewerID:     cd.UserID,
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err == nil {
		for _, b := range boards {
			if int(b.Idimageboard) == bid {
				if b.Title.Valid {
					title = b.Title.String
				}
				break
			}
		}
	}
	feed := imagebbsFeed(r, title, bid, rows)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func BoardAtomPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if !cd.HasGrant("imagebbs", "board", "see", int32(bid)) {
		_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd)
		return
	}
	rows, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUser(r.Context(), db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserParams{
		ViewerID:     cd.UserID,
		BoardID:      int32(bid),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	title := fmt.Sprintf("Board %d", bid)
	boards, err := queries.GetAllImageBoardsForUser(r.Context(), db.GetAllImageBoardsForUserParams{
		ViewerID:     cd.UserID,
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err == nil {
		for _, b := range boards {
			if int(b.Idimageboard) == bid {
				if b.Title.Valid {
					title = b.Title.String
				}
				break
			}
		}
	}
	feed := imagebbsFeed(r, title, bid, rows)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
