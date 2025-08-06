package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminBoardImagesPage lists all images for a board.
func AdminBoardImagesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Board *db.Imageboard
		Posts []*db.ListImagePostsByBoardForListerRow
	}

	vars := mux.Vars(r)
	bidStr := vars["board"]
	bid, _ := strconv.Atoi(bidStr)
	if bid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	const pageSize = 20
	offset := (page - 1) * pageSize

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Board Images"
	queries := cd.Queries()

	board, err := queries.GetImageBoardById(r.Context(), int32(bid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		log.Printf("get image board: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	posts, err := queries.ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
		ListerID:     cd.UserID,
		BoardID:      int32(bid),
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        int32(pageSize),
		Offset:       int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list posts: %v", err)
	}

	total := int64(0)
	counts, err := queries.AdminImageboardPostCounts(r.Context())
	if err == nil {
		for _, c := range counts {
			if c.Idimageboard == int32(bid) {
				total = c.Count
				break
			}
		}
	}

	numPages := int((total + int64(pageSize-1)) / int64(pageSize))
	base := fmt.Sprintf("/admin/imagebbs/board/%d/images", bid)
	vals := url.Values{}
	for i := 1; i <= numPages; i++ {
		vals.Set("page", strconv.Itoa(i))
		cd.PageLinks = append(cd.PageLinks, common.PageLink{Num: i, Link: base + "?" + vals.Encode(), Active: i == page})
	}
	if page < numPages {
		vals.Set("page", strconv.Itoa(page+1))
		cd.NextLink = base + "?" + vals.Encode()
	}
	if page > 1 {
		vals.Set("page", strconv.Itoa(page-1))
		cd.PrevLink = base + "?" + vals.Encode()
	}

	data := Data{Board: board, Posts: posts}
	handlers.TemplateHandler(w, r, "adminBoardImagesPage.gohtml", data)
}
