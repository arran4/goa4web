package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminCategoryEditPage displays a form to edit a forum category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := queries.GetForumCategoryById(r.Context(), db.GetForumCategoryByIdParams{
		Idforumcategory: int32(cid),
		ViewerID:        cd.UserID,
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
		return
	}
	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{
		ViewerID: cd.UserID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Forum Category %d", cid)
	data := struct {
		Category   *db.Forumcategory
		Categories []*db.Forumcategory
	}{
		Category:   cat,
		Categories: cats,
	}
	handlers.TemplateHandler(w, r, "forumAdminCategoryEditPage.gohtml", data)
}

// AdminCategoryEditSubmit processes updates to an existing forum category.
func AdminCategoryEditSubmit(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idforumcategory] = c.ForumcategoryIdforumcategory
	}
	if path, loop := algorithms.WouldCreateLoop(parents, int32(categoryId), int32(pcid)); loop {
		http.Redirect(w, r, "?error="+fmt.Sprintf("loop %v", path), http.StatusTemporaryRedirect)
		return
	}

	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	if err := queries.AdminUpdateForumCategory(r.Context(), db.AdminUpdateForumCategoryParams{
		Title:                        sql.NullString{Valid: true, String: name},
		Description:                  sql.NullString{Valid: true, String: desc},
		Idforumcategory:              int32(categoryId),
		ForumcategoryIdforumcategory: int32(pcid),
		LanguageIdlanguage:           int32(languageID),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	redirectURL := "/admin/forum/categories"
	if strings.HasSuffix(r.URL.Path, "/edit") {
		redirectURL = fmt.Sprintf("/admin/forum/categories/category/%d", categoryId)
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// AdminCategoryDeletePage removes a forum category.
func AdminCategoryDeletePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := queries.AdminDeleteForumCategory(r.Context(), int32(cid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
}
