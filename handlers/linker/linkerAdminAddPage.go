package linker

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func AdminAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Categories         []*db.LinkerCategory
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
	}

	categoryRows, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomLinkerIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminAddPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func AdminAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	if err := queries.CreateLinkerItem(r.Context(), db.CreateLinkerItemParams{
		UsersIdusers:     uid,
		LinkerCategoryID: int32(category),
		Title:            sql.NullString{Valid: true, String: title},
		Url:              sql.NullString{Valid: true, String: url},
		Description:      sql.NullString{Valid: true, String: description},
	}); err != nil {
		log.Printf("createLinkerItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)

}
