package faq

import (
	"database/sql"
	"errors"
	"fmt"
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

	data := Data{}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "FAQ"

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	var currentCategoryFAQs CategoryFAQs

	faqRows, err := queries.GetAllAnsweredFAQWithFAQCategoriesForUser(r.Context(), db.GetAllAnsweredFAQWithFAQCategoriesForUserParams{
		ViewerID: cd.UserID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllAnsweredFAQWithFAQCategories Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
	if data.HasGrant(common.SectionFAQ, common.ItemQuestion, common.ActionPost, 0) {
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
