package main

import (
	"log"
	"net/http"
)

func faqPage(w http.ResponseWriter, r *http.Request) {
	type FAQ struct {
		CategoryID int
		Question   string
		Answer     string
	}

	type CategoryFAQs struct {
		Category *show_questionsRow
		FAQs     []*FAQ
	}

	type Data struct {
		*CoreData
		FAQ []*CategoryFAQs
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	var currentCategoryFAQs CategoryFAQs

	faqRows, err := queries.show_questions(r.Context())
	if err != nil {
		log.Printf("show_questions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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

	if err := getCompiledTemplates().ExecuteTemplate(w, "faqPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomFAQIndex(data *CoreData) {
	userHasAdmin := true // TODO
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Ask",
		Link: "/faq/ask",
	})
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Question Qontrols",
			Link: "/faq/admin/questions",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Answer",
			Link: "/faq/admin/answer",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Category Controls",
			Link: "/faq/admin/categories",
		})
	}
}
