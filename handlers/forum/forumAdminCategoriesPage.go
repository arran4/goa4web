package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/gorilla/mux"
	"html/template"
	"strings"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories []*db.GetAllForumCategoriesWithSubcategoryCountRow
		Tree       template.HTML
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Categories"
	queries := cd.Queries()

	data := Data{}

	categoryRows, err := queries.GetAllForumCategoriesWithSubcategoryCount(r.Context(), db.GetAllForumCategoriesWithSubcategoryCountParams{ViewerID: cd.UserID})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows
	catsAll, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
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

	handlers.TemplateHandler(w, r, "forumAdminCategoriesPage.gohtml", data)
}

func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
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
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idforumcategory:              int32(categoryId),
		ForumcategoryIdforumcategory: int32(pcid),
		LanguageIdlanguage:           int32(languageID),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
}

func AdminCategoryCreatePage(w http.ResponseWriter, r *http.Request) {
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

func AdminCategoryDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
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
