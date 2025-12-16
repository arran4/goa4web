package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

type SearchPageData struct {
	SearchWords       string
	CanSearchForum    bool
	CanSearchNews     bool
	CanSearchLinker   bool
	CanSearchBlogs    bool
	CanSearchWritings bool
	AnySearch         bool
}

func GetSearchPageData(cd *common.CoreData) SearchPageData {
	hasGeneralSearch := cd.HasGrant("search", "", "search", 0)

	canSeeForum := cd.HasGrant("forum", "", "see", 0) || cd.HasGrant("forum", "", "view", 0) || cd.HasGrant("forum", "", "search", 0)
	canSeeNews := cd.HasGrant("news", "", "see", 0) || cd.HasGrant("news", "", "view", 0) || cd.HasGrant("news", "", "search", 0)
	canSeeLinker := cd.HasGrant("linker", "", "see", 0) || cd.HasGrant("linker", "", "view", 0) || cd.HasGrant("linker", "", "search", 0)
	canSeeBlogs := cd.HasGrant("blogs", "", "see", 0) || cd.HasGrant("blogs", "", "view", 0) || cd.HasGrant("blogs", "", "search", 0)
	canSeeWritings := cd.HasGrant("writing", "", "see", 0) || cd.HasGrant("writing", "", "view", 0) || cd.HasGrant("writing", "", "search", 0)

	data := SearchPageData{
		CanSearchForum:    hasGeneralSearch && canSeeForum,
		CanSearchNews:     hasGeneralSearch && canSeeNews,
		CanSearchLinker:   hasGeneralSearch && canSeeLinker,
		CanSearchBlogs:    hasGeneralSearch && canSeeBlogs,
		CanSearchWritings: hasGeneralSearch && canSeeWritings,
	}
	data.AnySearch = data.CanSearchForum || data.CanSearchNews || data.CanSearchLinker || data.CanSearchBlogs || data.CanSearchWritings

	return data
}

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Search"

	data := GetSearchPageData(cd)

	handlers.TemplateHandler(w, r, "searchPage", data)
}


func SearchPage(w http.ResponseWriter, r *http.Request) {
	t := NewSearchTask().(*searchTask)
	t.Get(w, r)
}