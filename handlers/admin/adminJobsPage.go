package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/tasks"
)

// adminJobsListLimit caps the number of durable jobs rendered on /admin/jobs.
const adminJobsListLimit = 100

// AdminJobsPageTmpl is the template path for the durable jobs admin page.
const AdminJobsPageTmpl tasks.Template = "domains/admin/jobsPage.gohtml"

// AdminJobsPage renders the durable background jobs management page.
func AdminJobsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Background Jobs"

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	typeFilter := strings.TrimSpace(r.URL.Query().Get("type"))
	svc := cd.JobService()
	if svc == nil {
		svc = jobs.NewConfiguredService(cd.Queries())
	}

	jobRows, err := svc.AdminList(r.Context(), statusFilter, typeFilter, adminJobsListLimit, 0)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list durable jobs: %w", err))
		return
	}

	data := struct {
		Jobs         []*db.DurableJob
		StatusFilter string
		TypeFilter   string
		FlashError   string
	}{
		Jobs:         jobRows,
		StatusFilter: statusFilter,
		TypeFilter:   typeFilter,
		FlashError:   strings.TrimSpace(r.URL.Query().Get("error")),
	}
	_ = AdminJobsPageTmpl.Handle(w, r, data)
}

// AdminJobsAction handles operator actions (enqueue-url, retry, cancel, purge) on /admin/jobs.
func AdminJobsAction(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := r.ParseForm(); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	svc := cd.JobService()
	if svc == nil {
		svc = jobs.NewConfiguredService(cd.Queries())
	}

	action := strings.TrimSpace(r.FormValue("action"))
	switch action {
	case "enqueue_url":
		rawURL := strings.TrimSpace(r.FormValue("url"))
		canonical := common.ValidateEnqueueExternalURL(rawURL)
		if canonical == "" {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape("Invalid or disallowed external URL"), http.StatusSeeOther)
			return
		}
		force := r.FormValue("force") == "1" || r.FormValue("force") == "true"
		if _, err := svc.Enqueue(r.Context(), jobs.EnqueueRequest{
			JobType:    jobs.JobTypeExternalLink,
			TargetType: jobs.JobTypeExternalLink,
			TargetKey:  canonical,
			ForceRetry: force,
		}); err != nil {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
			return
		}
	case "retry":
		id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("id")), 10, 64)
		if err != nil || id <= 0 {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape("Invalid job ID"), http.StatusSeeOther)
			return
		}
		if _, err := svc.AdminRetry(r.Context(), id); err != nil {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
			return
		}
	case "cancel":
		id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("id")), 10, 64)
		if err != nil || id <= 0 {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape("Invalid job ID"), http.StatusSeeOther)
			return
		}
		if _, err := svc.AdminCancel(r.Context(), id); err != nil {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
			return
		}
	case "purge":
		statusFilter := strings.TrimSpace(r.FormValue("status"))
		if _, err := svc.AdminPurge(r.Context(), statusFilter); err != nil {
			http.Redirect(w, r, "/admin/jobs?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, "/admin/jobs", http.StatusSeeOther)
}
