package stats

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/arran4/goa4web/internal/db"
)

// BuildUsageStatsData collects usage statistics from the database.
func BuildUsageStatsData(ctx context.Context, queries db.Querier, customQueries db.CustomQueries, startYear int) UsageStatsData {
	data := UsageStatsData{
		StartYear: startYear,
	}

	var wg sync.WaitGroup
	errCh := make(chan string)
	var errWG sync.WaitGroup

	errWG.Add(1)
	go func() {
		defer errWG.Done()
		for e := range errCh {
			data.Errors = append(data.Errors, e)
		}
	}()

	addErr := func(name string, err error) {
		errCh <- fmt.Errorf("%s: %w", name, err).Error()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.AdminForumTopicThreadCounts(ctx); err == nil {
			data.ForumTopics = rows
		} else {
			addErr("forum topic counts", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.AdminForumHandlerThreadCounts(ctx); err == nil {
			data.ForumHandlers = rows
		} else {
			addErr("forum handler counts", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.AdminForumCategoryThreadCounts(ctx); err == nil {
			data.ForumCategories = rows
		} else {
			addErr("forum category counts", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.AdminImageboardPostCounts(ctx); err == nil {
			data.Imageboards = rows
		} else {
			addErr("imageboard post counts", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.AdminUserPostCounts(ctx); err == nil {
			data.Users = rows
		} else {
			addErr("user post counts", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.AdminWritingCategoryCounts(ctx); err == nil {
			data.WritingCategories = rows
		} else {
			addErr("writing category counts", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if rows, err := queries.GetLinkerCategoryLinkCounts(ctx); err == nil {
			data.LinkerCategories = rows
		} else {
			addErr("linker category counts", err)
		}
	}()

	if customQueries != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rows, err := customQueries.MonthlyUsageCounts(ctx, int32(startYear)); err == nil {
				data.Monthly = rows
			} else {
				addErr("monthly usage counts", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if rows, err := customQueries.UserMonthlyUsageCounts(ctx, int32(startYear)); err == nil {
				data.UserMonthly = rows
			} else {
				addErr("user monthly usage counts", err)
			}
		}()
	} else {
		log.Printf("stats: customQueries is nil, skipping monthly usage counts")
	}

	wg.Wait()

	ensureHandler := func(h string) {
		for _, r := range data.ForumHandlers {
			if r.Handler == h {
				return
			}
		}
		data.ForumHandlers = append(data.ForumHandlers, &db.AdminForumHandlerThreadCountsRow{Handler: h, Threads: 0, Comments: 0})
	}
	ensureHandler("private")
	ensureHandler("all")
	sort.Slice(data.ForumHandlers, func(i, j int) bool { return data.ForumHandlers[i].Handler < data.ForumHandlers[j].Handler })

	close(errCh)
	errWG.Wait()

	return data
}
