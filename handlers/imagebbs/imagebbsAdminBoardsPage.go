package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminBoardsPage(w http.ResponseWriter, r *http.Request) {
	type BoardRow struct {
		*db.Imageboard
		Threads int32
		Visible bool
		Nsfw    bool
	}
	type Data struct {
		*common.CoreData
		Boards []*BoardRow
		Tree   template.HTML
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Boards"
	data := Data{
		CoreData: cd,
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	boardRows, err := data.CoreData.ImageBoards()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllImageBoards Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	children := map[int32][]*BoardRow{}
	for _, b := range boardRows {
		threads, err := queries.AdminCountThreadsByBoard(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("countThreads error: %s", err)
			threads = 0
		}
		row := &BoardRow{
			Imageboard: b,
			Threads:    int32(threads),
			Visible:    true,
			Nsfw:       false,
		}
		data.Boards = append(data.Boards, row)
		children[b.ImageboardIdimageboard] = append(children[b.ImageboardIdimageboard], row)
	}

	var build func(parent int32) string
	build = func(parent int32) string {
		var sb strings.Builder
		if cs, ok := children[parent]; ok {
			sb.WriteString("<ul>")
			for _, c := range cs {
				sb.WriteString("<li>")
				sb.WriteString(template.HTMLEscapeString(c.Title.String))
				sb.WriteString(build(c.Idimageboard))
				sb.WriteString("</li>")
			}
			sb.WriteString("</ul>")
		}
		return sb.String()
	}
	data.Tree = template.HTML(build(0))

	handlers.TemplateHandler(w, r, "adminBoardsPage.gohtml", data)
}
