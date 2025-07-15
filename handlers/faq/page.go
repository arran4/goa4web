package faq

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type FAQ struct {
		CategoryID int
		Question   string
		Answer     string
	}

	type CategoryFAQs struct {
		Category *db.GetAllAnsweredFAQWithFAQCategoriesRow
		FAQs     []*FAQ
	}

	type Data struct {
		*corecommon.CoreData
		FAQ []*CategoryFAQs
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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

	common.TemplateHandler(w, r, "faqPage", data)
}

func CustomFAQIndex(data *corecommon.CoreData) {
	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
		Name: "Ask",
		Link: "/faq/ask",
	})
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "Question Qontrols",
			Link: "/admin/faq/questions",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "Answer",
			Link: "/admin/faq/answer",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "Category Controls",
			Link: "/admin/faq/categories",
		})
	}
}
