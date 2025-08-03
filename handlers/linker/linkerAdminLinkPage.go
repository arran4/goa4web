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
	"github.com/gorilla/mux"
)

// adminLinkPage displays the edit form for a linker item.
func adminLinkPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["link"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	cats, _ := queries.GetAllLinkerCategories(r.Context())
	langs, _ := cd.Languages()

	cd.PageTitle = fmt.Sprintf("Edit Link %d", id)
	data := struct {
		*common.CoreData
		Link               *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
		Categories         []*db.LinkerCategory
		Languages          []*db.Language
		Selected           int
		SelectedLanguageId int
	}{
		CoreData:           cd,
		Link:               link,
		Categories:         cats,
		Languages:          langs,
		Selected:           int(link.LinkerCategoryID),
		SelectedLanguageId: int(link.LanguageIdlanguage),
	}

	handlers.TemplateHandler(w, r, "adminLinkPage.gohtml", data)
}

// editLinkTask updates an existing linker item.
type editLinkTask struct{ tasks.TaskString }

var AdminEditLinkTask = &editLinkTask{TaskString: TaskUpdate}

var _ tasks.Task = (*editLinkTask)(nil)

func (editLinkTask) Page(w http.ResponseWriter, r *http.Request) { adminLinkPage(w, r) }

func (editLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["link"])
	title := r.PostFormValue("title")
	URL := r.PostFormValue("URL")
	desc := r.PostFormValue("desc")
	cat, _ := strconv.Atoi(r.PostFormValue("category"))
	lang, _ := strconv.Atoi(r.PostFormValue("language"))

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	if err := queries.AdminUpdateLinkerItem(r.Context(), db.AdminUpdateLinkerItemParams{
		Title:              sql.NullString{Valid: true, String: title},
		Url:                sql.NullString{Valid: true, String: URL},
		Description:        sql.NullString{Valid: true, String: desc},
		LinkerCategoryID:   int32(cat),
		LanguageIdlanguage: int32(lang),
		Idlinker:           int32(id),
	}); err != nil {
		return fmt.Errorf("update linker item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RedirectHandler(fmt.Sprintf("/admin/linker/link/%d", id))
}
