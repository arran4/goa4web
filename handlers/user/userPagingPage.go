package user

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/tasks"
)

type PagingSaveTask struct{ tasks.TaskString }

var pagingSaveTask = &PagingSaveTask{TaskString: tasks.TaskString(TaskSaveSize)}
var _ tasks.Task = (*PagingSaveTask)(nil)

func userPagingPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Pagination"
	handlers.TemplateHandler(w, r, "user/pagingPage.gohtml", struct{}{})
}

func (PagingSaveTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	size, _ := strconv.Atoi(r.FormValue("size"))
	if size < cd.Config.PageSizeMin {
		size = cd.Config.PageSizeMin
	}
	if size > cd.Config.PageSizeMax {
		size = cd.Config.PageSizeMax
	}
	queries := cd.Queries()

	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("preference load: %v", err)
		return fmt.Errorf("preference load fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if errors.Is(err, sql.ErrNoRows) {
		err = queries.InsertPreferenceForLister(r.Context(), db.InsertPreferenceForListerParams{
			LanguageID: sql.NullInt32{},
			ListerID:   uid,
			PageSize:   int32(size),
			Timezone:   sql.NullString{},
		})
	} else {
		pref.PageSize = int32(size)
		err = queries.UpdatePreferenceForLister(r.Context(), db.UpdatePreferenceForListerParams{
			LanguageID: pref.LanguageID,
			ListerID:   uid,
			PageSize:   pref.PageSize,
			Timezone:   pref.Timezone,
		})
	}
	if err != nil {
		log.Printf("save paging: %v", err)
		return fmt.Errorf("save paging fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/paging"}
}
