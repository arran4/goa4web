package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"html/template"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/forum"
	db "github.com/arran4/goa4web/internal/db"
)

func stubFuncs() template.FuncMap {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &corecorecommon.CoreData{}
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
		{"newsPage", struct{ *corecorecommon.CoreData }{&corecorecommon.CoreData{}}},
		{"faqPage", struct {
			*corecorecommon.CoreData
			FAQ any
		}{&corecorecommon.CoreData{}, nil}},
		{"userPage", struct{ *corecorecommon.CoreData }{&corecorecommon.CoreData{}}},
		{"linkerPage", struct {
			*corecorecommon.CoreData
			Categories any
			Links      any
			HasOffset  bool
			CatId      int32
		}{&corecorecommon.CoreData{}, nil, nil, false, 0}},
		{"forumPage", struct {
			*corecorecommon.CoreData
			Categories          []*forum.ForumcategoryPlus
			CategoryBreadcrumbs []*forum.ForumcategoryPlus
			Category            *forum.ForumcategoryPlus
			Admin               bool
		}{&corecorecommon.CoreData{}, nil, nil, nil, false}},
		{"bookmarksPage", struct{ *corecorecommon.CoreData }{&corecorecommon.CoreData{}}},
		{"imagebbsPage", struct {
			*corecorecommon.CoreData
			Boards any
		}{&corecorecommon.CoreData{}, nil}},
		{"blogsPage", struct {
			*corecorecommon.CoreData
			Rows     any
			IsOffset bool
			UID      string
			Blogs    []struct{ Username string }
		}{&corecorecommon.CoreData{}, nil, false, "", []struct{ Username string }{{"test"}}}},
		{"writingsPage", struct {
			*corecorecommon.CoreData
			Categories        []*db.WritingCategory
			WritingCategoryID int32
			CategoryId        int32
			IsAdmin           bool
		}{&corecorecommon.CoreData{}, nil, 0, 0, false}},
		{"linkerCategoryPage", struct {
			*corecorecommon.CoreData
			Offset      int
			HasOffset   bool
			CatId       int
			CommentOnId int
			ReplyToId   int
			Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
		}{&corecorecommon.CoreData{}, 0, false, 0, 0, 0, nil}},
		{"writingsCategoryPage", struct {
			*corecorecommon.CoreData
			Categories          []*db.WritingCategory
			CategoryBreadcrumbs []*db.WritingCategory
			EditingCategoryId   int32
			CategoryId          int32
			WritingCategoryID   int32
			IsAdmin             bool
			IsWriter            bool
			Abstracts           []*db.GetPublicWritingsInCategoryRow
		}{&corecorecommon.CoreData{}, nil, nil, 0, 0, 0, false, false, nil}},
		{"searchPage", struct {
			*corecorecommon.CoreData
			SearchWords string
		}{&corecorecommon.CoreData{}, ""}},
		{"adminSearchPage", struct {
			*corecorecommon.CoreData
			Stats struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }
		}{&corecorecommon.CoreData{}, struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }{}}},
		{"adminPage", struct {
			*corecorecommon.CoreData
			AdminLinks []corecommon.IndexItem
			Stats      adminStats
		}{&corecorecommon.CoreData{}, nil, adminStats{}}},
		{"forumAdminPage", struct {
			*corecorecommon.CoreData
			Stats struct{ Categories, Topics, Threads int64 }
		}{&corecorecommon.CoreData{}, struct{ Categories, Topics, Threads int64 }{}}},
		{"imagebbsAdminPage", struct{ *corecorecommon.CoreData }{&corecorecommon.CoreData{}}},
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
