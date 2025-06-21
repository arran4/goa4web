package main

import (
	"log"
	"net/http"
)

func adminUsageStatsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		ForumTopics       []*BoardPostCountRow
		ForumCategories   []*CategoryCountRow
		WritingCategories []*CategoryCountRow
		LinkerCategories  []*GetLinkerCategoryLinkCountsRow
		Imageboards       []*BoardPostCountRow
		Users             []*UserPostCountRow
		Monthly           []*MonthlyUsageRow
		UserMonthly       []*UserMonthlyUsageRow
		StartYear         int
	}
	data := Data{CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData)}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	var err error
	if data.ForumTopics, err = queries.ForumTopicThreadCounts(r.Context()); err != nil {
		log.Printf("forum topic counts: %v", err)
	}
	if data.ForumCategories, err = queries.ForumCategoryThreadCounts(r.Context()); err != nil {
		log.Printf("forum category counts: %v", err)
	}
	if data.Imageboards, err = queries.ImageboardPostCounts(r.Context()); err != nil {
		log.Printf("imageboard post counts: %v", err)
	}
	if data.Users, err = queries.UserPostCounts(r.Context()); err != nil {
		log.Printf("user post counts: %v", err)
	}
	if data.WritingCategories, err = queries.WritingCategoryCounts(r.Context()); err != nil {
		log.Printf("writing category counts: %v", err)
	}
	if data.LinkerCategories, err = queries.GetLinkerCategoryLinkCounts(r.Context()); err != nil {
		log.Printf("linker category counts: %v", err)
	}
	if data.Monthly, err = queries.MonthlyUsageCounts(r.Context(), int32(statsStartYear)); err != nil {
		log.Printf("monthly usage counts: %v", err)
	}
	if data.UserMonthly, err = queries.UserMonthlyUsageCounts(r.Context(), int32(statsStartYear)); err != nil {
		log.Printf("user monthly usage counts: %v", err)
	}
	data.StartYear = statsStartYear

	if err := renderTemplate(w, r, "adminUsageStatsPage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
