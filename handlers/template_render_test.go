package handlers_test

import (
	"bytes"
	"github.com/arran4/goa4web/handlers/forum"
	"net/http/httptest"
	"testing"

	"html/template"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func stubFuncs() template.FuncMap {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	m := cd.Funcs(req)
	m["LatestNews"] = func() (any, error) { return nil, nil }
	return m
}

func TestPageTemplatesRender(t *testing.T) {
	tmpl := templates.GetCompiledSiteTemplates(stubFuncs())

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
		{"newsPage", struct{ *common.CoreData }{&common.CoreData{}}},
		{"faqPage", struct {
			*common.CoreData
			FAQ any
		}{&common.CoreData{}, nil}},
		{"userPage", struct{ *common.CoreData }{&common.CoreData{}}},
		{"linkerPage", struct {
			*common.CoreData
			Categories any
			Links      any
			HasOffset  bool
			CatId      int32
		}{&common.CoreData{}, nil, nil, false, 0}},
		{"forumPage", struct {
			*common.CoreData
			Categories          []*forum.ForumcategoryPlus
			CategoryBreadcrumbs []*forum.ForumcategoryPlus
			Category            *forum.ForumcategoryPlus
			Admin               bool
		}{&common.CoreData{}, nil, nil, nil, false}},
		{"bookmarksPage", struct{ *common.CoreData }{&common.CoreData{}}},
		{"imagebbsPage", struct {
			*common.CoreData
			Boards any
		}{&common.CoreData{}, nil}},
		{"blogsPage", struct {
			*common.CoreData
			Rows     any
			IsOffset bool
			UID      string
			Blogs    []struct{ Username string }
		}{&common.CoreData{}, nil, false, "", []struct{ Username string }{{"test"}}}},
		{"writingsPage", struct {
			*common.CoreData
			Categories        []*db.WritingCategory
			WritingCategoryID int32
			CategoryId        int32
		}{&common.CoreData{}, nil, 0, 0}},
		{"linkerCategoryPage", struct {
			*common.CoreData
			Offset      int
			HasOffset   bool
			CatId       int
			CommentOnId int
			ReplyToId   int
			Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginatedRow
		}{&common.CoreData{}, 0, false, 0, 0, 0, nil}},
		{"writingsCategoryPage", struct {
			*common.CoreData
			Categories          []*db.WritingCategory
			CategoryBreadcrumbs []*db.WritingCategory
			CategoryId          int32
			WritingCategoryID   int32
			Abstracts           []*db.GetPublicWritingsInCategoryRow
		}{&common.CoreData{}, nil, nil, 0, 0, nil}},
		{"searchPage", struct {
			*common.CoreData
			SearchWords string
		}{&common.CoreData{}, ""}},
		{"adminSearchPage", struct {
			*common.CoreData
			Stats struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }
		}{&common.CoreData{}, struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }{}}},
		{"adminPage", struct {
			*common.CoreData
			AdminSections []common.AdminSection
			Stats         adminStats
		}{&common.CoreData{}, nil, adminStats{}}},
		{"forumAdminPage", struct {
			*common.CoreData
			Stats struct{ Categories, Topics, Threads int64 }
		}{&common.CoreData{}, struct{ Categories, Topics, Threads int64 }{}}},
		{"imagebbsAdminPage", struct {
			*common.CoreData
			Stats []*db.ImageboardPostCountsRow
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
