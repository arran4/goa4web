package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

func Page(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		Categories              []*ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		Category                *ForumcategoryPlus
		Back                    bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum"
	cd.LoadSelectionsFromRequest(r)
	_, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	imageURL, _ := common.MakeImageURL(cd.AbsoluteURL(), "Forum", "A place for discussion.", cd.ShareSignKey, false)
	cd.OpenGraph = &common.OpenGraph{
		Title:       "Forum",
		Description: "A place for discussion.",
		Image:       imageURL,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "website",
	}

	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	data := &Data{
		Admin: cd.IsAdmin() && cd.IsAdminMode(),
	}

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = rootCategory.Categories
		d.Category = rootCategory
		d.Back = false
		return &d
	}
	data.CopyDataToSubCategories = copyDataToSubCategories

	categoryRows, err := cd.ForumCategories()
	if err != nil {
		log.Printf("getAllForumCategories Error: %s", err)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	var topicRows []*ForumtopicPlus
	rows, err := cd.ForumTopics(int32(categoryId))
	if err != nil {
		log.Printf("ForumTopics Error: %s", err)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	for _, row := range rows {
		var lbls []templates.TopicLabel
		if pub, _, err := cd.ThreadPublicLabels(row.Idforumtopic); err == nil {
			for _, l := range pub {
				lbls = append(lbls, templates.TopicLabel{Name: l, Type: "public"})
			}
		} else {
			log.Printf("list public labels: %v", err)
		}
		topicRows = append(topicRows, &ForumtopicPlus{
			Idforumtopic:                 row.Idforumtopic,
			Lastposter:                   row.Lastposter,
			ForumcategoryIdforumcategory: row.ForumcategoryIdforumcategory,
			Title:                        row.Title,
			Description:                  row.Description,
			Threads:                      row.Threads,
			Comments:                     row.Comments,
			Lastaddition:                 row.Lastaddition,
			Lastposterusername:           row.Lastposterusername,
			Labels:                       lbls,
		})
	}

	categoryTree := NewCategoryTree(categoryRows, topicRows)

	if categoryId == 0 {
		data.Categories = categoryTree.CategoryChildrenLookup[int32(categoryId)]
	} else if cat, ok := categoryTree.CategoryLookup[int32(categoryId)]; ok && cat != nil {
		data.Categories = []*ForumcategoryPlus{cat}
		data.Category = cat
		data.Back = true
	}

	ForumPageTmpl.Handle(w, r, data)
}

const ForumPageTmpl tasks.Template = "forum/page.gohtml"
