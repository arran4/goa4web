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
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idforumcategory] = c.ForumcategoryIdforumcategory
	}
	if path, loop := algorithms.WouldCreateLoop(parents, 0, int32(pcid)); loop {
		http.Redirect(w, r, "?error="+fmt.Sprintf("loop %v", path), http.StatusTemporaryRedirect)
		return
	}

	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	if err := queries.AdminCreateForumCategory(r.Context(), db.AdminCreateForumCategoryParams{
		ForumcategoryIdforumcategory: int32(pcid),
		LanguageIdlanguage:           int32(languageID),
		Title:                        sql.NullString{Valid: true, String: name},
		Description:                  sql.NullString{Valid: true, String: desc},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
}
