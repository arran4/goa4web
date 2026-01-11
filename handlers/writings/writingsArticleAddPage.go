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
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func ArticleAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Languages []*db.Language
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Add Article"
	data := Data{}

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	WritingsArticleAddPageTmpl.Handle(w, r, data)
}

const WritingsArticleAddPageTmpl handlers.Page = "writings/articleAddPage.gohtml"

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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	articleId, err := cd.CreateWriting(int32(categoryId), int32(languageId), title, abstract, body, private)
	if err != nil {
		log.Printf("insertWriting Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	if articleId == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	var author string
	queries := cd.Queries()
	if u, err := queries.SystemGetUserByID(r.Context(), uid); err == nil {
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
