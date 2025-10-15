package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
)

// AdminCategoryCreatePage renders a form to create a new forum category.
func AdminCategoryCreatePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := struct {
		Categories []*db.Forumcategory
	}{Categories: cats}
	cd.PageTitle = "Create Forum Category"
	handlers.TemplateHandler(w, r, "forumAdminCategoryCreatePage.gohtml", data)
}

// AdminCategoryCreateSubmit handles creation of a new forum category.
func AdminCategoryCreateSubmit(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	_ = name
	_ = desc
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idforumcategory] = c.ForumcategoryIdforumcategory
	}
	if path, loop := algorithms.WouldCreateLoop(parents, 0, int32(pcid)); loop {
		handlers.RedirectSeeOtherWithMessage(w, r, "", fmt.Sprintf("loop %v", path))
		return
	}

	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	_ = languageID // TODO: implement category creation
	http.Redirect(w, r, "/admin/forum/categories", http.StatusSeeOther)
}
