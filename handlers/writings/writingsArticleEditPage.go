package writings

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	searchworker "github.com/arran4/goa4web/workers/searchworker"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

func ArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Writing            *db.GetWritingByIdForUserDescendingByPublishedDateRow
		UserId             int32
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
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

	writing := r.Context().Value(common.KeyWriting).(*db.GetWritingByIdForUserDescendingByPublishedDateRow)
	data.Writing = writing

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "articleEditPage.gohtml", data)
}

func ArticleEditActionPage(w http.ResponseWriter, r *http.Request) {
	// RequireWritingAuthor middleware loads the writing and validates access.
	writing := r.Context().Value(common.KeyWriting).(*db.GetWritingByIdForUserDescendingByPublishedDateRow)

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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

	fullText := strings.Join([]string{abstract, title, body}, " ")
	if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeWriting, ID: writing.Idwriting, Text: fullText}
		}
	}
}
