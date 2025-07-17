package forum

import (
	"database/sql"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"
	"strconv"

	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	blogs "github.com/arran4/goa4web/handlers/blogs"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	searchutil "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/email"
	"github.com/gorilla/mux"
)

// CreateThreadTask handles creating a new forum thread.
type CreateThreadTask struct{ tasks.TaskString }

var createThreadTask = &CreateThreadTask{TaskString: TaskCreateThread}

func (CreateThreadTask) Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages          []*db.Language
		SelectedLanguageId int
	}

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	cd := r.Context().Value(handlers.KeyCoreData).(*CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	blogs.CustomBlogIndex(data.CoreData, r)

	handlers.TemplateHandler(w, r, "threadNewPage.gohtml", data)
}

func (CreateThreadTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	topicId, err := strconv.Atoi(vars["topic"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	allowed, err := UserCanCreateThread(r.Context(), queries, int32(topicId), uid)
	if err != nil {
		log.Printf("UserCanCreateThread error: %v", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	threadId, err := queries.MakeThread(r.Context(), int32(topicId))
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	var topicTitle, author string
	if trow, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{ViewerID: uid, Idforumtopic: int32(topicId), ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0}}); err == nil {
		topicTitle = trow.Title.String
	}
	if u, err := queries.GetUserById(r.Context(), uid); err == nil {
		author = u.Username.String
	}
	if cd, ok := r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["thread"] = notif.ThreadInfo{TopicTitle: topicTitle, Author: author}
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d", topicId, threadId)

	provider := email.ProviderFromConfig(config.AppRuntimeConfig)

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      int32(threadId),
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

	// TODO move to searchworker that is automatically activated by a event.
	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchutil.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func ThreadNewCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	endUrl := fmt.Sprintf("/forum/topic/%d", topicId)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
