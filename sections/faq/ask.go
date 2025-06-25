package faq

import (
	"context"
	"database/sql"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func AskPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		SelectedLanguageId: resolveDefaultLanguageID(r.Context(), queries),
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomFAQIndex(data.CoreData)

	if err := templates.RenderTemplate(w, "askPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AskActionPage(w http.ResponseWriter, r *http.Request) {
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

	if err := queries.CreateFAQQuestion(r.Context(), db.CreateFAQQuestionParams{
		Question: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(languageId),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	// TODO notify admin

	http.Redirect(w, r, "/faq", http.StatusTemporaryRedirect)
}

// resolveDefaultLanguageID converts the configured language name to its ID.
func resolveDefaultLanguageID(ctx context.Context, q *db.Queries) int32 {
	if runtimeconfig.AppRuntimeConfig.DefaultLanguage == "" {
		return 0
	}
	id, err := q.GetLanguageIDByName(ctx, sql.NullString{String: runtimeconfig.AppRuntimeConfig.DefaultLanguage, Valid: true})
	if err != nil {
		return 0
	}
	return id
}
