package faq

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AskTask struct{ tasks.TaskString }

var askTask = &AskTask{TaskString: TaskAsk}

func (AskTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskAsk)(r, m)
}

func (AskTask) Page(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			r.PostForm.Del("task")
		}
		askTask.Action(w, r)
		return
	}

	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "askPage.gohtml", data)
}

func (AskTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := handlers.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if err := queries.CreateFAQQuestion(r.Context(), db.CreateFAQQuestionParams{
		Question: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(languageId),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	evt := cd.Event()
	evt.Path = "/admin/faq"
	if evt.Data == nil {
		evt.Data = map[string]any{}
	}
	evt.Data["question"] = text

	// The BusWorker sends notifications based on event metadata.
	// Setting Admin=true signals administrators should be alerted.

	http.Redirect(w, r, "/faq", http.StatusTemporaryRedirect)
}
