package writings

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/searchworker"
	"strings"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func ArticleAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages []*db.Language
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.CoreData.PageTitle = "Add Article"

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "articleAddPage.gohtml", data)
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

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

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
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Title"] = title
			evt.Data["Author"] = author
			evt.Data["target"] = notifications.Target{Type: "writing", ID: int32(articleId)}
		}
	}

	fullText := strings.Join([]string{abstract, title, body}, " ")
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeWriting, ID: int32(articleId), Text: fullText}
		}
	}
}
