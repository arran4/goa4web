package blogs

import (
	"database/sql"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"

	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

func BlogAddPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Mode               string
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
		Mode:               "Add",
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	common.TemplateHandler(w, r, "blogAddPage.gohtml", data)
}

func BlogAddActionPage(w http.ResponseWriter, r *http.Request) {
	if err := common.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	id, err := queries.CreateBlogEntry(r.Context(), db.CreateBlogEntryParams{
		UsersIdusers:       uid,
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

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", id), http.StatusTemporaryRedirect)
}
