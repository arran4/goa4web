package goa4web

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func faqPage(w http.ResponseWriter, r *http.Request) {
	type FAQ struct {
		CategoryID int
		Question   string
		Answer     string
	}

	type CategoryFAQs struct {
		Category *GetAllAnsweredFAQWithFAQCategoriesRow
		FAQs     []*FAQ
	}

	type Data struct {
		*CoreData
		FAQ []*CategoryFAQs
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	var currentCategoryFAQs CategoryFAQs

	faqRows, err := queries.GetAllAnsweredFAQWithFAQCategories(r.Context())
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

	CustomFAQIndex(data.CoreData)

	if err := templates.RenderTemplate(w, "page.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomFAQIndex(data *CoreData) {
	userHasAdmin := data.HasRole("administrator")
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Ask",
		Link: "/faq/ask",
	})
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Question Qontrols",
			Link: "/admin/faq/questions",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Answer",
			Link: "/admin/faq/answer",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Category Controls",
			Link: "/admin/faq/categories",
		})
	}
}
