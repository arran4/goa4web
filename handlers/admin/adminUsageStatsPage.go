package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
)

func AdminUsageStatsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		ForumTopics       []*db.ForumTopicThreadCountsRow
		ForumCategories   []*db.ForumCategoryThreadCountsRow
		WritingCategories []*db.WritingCategoryCountsRow
		LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
		Imageboards       []*db.ImageboardPostCountsRow
		Users             []*db.UserPostCountsRow
		Monthly           []*db.MonthlyUsageRow
		UserMonthly       []*db.UserMonthlyUsageRow
		StartYear         int
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

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
	if data.LinkerCategories, err = data.LinkerCategoryCounts(); err != nil {
		log.Printf("linker category counts: %v", err)
	}
	if data.Monthly, err = queries.MonthlyUsageCounts(r.Context(), int32(config.AppRuntimeConfig.StatsStartYear)); err != nil {
		log.Printf("monthly usage counts: %v", err)
	}
	if data.UserMonthly, err = queries.UserMonthlyUsageCounts(r.Context(), int32(config.AppRuntimeConfig.StatsStartYear)); err != nil {
		log.Printf("user monthly usage counts: %v", err)
	}
	data.StartYear = config.AppRuntimeConfig.StatsStartYear

	handlers.TemplateHandler(w, r, "usageStatsPage.gohtml", data)
}
