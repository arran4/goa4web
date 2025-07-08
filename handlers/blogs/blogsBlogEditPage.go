package blogs

import (
	"database/sql"
	"fmt"
	db "github.com/arran4/goa4web/internal/db"

	corelanguage "github.com/arran4/goa4web/core/language"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/runtimeconfig"
)

func BlogEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	if !(cd.HasRole("writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*CoreData
		Languages          []*db.Language
		Blog               *db.GetBlogEntryForUserByIdRow
		SelectedLanguageId int
		Mode               string
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, runtimeconfig.AppRuntimeConfig.DefaultLanguage)),
		Mode:               "Edit",
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	row := r.Context().Value(common.KeyBlogEntry).(*db.GetBlogEntryForUserByIdRow)
	data.Blog = row

	CustomBlogIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "blogEditPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func BlogEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	row := r.Context().Value(common.KeyBlogEntry).(*db.GetBlogEntryForUserByIdRow)

	err = queries.UpdateBlogEntry(r.Context(), db.UpdateBlogEntryParams{
		Idblogs:            row.Idblogs,
		LanguageIdlanguage: int32(languageId),
		Blog: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", row.Idblogs), http.StatusTemporaryRedirect)
}
