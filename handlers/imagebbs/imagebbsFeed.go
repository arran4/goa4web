package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func RssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if _, ok := mux.Vars(r)["username"]; ok {
		u, err := handlers.VerifyFeedRequest(r, "/imagebbs/rss")
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		cd.UserID = u.Idusers
	}

	queries := cd.Queries()
	boards, err := queries.ListBoardsForLister(r.Context(), db.ListBoardsForListerParams{
		ListerID:     cd.UserID,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        200,
		Offset:       0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query boards error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	var posts []*db.ListImagePostsByBoardForListerRow
	for _, b := range boards {
		rows, err := queries.ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
			ListerID:     cd.UserID,
			BoardID:      sql.NullInt32{Int32: b.Idimageboard, Valid: true},
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:        200,
			Offset:       0,
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("feed query error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
		posts = append(posts, rows...)
	}
	feed := cd.ImageBBSFeed(r, "ImageBBS", 0, posts)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func AtomPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if _, ok := mux.Vars(r)["username"]; ok {
		u, err := handlers.VerifyFeedRequest(r, "/imagebbs/atom")
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		cd.UserID = u.Idusers
	}

	queries := cd.Queries()
	boards, err := queries.ListBoardsForLister(r.Context(), db.ListBoardsForListerParams{
		ListerID:     cd.UserID,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        200,
		Offset:       0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query boards error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	var posts []*db.ListImagePostsByBoardForListerRow
	for _, b := range boards {
		rows, err := queries.ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
			ListerID:     cd.UserID,
			BoardID:      sql.NullInt32{Int32: b.Idimageboard, Valid: true},
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:        200,
			Offset:       0,
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("feed query error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
		posts = append(posts, rows...)
	}
	feed := cd.ImageBBSFeed(r, "ImageBBS", 0, posts)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func BoardRssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)
	queries := cd.Queries()
	if !cd.HasGrant("imagebbs", "board", "see", int32(bid)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	rows, err := queries.ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
		ListerID:     cd.UserID,
		BoardID:      sql.NullInt32{Int32: int32(bid), Valid: true},
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        200,
		Offset:       0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	title := fmt.Sprintf("Board %d", bid)
	boards, err := queries.ListBoardsForLister(r.Context(), db.ListBoardsForListerParams{
		ListerID:     cd.UserID,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        200,
		Offset:       0,
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
	feed := cd.ImageBBSFeed(r, title, bid, rows)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func BoardAtomPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)
	queries := cd.Queries()
	if !cd.HasGrant("imagebbs", "board", "see", int32(bid)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	rows, err := queries.ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
		ListerID:     cd.UserID,
		BoardID:      sql.NullInt32{Int32: int32(bid), Valid: true},
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        200,
		Offset:       0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("feed query error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	title := fmt.Sprintf("Board %d", bid)
	boards, err := queries.ListBoardsForLister(r.Context(), db.ListBoardsForListerParams{
		ListerID:     cd.UserID,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        200,
		Offset:       0,
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
	feed := cd.ImageBBSFeed(r, title, bid, rows)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}
