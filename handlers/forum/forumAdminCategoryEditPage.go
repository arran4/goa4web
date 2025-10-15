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
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := cd.ForumCategory(int32(cid))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
		return
	}
	cats, err := cd.ForumCategories()
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
	_ = name
	_ = desc
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	_ = queries
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	cats, err := cd.ForumCategories()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idforumcategory] = c.ForumcategoryIdforumcategory
	}
	if path, loop := algorithms.WouldCreateLoop(parents, int32(categoryId), int32(pcid)); loop {
		handlers.RedirectSeeOtherWithMessage(w, r, "", fmt.Sprintf("loop %v", path))
		return
	}

	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	_ = languageID // TODO: implement category update

	redirectURL := "/admin/forum/categories"
	if strings.HasSuffix(r.URL.Path, "/edit") {
		redirectURL = fmt.Sprintf("/admin/forum/categories/category/%d", categoryId)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// AdminCategoryDeletePage removes a forum category.
func AdminCategoryDeletePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if err := cd.Queries().AdminDeleteForumCategory(r.Context(), int32(cid)); err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	http.Redirect(w, r, "/admin/forum/categories", http.StatusSeeOther)
}
