package user

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/runtimeconfig"
)

func userPagingPage(w http.ResponseWriter, r *http.Request) {
	pref, _ := r.Context().Value(common.KeyPreference).(*db.Preference)
	size := runtimeconfig.AppRuntimeConfig.PageSizeDefault
	if pref != nil {
		size = int(pref.PageSize)
	}
	data := struct {
		*common.CoreData
		Size int
		Min  int
		Max  int
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Size:     size,
		Min:      runtimeconfig.AppRuntimeConfig.PageSizeMin,
		Max:      runtimeconfig.AppRuntimeConfig.PageSizeMax,
	}
	if err := templates.RenderTemplate(w, "pagingPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userPagingSaveActionPage(w http.ResponseWriter, r *http.Request) {
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
	if size < runtimeconfig.AppRuntimeConfig.PageSizeMin {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMin
	}
	if size > runtimeconfig.AppRuntimeConfig.PageSizeMax {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMax
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	pref, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertPreference(r.Context(), db.InsertPreferenceParams{
				LanguageIdlanguage: 0,
				UsersIdusers:       uid,
				PageSize:           int32(size),
			})
		}
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
