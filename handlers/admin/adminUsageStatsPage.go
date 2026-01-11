package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

const (
	// usageTimeout defines the maximum duration allowed for loading usage statistics
	usageTimeout = 5 * time.Minute
)

func AdminUsageStatsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Errors            []string
		ForumTopics       []*db.AdminForumTopicThreadCountsRow
		ForumHandlers     []*db.AdminForumHandlerThreadCountsRow
		ForumCategories   []*db.AdminForumCategoryThreadCountsRow
		WritingCategories []*db.AdminWritingCategoryCountsRow
		LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
		Imageboards       []*db.AdminImageboardPostCountsRow
		Users             []*db.AdminUserPostCountsRow
		Monthly           []*db.MonthlyUsageRow
		UserMonthly       []*db.UserMonthlyUsageRow
		StartYear         int
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Usage Stats"
	data := Data{}
	queries := cd.Queries()

	ctx, cancel := context.WithTimeout(r.Context(), usageTimeout)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan string)
	var errWG sync.WaitGroup

	log.Print("start error consumer")
	errWG.Add(1)
	go func() {
		defer func() {
			log.Print("stop error consumer")
			errWG.Done()
		}()
		for e := range errCh {
			log.Printf("error reported: %s", e)
			data.Errors = append(data.Errors, e)
		}
	}()

	addErr := func(name string, err error) {
		log.Printf("%s: %v", name, err)
		errCh <- fmt.Errorf("%s: %w", name, err).Error()
	}

	log.Print("start forum topic counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop forum topic counts")
			wg.Done()
		}()
		if rows, err := queries.AdminForumTopicThreadCounts(ctx); err == nil {
			data.ForumTopics = rows
		} else {
			addErr("forum topic counts", err)
		}
	}()

	log.Print("start forum handler counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop forum handler counts")
			wg.Done()
		}()
		if rows, err := queries.AdminForumHandlerThreadCounts(ctx); err == nil {
			data.ForumHandlers = rows
		} else {
			addErr("forum handler counts", err)
		}
	}()

	log.Print("start forum category counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop forum category counts")
			wg.Done()
		}()
		if rows, err := queries.AdminForumCategoryThreadCounts(ctx); err == nil {
			data.ForumCategories = rows
		} else {
			addErr("forum category counts", err)
		}
	}()

	log.Print("start imageboard post counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop imageboard post counts")
			wg.Done()
		}()
		if rows, err := queries.AdminImageboardPostCounts(ctx); err == nil {
			data.Imageboards = rows
		} else {
			addErr("imageboard post counts", err)
		}
	}()

	log.Print("start user post counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop user post counts")
			wg.Done()
		}()
		if rows, err := queries.AdminUserPostCounts(ctx); err == nil {
			data.Users = rows
		} else {
			addErr("user post counts", err)
		}
	}()

	log.Print("start writing category counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop writing category counts")
			wg.Done()
		}()
		if rows, err := queries.AdminWritingCategoryCounts(ctx); err == nil {
			data.WritingCategories = rows
		} else {
			addErr("writing category counts", err)
		}
	}()

	log.Print("start linker category counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop linker category counts")
			wg.Done()
		}()
		if rows, err := queries.GetLinkerCategoryLinkCounts(ctx); err == nil {
			data.LinkerCategories = rows
		} else {
			addErr("linker category counts", err)
		}
	}()

	cq := cd.CustomQueries()

	log.Print("start monthly usage counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop monthly usage counts")
			wg.Done()
		}()
		if rows, err := cq.MonthlyUsageCounts(ctx, int32(cd.Config.StatsStartYear)); err == nil {
			data.Monthly = rows
		} else {
			addErr("monthly usage counts", err)
		}
	}()

	log.Print("start user monthly usage counts")
	wg.Add(1)
	go func() {
		defer func() {
			log.Print("stop user monthly usage counts")
			wg.Done()
		}()
		if rows, err := cq.UserMonthlyUsageCounts(ctx, int32(cd.Config.StatsStartYear)); err == nil {
			data.UserMonthly = rows
		} else {
			addErr("user monthly usage counts", err)
		}
	}()

	log.Print("wait for goroutines")
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

	log.Print("close error channel")
	close(errCh)
	errWG.Wait()
	data.StartYear = cd.Config.StatsStartYear

	log.Print("render usage stats page")
	AdminUsageStatsPageTmpl.Handle(w, r, data)
}

const AdminUsageStatsPageTmpl handlers.Page = "admin/usageStatsPage.gohtml"
