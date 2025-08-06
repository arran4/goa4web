package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories []*db.AdminListForumCategoriesWithSubcategoryAndTopicCountsRow
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

	categoryRows, err := queries.AdminListForumCategoriesWithSubcategoryAndTopicCounts(r.Context(), db.AdminListForumCategoriesWithSubcategoryAndTopicCountsParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("adminListForumCategoriesWithSubcategoryAndTopicCounts: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	totalCount, err := queries.AdminCountForumCategories(r.Context())
	if err != nil {
		log.Printf("adminCountForumCategories: %v", err)
	}

	numPages := int((totalCount + int64(pageSize-1)) / int64(pageSize))
	base := "/admin/forum/categories"
	for i := 1; i <= numPages; i++ {
		cd.PageLinks = append(cd.PageLinks, common.PageLink{Num: i, Link: fmt.Sprintf("%s?page=%d", base, i), Active: i == page})
	}
	if page < numPages {
		cd.NextLink = fmt.Sprintf("%s?page=%d", base, page+1)
	}
	if page > 1 {
		cd.PrevLink = fmt.Sprintf("%s?page=%d", base, page-1)
	}

	catsAll, err := queries.GetAllForumCategories(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("getAllForumCategories: %v", err)
	}
	data := Data{Categories: categoryRows}
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["category"])
	if err != nil {
		http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
		return
	}

	if r.Method == http.MethodGet {
		cat, err := queries.GetForumCategoryById(r.Context(), int32(categoryID))
		if err != nil {
			http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
			return
		}
		cats, err := queries.GetAllForumCategories(r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		topics, err := queries.GetForumTopicsByCategoryId(r.Context(), int32(categoryID))
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getForumTopicsByCategoryId: %v", err)
		}
		cd.PageTitle = fmt.Sprintf("Edit Forum Category %d", categoryID)
		data := struct {
			Category   *db.Forumcategory
			Categories []*db.Forumcategory
			TopicCount int
		}{Category: cat, Categories: cats, TopicCount: len(topics)}
		handlers.TemplateHandler(w, r, "forumAdminCategoryEditPage.gohtml", data)
		return
	}

	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	cats, err := queries.GetAllForumCategories(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idforumcategory] = c.ForumcategoryIdforumcategory
	}
	if path, loop := algorithms.WouldCreateLoop(parents, int32(categoryID), int32(pcid)); loop {
		http.Redirect(w, r, "?error="+fmt.Sprintf("loop %v", path), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.AdminUpdateForumCategory(r.Context(), db.AdminUpdateForumCategoryParams{
		Title:                        sql.NullString{Valid: true, String: name},
		Description:                  sql.NullString{Valid: true, String: desc},
		Idforumcategory:              int32(categoryID),
		ForumcategoryIdforumcategory: int32(pcid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/forum/categories/category/%d", categoryID), http.StatusTemporaryRedirect)
}

func AdminCategoryCreatePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if r.Method == http.MethodGet {
		cats, err := queries.GetAllForumCategories(r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		cd.PageTitle = "Create Forum Category"
		data := struct{ Categories []*db.Forumcategory }{Categories: cats}
		handlers.TemplateHandler(w, r, "forumAdminCategoryCreatePage.gohtml", data)
		return
	}

	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	cats, err := queries.GetAllForumCategories(r.Context())
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

	if err := queries.AdminCreateForumCategory(r.Context(), db.AdminCreateForumCategoryParams{
		ForumcategoryIdforumcategory: int32(pcid),
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
		http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
		return
	}
	if err := queries.AdminDeleteForumCategory(r.Context(), int32(cid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/admin/forum/categories", http.StatusTemporaryRedirect)
}
