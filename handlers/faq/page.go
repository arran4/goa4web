package faq

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type FAQ struct {
		CategoryID int
		Question   string
		Answer     string
	}

	type CategoryFAQs struct {
		Category *db.GetAllAnsweredFAQWithFAQCategoriesForUserRow
		FAQs     []*FAQ
	}

	type Data struct {
		FAQ []*CategoryFAQs
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{}
	cd.PageTitle = "FAQ"

	var currentCategoryFAQs CategoryFAQs

	faqRows, err := cd.AllAnsweredFAQWithCategories()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllAnsweredFAQWithFAQCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	for _, row := range faqRows {
		if currentCategoryFAQs.Category == nil || currentCategoryFAQs.Category.Idfaqcategories.Int32 != row.Idfaqcategories.Int32 {
			if currentCategoryFAQs.Category != nil && currentCategoryFAQs.Category.Idfaqcategories.Int32 != 0 {
				data.FAQ = append(data.FAQ, &currentCategoryFAQs)
			}
			currentCategoryFAQs = CategoryFAQs{Category: row}
		}
		currentCategoryFAQs.FAQs = append(currentCategoryFAQs.FAQs, &FAQ{CategoryID: int(row.Idfaqcategories.Int32), Question: row.Question.String, Answer: row.Answer.String})
	}

	if currentCategoryFAQs.Category != nil && currentCategoryFAQs.Category.Idfaqcategories.Int32 != 0 {
		data.FAQ = append(data.FAQ, &currentCategoryFAQs)
	}

	// index links provided via middleware

	handlers.TemplateHandler(w, r, "faqPage", data)
}

func CustomFAQIndex(data *common.CoreData, r *http.Request) {
	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	data.CustomIndexItems = []common.IndexItem{}
	if data.HasGrant("faq", "question", "post", 0) {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Ask",
			Link: "/faq/ask",
		})
	}
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Question Qontrols",
			Link: "/admin/faq/questions",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Category Controls",
			Link: "/admin/faq/categories",
		})
	}
}
