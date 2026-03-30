package handlers_test

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/handlers/forum"

	"html/template"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func stubFuncs() template.FuncMap {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{Config: config.NewRuntimeConfig()}
	return cd.Funcs(req)
}

func TestHappyPathPageTemplatesRender(t *testing.T) {
	tmpl := templates.GetCompiledSiteTemplates(stubFuncs(), templates.WithSilence(true))
	req := httptest.NewRequest("GET", "/", nil)

	type adminStats struct {
		Users        int64
		Languages    int64
		News         int64
		Blogs        int64
		ForumTopics  int64
		ForumThreads int64
		Writings     int64
	}

	pages := []struct {
		name string
		data any
	}{
		{"domains/news/page.gohtml", struct{}{}},
		{"domains/faq/page.gohtml", struct{ *common.CoreData }{&common.CoreData{}}},
		{"domains/user/page.gohtml", struct{ *common.CoreData }{&common.CoreData{}}},
		{"domains/linker/page.gohtml", struct {
			*common.CoreData
			Categories any
			Links      any
			HasOffset  bool
			CatId      int32
		}{&common.CoreData{}, nil, nil, false, 0}},
		{"domains/forum/page.gohtml", struct {
			*common.CoreData
			Categories []*forum.ForumcategoryPlus
			Category   *forum.ForumcategoryPlus
			Admin      bool
		}{&common.CoreData{}, nil, nil, false}},
		{"domains/bookmarks/page.gohtml", struct{ *common.CoreData }{&common.CoreData{}}},
		{"domains/imagebbs/page.gohtml", struct {
			*common.CoreData
			Boards any
		}{&common.CoreData{}, nil}},
		{"domains/domains/blogs/page.gohtml", struct{}{}},
		{"domains/writings/page.gohtml", struct {
			WritingCategoryID int32
		}{0}},
		{"domains/linker/categoryPage.gohtml", struct {
			*common.CoreData
			Offset      int
			HasOffset   bool
			CatId       int
			CommentOnId int
			ReplyToId   int
			Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginatedRow
		}{&common.CoreData{}, 0, false, 0, 0, 0, nil}},
		{"writingsCategoryPage", struct {
			Request           *http.Request
			CategoryId        int32
			WritingCategoryID int32
		}{req, 0, 0}},
		{"domains/domains/admin/searchPage.gohtml", struct {
			*common.CoreData
			SearchWords string
		}{&common.CoreData{}, ""}},
		{"domains/domains/admin/searchPage.gohtml", struct {
			*common.CoreData
			Stats struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }
		}{&common.CoreData{}, struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }{}}},
		{"domains/domains/blogs/adminPage.gohtml", struct {
			*common.CoreData
			Rows []struct {
				Username sql.NullString
				Email    string
				Roles    sql.NullString
				Idusers  int32
			}
		}{&common.CoreData{}, nil}},
		{"domains/admin/page.gohtml", struct {
			*common.CoreData
			AdminSections []common.AdminSection
			Stats         adminStats
		}{&common.CoreData{}, nil, adminStats{}}},
		{"domains/forum/adminPage.gohtml", struct {
			*common.CoreData
			Stats struct{ Categories, Topics, Threads int64 }
		}{&common.CoreData{}, struct{ Categories, Topics, Threads int64 }{}}},
		{"domains/imagebbs/adminPage.gohtml", struct {
			*common.CoreData
			Stats []*db.AdminImageboardPostCountsRow
		}{&common.CoreData{}, nil}},
	}

	for _, p := range pages {
		t.Run(p.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, p.name, p.data); err != nil {
				t.Fatalf("render %s: %v", p.name, err)
			}
		})
	}
}
