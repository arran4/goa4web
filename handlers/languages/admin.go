package languages

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminLanguageRedirect sends users to the language list.
func adminLanguageRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/languages", http.StatusMovedPermanently)
}

// adminLanguagesPage shows the list of available languages.
func adminLanguagesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Languages"
	AdminLanguagesPage.Handle(w, r, struct{}{})
}

const AdminLanguagesPage handlers.Page = "admin/languagesPage.gohtml"

// adminLanguagePage displays statistics for a specific language.
func adminLanguagePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	id, err := strconv.Atoi(mux.Vars(r)["language"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Bad Request"))
		return
	}
	var lang *db.Language
	if rows, err := cd.Languages(); err == nil {
		for _, l := range rows {
			if l.ID == int32(id) {
				lang = l
				break
			}
		}
	}
	if lang == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Not Found"))
		return
	}
	counts, err := cd.Queries().AdminLanguageUsageCounts(r.Context(), db.AdminLanguageUsageCountsParams{LangID: sql.NullInt32{Int32: int32(id), Valid: true}})
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	cd.PageTitle = "Language"
	data := struct {
		Language *db.Language
		Counts   *db.AdminLanguageUsageCountsRow
	}{
		Language: lang,
		Counts:   counts,
	}
	AdminLanguagePage.Handle(w, r, data)
}

const AdminLanguagePage handlers.Page = "admin/languagePage.gohtml"

// adminLanguageEditPage shows forms to rename or delete a language.
func adminLanguageEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	id, err := strconv.Atoi(mux.Vars(r)["language"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Bad Request"))
		return
	}
	var lang *db.Language
	if rows, err := cd.Languages(); err == nil {
		for _, l := range rows {
			if l.ID == int32(id) {
				lang = l
				break
			}
		}
	}
	if lang == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Not Found"))
		return
	}
	cd.PageTitle = "Edit Language"
	data := struct{ Language *db.Language }{Language: lang}
	AdminLanguageEditPage.Handle(w, r, data)
}

const AdminLanguageEditPage handlers.Page = "admin/languageEditPage.gohtml"

// adminLanguageNewPage shows the form to create a new language.
func adminLanguageNewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "New Language"
	AdminLanguageNewPage.Handle(w, r, struct{}{})
}

const AdminLanguageNewPage handlers.Page = "admin/languageNewPage.gohtml"
