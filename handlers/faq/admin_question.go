package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// AdminQuestionPageTask encapsulates the data for viewing a single FAQ question.
type AdminQuestionPageTask struct {
	common.CoreData
	Faq      *db.Faq
	Category *db.FaqCategory
	Author   *db.SystemGetUserByIDRow
}

// Breadcrumbs generates navigation breadcrumbs for the FAQ Question page.
func (p *AdminQuestionPageTask) Breadcrumbs() []common.Breadcrumb {
	crumbs := []common.Breadcrumb{
		{Title: "Admin", Link: "/admin"},
		{Title: "FAQ Categories", Link: "/admin/faq/categories"},
	}
	if p.Category != nil {
		crumbs = append(crumbs, common.Breadcrumb{
			Title: p.Category.Name.String,
			Link:  fmt.Sprintf("/admin/faq/categories/category/%d", p.Category.ID),
		})
	} else {
		crumbs = append(crumbs, common.Breadcrumb{
			Title: "Unassigned",
		})
	}
	crumbs = append(crumbs, common.Breadcrumb{
		Title: fmt.Sprintf("Question %d", p.Faq.ID),
	})
	return crumbs
}

// AdminQuestionPage displays a single FAQ question using the Task interface pattern.
func AdminQuestionPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	p := &AdminQuestionPageTask{
		CoreData: *cd,
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid question id"))
		return
	}

	queries := cd.Queries()
	p.Faq, err = queries.AdminGetFAQByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("question not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	if p.Faq.CategoryID.Valid {
		p.Category, err = queries.AdminGetFAQCategory(r.Context(), p.Faq.CategoryID.Int32)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	p.Author = cd.UserByID(p.Faq.AuthorID)
	cd.PageTitle = fmt.Sprintf("FAQ: %s", p.Faq.Question.String)

	// Inject the task into the event so CoreData.Breadcrumbs() can find it if needed.
	// We use the event bus to set the current task on the event.
	if evt := cd.Event(); evt != nil {
		evt.Task = p
	}

	AdminQuestionPageTmpl.Handle(w, r, p)
}

// Name returns the task name for the interface.
func (p *AdminQuestionPageTask) Name() string {
	return "admin-question-page"
}

// Action implements tasks.Task.
func (p *AdminQuestionPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

// Ensure AdminQuestionPageTask implements HasBreadcrumbs interface if defined
// (Not checking explicitly here as I don't have the interface definition handy,
// but based on `core/common/breadcrumb.go` logic, it checks `cd.event.Task` for `HasBreadcrumb`?)
// Let's check `core/common/breadcrumb.go` again.
// It seems `core/common/breadcrumb.go` logic is hardcoded switches!
// Wait, I missed something. I should re-read `core/common/breadcrumb.go`.
// Ah, `CoreData.Breadcrumbs()` calls `cd.adminBreadcrumbs()`.
// `cd.adminBreadcrumbs()` has a switch on `cd.currentSection` (which is "admin") and then checks conditions.
// It DOES NOT check `cd.event.Task`.
//
// However, the user said "shouldn't this be in Page form so that we can give it HasBreadcrumbs".
// This implies there IS a mechanism.
// Let's check `core/common/coredata.go` to see if there's a `HasBreadcrumbs` interface used anywhere.
// Maybe I missed it in `breadcrumb.go`.
//
// Let's check `core/common/breadcrumb.go` again very carefully.
/*
func (cd *CoreData) Breadcrumbs() []Breadcrumb {
    // ...
    switch cd.currentSection {
    // ...
    case "admin":
        crumbs, err = cd.adminBreadcrumbs()
    // ...
    }
    // ...
}
*/
// It seems hardcoded.
//
// BUT, maybe the user implies I should *add* it or use a pattern that *allows* it.
// Or maybe I missed where `cd.event.Task` is used.
//
// Let's search for `HasBreadcrumbs` in the codebase.
// I can't search easily.
//
// Let's assume the user knows what they are talking about.
// If I make it a struct with `Breadcrumbs()`, I might need to update `CoreData.Breadcrumbs()` to support it.
//
// "also ensure it's on the "sitemap" on the admin landing page using RegisterAdminControlCenter"
// I already did `RegisterAdminControlCenter` in `routes.go` for the *list* page.
//
// Re-reading the PR comment: "shouldn't this be in Page form so that we can give it HasBreadcrumbs"
//
// If I look at `handlers/admin/pages_admin.go`, maybe there is a generic way.
//
// I will implement the struct and method.
// And I will Update `core/common/breadcrumb.go` to check for the interface if `cd.event.Task` is set.
// OR I will just implement it and see.
//
// Actually, `handlers/admin/pages_admin.go` might have been refactored recently (Memory says: "Admin pages are refactored into structs implementing tasks.Task...").
// Memory says: "The `CoreData.Breadcrumbs()` method traverses the breadcrumb hierarchy by prioritizing `cd.currentPage` (set via `SetCurrentPage`) or `cd.event.Task` (if it implements `HasBreadcrumb`) before falling back to legacy path inference."
//
// AHA! Memory confirms: "`cd.event.Task` (if it implements `HasBreadcrumb`)".
// I missed this in my read of `breadcrumb.go` because I might have read an *old* version or I just missed the check (it might be at the top of `Breadcrumbs` method or inside `adminBreadcrumbs`).
//
// Let's re-read `core/common/breadcrumb.go` one more time to be absolutely sure.
//
// ...
// func (cd *CoreData) Breadcrumbs() []Breadcrumb {
//     if cd == nil || cd.queries == nil {
//         return nil
//     }
//     var (
//         crumbs []Breadcrumb
//         err    error
//     )
//     switch cd.currentSection {
// ...
//
// It looks like `cd.Breadcrumbs()` in the file I read DOES NOT have the task check.
// This means the Memory might be referring to a *different* codebase state or I read the wrong file (maybe `core/common/breadcrumb.go` vs `core/common/breadcrumbs.go`? No, I listed files).
//
// Wait, maybe `SetCurrentPage` mentioned in memory is the key.
//
// If the code I read is accurate, `Breadcrumbs` logic is indeed hardcoded.
// However, the user asks me to "give it HasBreadcrumbs".
//
// I will implement the struct and method `Breadcrumbs`.
// And I will Update `core/common/breadcrumb.go` to use it!
//
// Step 1: Update `handlers/faq/admin_question.go` to use struct `AdminQuestionPageTask`.
// Step 2: Implement `Breadcrumbs()` on it.
// Step 3: Update `core/common/breadcrumb.go` to check `cd.event.Task` for `HasBreadcrumbs` interface.
//
// Wait, I should verify if `HasBreadcrumbs` interface exists.
// I'll check `core/common/interfaces.go` or similar? Or just define it inline.
//
// Let's look for `type HasBreadcrumbs interface` or `Breadcrumb` usage.
//
// I'll search for `HasBreadcrumb` in `core/common` files.
// `grep -r "HasBreadcrumb" core/common`
//
// Also, I need to make sure `AdminQuestionPage` handler sets the task.
// `cd.SetEventTask(p)` seems correct.

const AdminQuestionPageTmpl tasks.Template = "faq/adminQuestionPage.gohtml"

// Matcher? It's a GET request, so maybe no matcher needed if registered as standard handler?
// But `tasks.Task` usually implies `Matcher` and `Action`.
// `AdminQuestionPage` is a GET view.
//
// `handlers.TaskHandler` wraps a task.
// But here we are using a standard handler function `AdminQuestionPage`.
//
// If I want it to be a "Page form" task, I should probably implement `tasks.Task` (Match, Action) or `tasks.Page`?
//
// If I use `handlers.TaskHandler`, it expects `Action` to return data/error.
//
// For a GET page, usually we just execute the template.
//
// I will stick to the struct pattern for data and breadcrumbs, and manual template execution in the handler, but I will invoke `cd.SetEventTask(p)`.
//
// And I will update `core/common/breadcrumb.go` to support it.
