package admin

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"sync"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
)

func AdminUsageStatsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Errors            []string
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
	queries := data.Queries()

	var wg sync.WaitGroup
	wg.Add(8)
	errCh := make(chan string, 8)

	go func() {
		defer wg.Done()
		if rows, err := queries.ForumTopicThreadCounts(r.Context()); err == nil {
			data.ForumTopics = rows
		} else {
			log.Printf("forum topic counts: %v", err)
			errCh <- fmt.Errorf("forum topic counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := queries.ForumCategoryThreadCounts(r.Context()); err == nil {
			data.ForumCategories = rows
		} else {
			log.Printf("forum category counts: %v", err)
			errCh <- fmt.Errorf("forum category counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := queries.ImageboardPostCounts(r.Context()); err == nil {
			data.Imageboards = rows
		} else {
			log.Printf("imageboard post counts: %v", err)
			errCh <- fmt.Errorf("imageboard post counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := queries.UserPostCounts(r.Context()); err == nil {
			data.Users = rows
		} else {
			log.Printf("user post counts: %v", err)
			errCh <- fmt.Errorf("user post counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := queries.WritingCategoryCounts(r.Context()); err == nil {
			data.WritingCategories = rows
		} else {
			log.Printf("writing category counts: %v", err)
			errCh <- fmt.Errorf("writing category counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := data.LinkerCategoryCounts(); err == nil {
			data.LinkerCategories = rows
		} else {
			log.Printf("linker category counts: %v", err)
			errCh <- fmt.Errorf("linker category counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := queries.MonthlyUsageCounts(r.Context(), int32(cd.Config.StatsStartYear)); err == nil {
			data.Monthly = rows
		} else {
			log.Printf("monthly usage counts: %v", err)
			errCh <- fmt.Errorf("monthly usage counts: %w", err).Error()
		}
	}()

	go func() {
		defer wg.Done()
		if rows, err := queries.UserMonthlyUsageCounts(r.Context(), int32(cd.Config.StatsStartYear)); err == nil {
			data.UserMonthly = rows
		} else {
			log.Printf("user monthly usage counts: %v", err)
			errCh <- fmt.Errorf("user monthly usage counts: %w", err).Error()
		}
	}()

	wg.Wait()
	close(errCh)
	for e := range errCh {
		data.Errors = append(data.Errors, e)
	}
	data.StartYear = cd.Config.StatsStartYear

	handlers.TemplateHandler(w, r, "usageStatsPage.gohtml", data)
}
