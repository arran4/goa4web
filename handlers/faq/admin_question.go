package faq

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type EditQuestionTask struct{ tasks.TaskString }
type DeleteQuestionTask struct{ tasks.TaskString }
type CreateQuestionTask struct{ tasks.TaskString }

var editQuestionTask = &EditQuestionTask{TaskString: TaskEdit}
var _ tasks.Task = (*EditQuestionTask)(nil)
var deleteQuestionTask = &DeleteQuestionTask{TaskString: TaskRemoveRemove}
var _ tasks.Task = (*DeleteQuestionTask)(nil)
var createQuestionTask = &CreateQuestionTask{TaskString: TaskCreate}
var _ tasks.Task = (*CreateQuestionTask)(nil)

func (EditQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskEdit)(r, m)
}

func (DeleteQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRemoveRemove)(r, m)
}

func (CreateQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreate)(r, m)
}

func AdminQuestionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories []*db.FaqCategory
		Rows       []*db.Faq
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	catrows, err := queries.GetAllFAQCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Categories = catrows

	rows, err := queries.GetAllFAQQuestions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "adminQuestionPage.gohtml", data)
}

func (DeleteQuestionTask) Action(w http.ResponseWriter, r *http.Request) {
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.DeleteFAQ(r.Context(), int32(faq)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (EditQuestionTask) Action(w http.ResponseWriter, r *http.Request) {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (CreateQuestionTask) Action(w http.ResponseWriter, r *http.Request) {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if _, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO faq (question, answer, faqCategories_idfaqCategories, users_idusers, language_idlanguage) VALUES (?, ?, ?, ?, ?)",
		sql.NullString{String: question, Valid: true},
		sql.NullString{String: answer, Valid: true},
		int32(category), uid, 1,
	); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
