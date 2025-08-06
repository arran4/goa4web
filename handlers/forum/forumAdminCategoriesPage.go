package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
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
		Categories []*db.ListForumCategoriesWithCountsPaginatedForViewerRow
		Tree       template.HTML
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin Categories"
	queries := cd.Queries()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := cd.PageSize()
	offset := (page - 1) * pageSize

	data := Data{}

	categoryRows, err := queries.ListForumCategoriesWithCountsPaginatedForViewer(r.Context(), db.ListForumCategoriesWithCountsPaginatedForViewerParams{
		ViewerID: cd.UserID,
		Limit:    int32(pageSize),
		Offset:   int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list categories: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Categories = categoryRows

	totalCount, err := queries.CountForumCategoriesForViewer(r.Context(), db.CountForumCategoriesForViewerParams{ViewerID: cd.UserID})
	if err != nil {
		log.Printf("count categories: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	numPages := int((totalCount + int64(pageSize-1)) / int64(pageSize))
	base := "/admin/forum/categories"
	vals := url.Values{}
	for i := 1; i <= numPages; i++ {
		vals.Set("page", strconv.Itoa(i))
		cd.PageLinks = append(cd.PageLinks, common.PageLink{Num: i, Link: base + "?" + vals.Encode(), Active: i == page})
	}
	if page < numPages {
		vals.Set("page", strconv.Itoa(page+1))
		cd.NextLink = base + "?" + vals.Encode()
	}
	if page > 1 {
		vals.Set("page", strconv.Itoa(page-1))
		cd.PrevLink = base + "?" + vals.Encode()
	}

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

	redirectURL := "/admin/forum/categories"
	if strings.HasSuffix(r.URL.Path, "/edit") {
		redirectURL = fmt.Sprintf("/admin/forum/categories/category/%d", categoryId)
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

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

func AdminCategoryDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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
