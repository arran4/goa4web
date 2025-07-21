package languages

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminLanguageRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/languages", http.StatusMovedPermanently)
}

func adminLanguagesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows []*db.Language
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "languagesPage.gohtml", data)
}

func adminLanguagesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid := r.PostFormValue("cid")
	cname := r.PostFormValue("cname")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.RenameLanguage(r.Context(), db.RenameLanguageParams{
		Nameof:     sql.NullString{Valid: true, String: cname},
		Idlanguage: int32(cidi),
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RenameLanguage: %w", err).Error())
	} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["LanguageID"] = cidi
			evt.Data["LanguageName"] = cname
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
func adminLanguagesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid := r.PostFormValue("cid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		var name string
		if rows, err := queries.FetchLanguages(r.Context()); err == nil {
			for _, l := range rows {
				if l.Idlanguage == int32(cidi) {
					name = l.Nameof.String
					break
				}
			}
		}
		if err := queries.DeleteLanguage(r.Context(), int32(cidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("DeleteLanguage: %w", err).Error())
		} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["LanguageID"] = cidi
				evt.Data["LanguageName"] = name
			}
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
func adminLanguagesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cname := r.PostFormValue("cname")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/languages",
	}
	if res, err := queries.InsertLanguage(r.Context(), sql.NullString{
		String: cname,
		Valid:  true,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	} else if id, err := res.LastInsertId(); err == nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["LanguageID"] = id
				evt.Data["LanguageName"] = cname
			}
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
