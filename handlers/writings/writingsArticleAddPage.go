package writings

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	searchutil "github.com/arran4/goa4web/internal/utils/searchutil"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func ArticleAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Languages []*db.Language
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData),
	}

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	common.TemplateHandler(w, r, "articleAddPage.gohtml", data)
}
func ArticleAddActionPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	articleId, err := queries.InsertWriting(r.Context(), db.InsertWritingParams{
		WritingCategoryID:  int32(categoryId),
		Title:              sql.NullString{Valid: true, String: title},
		Abstract:           sql.NullString{Valid: true, String: abstract},
		Writing:            sql.NullString{Valid: true, String: body},
		Private:            sql.NullBool{Valid: true, Bool: private},
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
	})
	if err != nil {
		log.Printf("insertWriting Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var author string
	if u, err := queries.GetUserById(r.Context(), uid); err == nil {
		author = u.Username.String
	}
	if cd, ok := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["writing"] = notifications.WritingInfo{Title: title, Author: author}
		}
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

		if searchutil.InsertWordsToWritingSearch(w, r, wordIds, queries, articleId) {
			return
		}
	}
}
