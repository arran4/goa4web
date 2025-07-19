package languages

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
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
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)

	rows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "languagesPage.gohtml", data)
}

func adminLanguagesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cid := r.PostFormValue("cid")
	cname := r.PostFormValue("cname")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.RenameLanguage(r.Context(), db.RenameLanguageParams{
		Nameof:     sql.NullString{Valid: true, String: cname},
		Idlanguage: int32(cidi),
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RenameLanguage: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
func adminLanguagesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cid := r.PostFormValue("cid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteLanguage(r.Context(), int32(cidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLanguage: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
func adminLanguagesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cname := r.PostFormValue("cname")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin/languages",
	}
	if err := queries.CreateLanguage(r.Context(), sql.NullString{
		String: cname,
		Valid:  true,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
