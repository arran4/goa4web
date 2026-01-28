package linker

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
)

// adminLinkPage displays the edit form for a linker item.
func adminLinkPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	link, id, err := cd.SelectedAdminLinkerItem(r)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}

	cd.PageTitle = fmt.Sprintf("Edit Link %d", id)
	data := struct {
		Link               *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
		Selected           int
		SelectedLanguageId int
	}{
		Link:               link,
		Selected:           int(link.CategoryID.Int32),
		SelectedLanguageId: int(link.LanguageID.Int32),
	}

	LinkerAdminLinkPageTmpl.Handle(w, r, data)
}

const LinkerAdminLinkPageTmpl tasks.Template = "linker/adminLinkPage.gohtml"

// editLinkTask updates an existing linker item.
type editLinkTask struct{ tasks.TaskString }

var AdminEditLinkTask = &editLinkTask{TaskString: TaskUpdate}

var _ tasks.Task = (*editLinkTask)(nil)

func (editLinkTask) Page(w http.ResponseWriter, r *http.Request) { adminLinkPage(w, r) }

func (editLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	id, err := cd.SelectedAdminLinkerItemID(r)
	if err != nil {
		return fmt.Errorf("link id not found %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	title := r.PostFormValue("title")
	URL := r.PostFormValue("URL")
	desc := r.PostFormValue("desc")
	cat, _ := strconv.Atoi(r.PostFormValue("category"))
	lang, _ := strconv.Atoi(r.PostFormValue("language"))
	queries := cd.Queries()

	if err := queries.AdminUpdateLinkerItem(r.Context(), db.AdminUpdateLinkerItemParams{
		Title:       sql.NullString{Valid: true, String: title},
		Url:         sql.NullString{Valid: true, String: URL},
		Description: sql.NullString{Valid: true, String: desc},
		CategoryID:  sql.NullInt32{Int32: int32(cat), Valid: cat != 0},
		LanguageID:  sql.NullInt32{Int32: int32(lang), Valid: lang != 0},
		ID:          id,
	}); err != nil {
		return fmt.Errorf("update linker item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RedirectHandler(fmt.Sprintf("/admin/linker/links/link/%d", id))
}
