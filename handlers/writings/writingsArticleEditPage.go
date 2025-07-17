package writings

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	searchworker "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
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
	cd := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
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

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	common.TemplateHandler(w, r, "articleEditPage.gohtml", data)
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
		Writing:            sql.NullString{Valid: true, String: body},
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
		wordIds, done := searchworker.SearchWordIdsFromText(w, r, text, queries)
		if done {
			return
		}

		if searchworker.InsertWordsToWritingSearch(w, r, wordIds, queries, int64(writing.Idwriting)) {
			return
		}
	}
}
