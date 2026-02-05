package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/tasks"

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
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
	ForumAdminCategoryEditPageTmpl.Handle(w, r, data)
}

const ForumAdminCategoryEditPageTmpl tasks.Template = "forum/forumAdminCategoryEditPage.gohtml"

// AdminCategoryEditSubmit processes updates to an existing forum category.
func AdminCategoryEditSubmit(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PostFormValue("name"))
	desc := strings.TrimSpace(r.PostFormValue("desc"))
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	vars := mux.Vars(r)
	categoryId, err := strconv.Atoi(vars["category"])
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	langValue := strings.TrimSpace(r.PostFormValue("language"))
	languageID := 0
	if langValue != "" {
		languageID, err = strconv.Atoi(langValue)
		if err != nil {
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}
	if name == "" {
		handlers.RedirectSeeOtherWithMessage(w, r, "", "category name cannot be empty")
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cat, err := queries.GetForumCategoryById(r.Context(), db.GetForumCategoryByIdParams{
		Idforumcategory: int32(categoryId),
		ViewerID:        cd.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
			return
		}
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	if langValue == "" && cat.LanguageID.Valid {
		languageID = int(cat.LanguageID.Int32)
	}

	cats, err := cd.ForumCategories()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	parents := make(map[int32]int32, len(cats)+1)
	parents[int32(categoryId)] = cat.ForumcategoryIdforumcategory
	parentExists := pcid == 0 || pcid == int(cat.ForumcategoryIdforumcategory)
	for _, c := range cats {
		parents[c.Idforumcategory] = c.ForumcategoryIdforumcategory
		if int(c.Idforumcategory) == pcid {
			parentExists = true
		}
	}
	if !parentExists {
		handlers.RedirectSeeOtherWithMessage(w, r, "", fmt.Sprintf("parent category %d not found", pcid))
		return
	}
	if path, loop := algorithms.WouldCreateLoop(parents, int32(categoryId), int32(pcid)); loop {
		handlers.RedirectSeeOtherWithMessage(w, r, "", fmt.Sprintf("loop %v", path))
		return
	}

	if err := queries.AdminUpdateForumCategory(r.Context(), db.AdminUpdateForumCategoryParams{
		Title:           sql.NullString{String: name, Valid: true},
		Description:     sql.NullString{String: desc, Valid: true},
		ParentID:        int32(pcid),
		LanguageID:      sql.NullInt32{Int32: int32(languageID), Valid: languageID != 0},
		Idforumcategory: int32(categoryId),
	}); err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Name"] = name
		if u := cd.UserByID(cd.UserID); u != nil {
			evt.Data["Username"] = u.Username.String
		}
	}

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
	cat, err := cd.ForumCategory(int32(cid))
	if err == nil && cat != nil {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Name"] = cat.Title.String
			if u := cd.UserByID(cd.UserID); u != nil {
				evt.Data["Username"] = u.Username.String
			}
		}
	}

	if err := cd.Queries().AdminDeleteForumCategory(r.Context(), int32(cid)); err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	http.Redirect(w, r, "/admin/forum/categories", http.StatusSeeOther)
}
