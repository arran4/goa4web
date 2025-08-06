package writings

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories []*db.WritingCategory
		Tree       template.HTML
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
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows
	children := map[int32][]*db.WritingCategory{}
	for _, c := range categoryRows {
		children[c.WritingCategoryID] = append(children[c.WritingCategoryID], c)
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

	handlers.TemplateHandler(w, r, "writingsAdminCategoriesPage.gohtml", data)
}
