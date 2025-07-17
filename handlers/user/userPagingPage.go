package user

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	handlers "github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/tasks"
)

type PagingSaveTask struct{ tasks.TaskString }

var pagingSaveTask = &PagingSaveTask{TaskString: tasks.TaskString(TaskSaveAll)}

func userPagingPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData)
	pref, _ := cd.Preference()
	size := config.AppRuntimeConfig.PageSizeDefault
	if pref != nil {
		size = int(pref.PageSize)
	}
	data := struct {
		*handlers.CoreData
		Size int
		Min  int
		Max  int
	}{
		CoreData: cd,
		Size:     size,
		Min:      config.AppRuntimeConfig.PageSizeMin,
		Max:      config.AppRuntimeConfig.PageSizeMax,
	}
	handlers.TemplateHandler(w, r, "pagingPage.gohtml", data)
}

func (PagingSaveTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/paging", http.StatusSeeOther)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	size, _ := strconv.Atoi(r.FormValue("size"))
	if size < config.AppRuntimeConfig.PageSizeMin {
		size = config.AppRuntimeConfig.PageSizeMin
	}
	if size > config.AppRuntimeConfig.PageSizeMax {
		size = config.AppRuntimeConfig.PageSizeMax
	}
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	cd := r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData)

	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("preference load: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		err = queries.InsertPreference(r.Context(), db.InsertPreferenceParams{
			LanguageIdlanguage: 0,
			UsersIdusers:       uid,
			PageSize:           int32(size),
		})
	} else {
		pref.PageSize = int32(size)
		err = queries.UpdatePreference(r.Context(), db.UpdatePreferenceParams{
			LanguageIdlanguage: pref.LanguageIdlanguage,
			UsersIdusers:       uid,
			PageSize:           pref.PageSize,
		})
	}
	if err != nil {
		log.Printf("save paging: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/usr/paging", http.StatusSeeOther)
}
