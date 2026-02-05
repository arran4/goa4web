package writings

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
)

func ArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Languages          []*db.Language
		SelectedLanguageId int
		Writing            *db.GetWritingForListerByIDRow
		UserId             int32
		AuthorLabels       []string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}
	cd.PageTitle = "Edit Article"

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	writing, err := cd.EditableArticle()
	if err != nil || writing == nil {
		log.Printf("current writing: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data.Writing = writing

	if als, err := cd.WritingAuthorLabels(writing.Idwriting); err == nil {
		data.AuthorLabels = als
	}

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data.Languages = languageRows

	WritingsArticleEditPageTmpl.Handle(w, r, data)
}

const WritingsArticleEditPageTmpl tasks.Template = "writings/articleEditPage.gohtml"

func ArticleEditActionPage(w http.ResponseWriter, r *http.Request) {
	// RequireWritingAuthor middleware loads the writing and validates access.
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.EditableArticle()
	if err != nil || writing == nil {
		log.Printf("current writing: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")

	queries := cd.Queries()

	if err := cd.UpdateWriting(writing, title, abstract, body, private, int32(languageId)); err != nil {
		log.Printf("updateWriting Error: %s", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
