package externallink

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	domainexternallink "github.com/arran4/goa4web/internal/externallink"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ReloadExternalLinkTask reloads OG metadata for a link.
type ReloadExternalLinkTask struct {
	tasks.TaskString
	URL string
}

var reloadExternalLinkTask = ReloadExternalLinkTask{TaskString: "admin:externallink:reload"}

// ensure conformance
var _ tasks.Task = ReloadExternalLinkTask{}
var _ tasks.BackgroundTasker = ReloadExternalLinkTask{}

func (t ReloadExternalLinkTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if t.URL == "" {
		return nil, nil
	}
	return processExternalLink(ctx, q, t.URL)
}

func (t ReloadExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL, _, _, _, err := cd.ResolveExternalLink(r)
	if err != nil {
		return fmt.Errorf("invalid link: %w", handlers.ErrForbidden)
	}
	if jobSvc := cd.JobService(); jobSvc != nil {
		canonicalURL := common.ValidateEnqueueExternalURL(rawURL)
		if canonicalURL == "" {
			return fmt.Errorf("invalid link: %w", handlers.ErrForbidden)
		}
		actorID := sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}
		if err := jobSvc.QueueObservedExternalLink(r.Context(), canonicalURL, actorID); err != nil && !errors.Is(err, jobs.ErrInCooldown) {
			log.Printf("queue reload external link: %v", err)
		}
		// Avoid attaching ReloadExternalLinkTask (a BackgroundTasker) when durable jobs own execution.
		cd.SetEventTask(tasks.TaskString(t.TaskString))
	} else {
		canonicalURL := common.CanonicalizeExternalURL(rawURL)
		if canonicalURL == "" {
			return fmt.Errorf("invalid link: %w", handlers.ErrForbidden)
		}

		cd.SetEventTask(ReloadExternalLinkTask{
			TaskString: t.TaskString,
			URL:        canonicalURL,
		})
	}

	// Return redirect to the current URL with msg parameter
	redirectURI := r.RequestURI
	if r.URL.Query().Get("msg") == "" {
		redirectURI += "&msg=Reloading+Open+Graph+data+in+the+background..."
	}
	return handlers.RedirectHandler(redirectURI)
}

func (t ReloadExternalLinkTask) Matcher() func(*http.Request, *mux.RouteMatch) bool {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}
}

// PrefetchExternalLinkTask prefetches OG metadata for a link.
type PrefetchExternalLinkTask struct {
	tasks.TaskString
	URL string
}

var prefetchExternalLinkTask = PrefetchExternalLinkTask{TaskString: "externallink:prefetch"}

// ensure conformance
var _ tasks.Task = PrefetchExternalLinkTask{}
var _ tasks.BackgroundTasker = PrefetchExternalLinkTask{}

func (t PrefetchExternalLinkTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if t.URL == "" {
		return nil, nil
	}
	return processExternalLink(ctx, q, t.URL)
}

func (t PrefetchExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL := r.FormValue("url")
	if rawURL == "" {
		if evt := cd.Event(); evt != nil {
			evt.Task = tasks.TaskString("MISSING")
		}
		return handlers.TextByteWriter("missing url")
	}
	if jobSvc := cd.JobService(); jobSvc != nil {
		if cd.UserID == 0 {
			if evt := cd.Event(); evt != nil {
				evt.Task = tasks.TaskString("MISSING")
			}
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			})
		}
		canonicalURL := common.ValidateEnqueueExternalURL(rawURL)
		if canonicalURL == "" {
			if evt := cd.Event(); evt != nil {
				evt.Task = tasks.TaskString("MISSING")
			}
			return handlers.TextByteWriter("invalid url")
		}
		actorID := sql.NullInt32{Int32: cd.UserID, Valid: true}
		if err := jobSvc.QueueObservedExternalLink(r.Context(), canonicalURL, actorID); err != nil && !errors.Is(err, jobs.ErrInCooldown) {
			log.Printf("queue prefetch external link: %v", err)
		}
		cd.SetEventTask(tasks.TaskString(t.TaskString))
		return handlers.TextByteWriter("ok")
	}
	canonicalURL := common.CanonicalizeExternalURL(rawURL)
	if canonicalURL == "" {
		if evt := cd.Event(); evt != nil {
			evt.Task = tasks.TaskString("MISSING")
		}
		return handlers.TextByteWriter("invalid url")
	}

	cd.SetEventTask(PrefetchExternalLinkTask{
		TaskString: t.TaskString,
		URL:        canonicalURL,
	})

	return handlers.TextByteWriter("ok")
}

func processExternalLink(ctx context.Context, q db.Querier, targetURL string) (tasks.Task, error) {
	var cd *common.CoreData
	if v, ok := ctx.Value(consts.KeyCoreData).(*common.CoreData); ok && v != nil {
		cd = v
	} else {
		cd = common.NewCoreData(ctx, q, nil)
	}

	cacheErr, err := domainexternallink.RefreshMetadata(ctx, cd, targetURL, func(ctx context.Context) (int32, error) {
		id, err := domainexternallink.ResolveRecord(ctx, cd, targetURL)
		if err != nil {
			// The legacy task skips persistence if it cannot resolve a record.
			log.Printf("resolve link for background metadata refresh: %v", err)
		}
		return id, nil
	}, func(ctx context.Context, target string) (*opengraph.Info, error) {
		return opengraph.Fetch(ctx, target, cd.HTTPClient())
	}, func(_ context.Context, image string) (string, error) {
		return cd.DownloadAndCacheImage(image)
	})
	if cacheErr != nil {
		log.Printf("refresh external link image cache: %v", cacheErr)
	}
	return nil, err
}
