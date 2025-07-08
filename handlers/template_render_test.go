package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/forum"
	db "github.com/arran4/goa4web/internal/db"
	"html/template"
)

func stubFuncs() template.FuncMap {
	req := httptest.NewRequest("GET", "/", nil)
	m := corecommon.NewFuncs(req)
	m["LatestNews"] = func() (any, error) { return nil, nil }
	return m
}

func TestPageTemplatesRender(t *testing.T) {
	tmpl := templates.GetCompiledTemplates(stubFuncs())

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
		{"newsPage", struct{ *corecommon.CoreData }{&corecommon.CoreData{}}},
		{"faqPage", struct {
			*corecommon.CoreData
			FAQ any
		}{&corecommon.CoreData{}, nil}},
		{"userPage", struct{ *corecommon.CoreData }{&corecommon.CoreData{}}},
		{"linkerPage", struct {
			*corecommon.CoreData
			Categories any
			Links      any
			HasOffset  bool
			CatId      int32
		}{&corecommon.CoreData{}, nil, nil, false, 0}},
		{"forumPage", struct {
			*corecommon.CoreData
			Categories          []*forum.ForumcategoryPlus
			CategoryBreadcrumbs []*forum.ForumcategoryPlus
			Category            *forum.ForumcategoryPlus
			Admin               bool
		}{&corecommon.CoreData{}, nil, nil, nil, false}},
		{"bookmarksPage", struct{ *corecommon.CoreData }{&corecommon.CoreData{}}},
		{"imagebbsPage", struct {
			*corecommon.CoreData
			Boards any
		}{&corecommon.CoreData{}, nil}},
		{"blogsPage", struct {
			*corecommon.CoreData
			Rows     any
			IsOffset bool
			UID      string
			Blogs    []struct{ Username string }
		}{&corecommon.CoreData{}, nil, false, "", []struct{ Username string }{{"test"}}}},
		{"writingsPage", struct {
			*corecommon.CoreData
			Categories                       []*db.Writingcategory
			WritingcategoryIdwritingcategory int32
			CategoryId                       int32
			IsAdmin                          bool
		}{&corecommon.CoreData{}, nil, 0, 0, false}},
		{"linkerCategoryPage", struct {
			*corecommon.CoreData
			Offset      int
			HasOffset   bool
			CatId       int
			CommentOnId int
			ReplyToId   int
			Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
		}{&corecommon.CoreData{}, 0, false, 0, 0, 0, nil}},
		{"writingsCategoryPage", struct {
			*corecommon.CoreData
			Categories                       []*db.Writingcategory
			CategoryBreadcrumbs              []*db.Writingcategory
			EditingCategoryId                int32
			CategoryId                       int32
			WritingcategoryIdwritingcategory int32
			IsAdmin                          bool
			IsWriter                         bool
			Abstracts                        []*db.GetPublicWritingsInCategoryRow
		}{&corecommon.CoreData{}, nil, nil, 0, 0, 0, false, false, nil}},
		{"searchPage", struct {
			*corecommon.CoreData
			SearchWords string
		}{&corecommon.CoreData{}, ""}},
		{"adminSearchPage", struct {
			*corecommon.CoreData
			Stats struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }
		}{&corecommon.CoreData{}, struct{ Words, Comments, News, Blogs, Linker, Writing, Writings, Images int64 }{}}},
		{"adminPage", struct {
			*corecommon.CoreData
			AdminLinks []corecommon.IndexItem
			Stats      adminStats
		}{&corecommon.CoreData{}, nil, adminStats{}}},
		{"forumAdminPage", struct {
			*corecommon.CoreData
			Stats struct{ Categories, Topics, Threads int64 }
		}{&corecommon.CoreData{}, struct{ Categories, Topics, Threads int64 }{}}},
		{"imagebbsAdminPage", struct{ *corecommon.CoreData }{&corecommon.CoreData{}}},
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
