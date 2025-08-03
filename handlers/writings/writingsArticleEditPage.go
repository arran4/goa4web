package writings

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/searchworker"
	"strings"

	"github.com/arran4/goa4web/core"
)

func ArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Writing            *db.GetWritingForListerByIDRow
		UserId             int32
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}
	data.CoreData.PageTitle = "Edit Article"

	// article ID is validated by the RequireWritingAuthor middleware, so we
	// no longer need to parse it here.

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	writing, err := cd.CurrentWriting()
	if err != nil || writing == nil {
		log.Printf("current writing: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.CurrentWriting()
	if err != nil || writing == nil {
		log.Printf("current writing: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")

	queries := cd.Queries()

	if err := queries.UpdateWritingForWriter(r.Context(), db.UpdateWritingForWriterParams{
		Title:      sql.NullString{Valid: true, String: title},
		Abstract:   sql.NullString{Valid: true, String: abstract},
		Content:    sql.NullString{Valid: true, String: body},
		Private:    sql.NullBool{Valid: true, Bool: private},
		LanguageID: int32(languageId),
		WritingID:  writing.Idwriting,
		WriterID:   cd.UserID,
		GranteeID:  sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	}); err != nil {
		log.Printf("updateWriting Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			author := ""
			if writing.Writerusername.Valid {
				author = writing.Writerusername.String
			}
			evt.Data["Title"] = title
			evt.Data["Author"] = author
			evt.Data["PostURL"] = cd.AbsoluteURL(fmt.Sprintf("/writings/article/%d", writing.Idwriting))
			evt.Data["target"] = notif.Target{Type: "writing", ID: writing.Idwriting}
		}
	}

	if err := queries.SystemDeleteWritingSearchByWritingID(r.Context(), writing.Idwriting); err != nil {
		log.Printf("writingSearchDelete Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fullText := strings.Join([]string{abstract, title, body}, " ")
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeWriting, ID: writing.Idwriting, Text: fullText}
		}
	}
}
