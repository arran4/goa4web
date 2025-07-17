package blogs

import (
	"database/sql"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/config"
)

func BlogEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
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
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
		Mode:               "Edit",
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	row := r.Context().Value(common.KeyBlogEntry).(*db.GetBlogEntryForUserByIdRow)
	data.Blog = row

	handlers.TemplateHandler(w, r, "blogEditPage.gohtml", data)
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
