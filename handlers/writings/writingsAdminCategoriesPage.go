package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories    []*db.WritingCategory
		AllCategories []*db.WritingCategory
		Tree          template.HTML
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writing Categories"
	data := Data{}

	categoryRows, err := cd.WritingCategories()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("writingCategories Error: %s", err)
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	pageSize := cd.PageSize()
	end := offset + pageSize
	if end > len(categoryRows) {
		end = len(categoryRows)
	}
	hasMore := len(categoryRows) > end
	base := "/admin/writings/categories"
	if hasMore {
		cd.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+pageSize)
	}
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-pageSize)
	}
	data.AllCategories = categoryRows
	data.Categories = categoryRows[offset:end]
	children := map[int32][]*db.WritingCategory{}
	for _, c := range categoryRows {
		var pid int32
		if c.WritingCategoryID.Valid {
			pid = c.WritingCategoryID.Int32
		}
		children[pid] = append(children[pid], c)
	}
	var build func(parent int32) string
	build = func(parent int32) string {
		var sb strings.Builder
		if cs, ok := children[parent]; ok {
			sb.WriteString("<ul>")
			for _, c := range cs {
				sb.WriteString("<li>")
				sb.WriteString(template.HTMLEscapeString(c.Title.String))
				sb.WriteString(build(c.Idwritingcategory))
				sb.WriteString("</li>")
			}
			sb.WriteString("</ul>")
		}
		return sb.String()
	}
	data.Tree = template.HTML(build(0))

	WritingsAdminCategoriesPageTmpl.Handle(w, r, data)
}

const WritingsAdminCategoriesPageTmpl tasks.Template = "writings/writingsAdminCategoriesPage.gohtml"
