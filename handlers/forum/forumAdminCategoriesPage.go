package forum

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/internal/tasks"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminCategoriesPage displays forum categories with pagination.
func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories []*db.AdminListForumCategoriesWithCountsRow
		Tree       template.HTML
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Categories"
	queries := cd.Queries()

	offset := cd.Offset()
	ps := cd.PageSize()

	total, err := queries.AdminCountForumCategories(r.Context(), db.AdminCountForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	rows, err := queries.AdminListForumCategoriesWithCounts(r.Context(), db.AdminListForumCategoriesWithCountsParams{ViewerID: cd.UserID, Limit: int32(ps), Offset: int32(offset)})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("AdminListForumCategoriesWithCounts: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	base := "/admin/forum/categories"
	cd.SetPagination(&common.OffsetPagination{
		TotalItems: int(total),
		PageSize:   ps,
		Offset:     offset,
		BaseURL:    base,
	})

	data := Data{Categories: rows}

	catsAll, err := cd.ForumCategories()
	if err == nil {
		children := map[int32][]*db.Forumcategory{}
		for _, c := range catsAll {
			children[c.ForumcategoryIdforumcategory] = append(children[c.ForumcategoryIdforumcategory], c)
		}
		var build func(parent int32) string
		build = func(parent int32) string {
			var sb strings.Builder
			if cs, ok := children[parent]; ok {
				sb.WriteString("<ul>")
				for _, c := range cs {
					sb.WriteString("<li>")
					sb.WriteString(template.HTMLEscapeString(c.Title.String))
					sb.WriteString(build(c.Idforumcategory))
					sb.WriteString("</li>")
				}
				sb.WriteString("</ul>")
			}
			return sb.String()
		}
		data.Tree = template.HTML(build(0))
	}

	ForumAdminCategoriesPageTmpl.Handle(w, r, data)
}

const ForumAdminCategoriesPageTmpl tasks.Template = "forum/forumAdminCategoriesPage.gohtml"
