package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminBoardsPage shows a paginated list of image boards.
func AdminBoardsPage(w http.ResponseWriter, r *http.Request) {
	type BoardRow struct {
		*db.Imageboard
		Threads int32
	}
	type Data struct {
		Boards []*BoardRow
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	const pageSize = 20
	start := (page - 1) * pageSize
	end := start + pageSize

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Boards"
	queries := cd.Queries()

	boards, err := cd.ImageBoards()
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getAllImageBoards Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	total := len(boards)
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	var data Data
	for _, b := range boards[start:end] {
		threads, err := queries.AdminCountThreadsByBoard(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("countThreads error: %s", err)
			threads = 0
		}
		data.Boards = append(data.Boards, &BoardRow{Imageboard: b, Threads: int32(threads)})
	}

	numPages := (total + pageSize - 1) / pageSize
	base := "/admin/imagebbs/boards"
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

	handlers.TemplateHandler(w, r, "adminBoardsPage.gohtml", data)
}
