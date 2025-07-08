package writings

import (
	"database/sql"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchutil"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/runtimeconfig"
)

func ArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Writing            *db.GetWritingByIdForUserDescendingByPublishedDateRow
		UserId             int32
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData),
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, runtimeconfig.AppRuntimeConfig.DefaultLanguage)),
	}

	// article ID is validated by the RequireWritingAuthor middleware, so we
	// no longer need to parse it here.

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	writing := r.Context().Value(hcommon.KeyWriting).(*db.GetWritingByIdForUserDescendingByPublishedDateRow)
	data.Writing = writing

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomWritingsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "articleEditPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ArticleEditActionPage(w http.ResponseWriter, r *http.Request) {
	// RequireWritingAuthor middleware loads the writing and validates access.
	writing := r.Context().Value(hcommon.KeyWriting).(*db.GetWritingByIdForUserDescendingByPublishedDateRow)

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	if err := queries.UpdateWriting(r.Context(), db.UpdateWritingParams{
		Title:              sql.NullString{Valid: true, String: title},
		Abstract:           sql.NullString{Valid: true, String: abstract},
		Writting:           sql.NullString{Valid: true, String: body},
		Private:            sql.NullBool{Valid: true, Bool: private},
		LanguageIdlanguage: int32(languageId),
		Idwriting:          writing.Idwriting,
	}); err != nil {
		log.Printf("updateWriting Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.WritingSearchDelete(r.Context(), writing.Idwriting); err != nil {
		log.Printf("writingSearchDelete Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, text := range []string{
		abstract,
		title,
		body,
	} {
		wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
		if done {
			return
		}

		if searchutil.InsertWordsToWritingSearch(w, r, wordIds, queries, int64(writing.Idwriting)) {
			return
		}
	}
}
