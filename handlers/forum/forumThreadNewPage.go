package forum

import (
	"database/sql"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	blogs "github.com/arran4/goa4web/handlers/blogs"
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchutil"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func ThreadNewPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages          []*db.Language
		SelectedLanguageId int
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(hcommon.KeyCoreData).(*CoreData),
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries)),
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	blogs.CustomBlogIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "threadNewPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ThreadNewActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	topicId, err := strconv.Atoi(vars["topic"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	// TODO check if the user has the right right to topic

	threadId, err := queries.MakeThread(r.Context(), int32(topicId))
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d", topicId, threadId)

	provider := getEmailProvider()

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: int32(threadId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := PostUpdate(r.Context(), queries, int32(threadId), int32(topicId)); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchutil.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	notifyThreadSubscribers(r.Context(), provider, queries, int32(threadId), uid, endUrl)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func ThreadNewCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	endUrl := fmt.Sprintf("/forum/topic/%d", topicId)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
