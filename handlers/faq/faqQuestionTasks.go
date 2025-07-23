package faq

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// EditQuestionTask modifies an existing FAQ entry.
type EditQuestionTask struct{ tasks.TaskString }

// DeleteQuestionTask removes a question from the FAQ.
type DeleteQuestionTask struct{ tasks.TaskString }

// CreateQuestionTask creates a new FAQ entry.
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

func (DeleteQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		return fmt.Errorf("faq id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.DeleteFAQ(r.Context(), int32(faq)); err != nil {
		return fmt.Errorf("delete faq fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}

func (EditQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		return fmt.Errorf("category parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		return fmt.Errorf("faq id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
	}); err != nil {
		return fmt.Errorf("update faq question fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}

func (CreateQuestionTask) Action(w http.ResponseWriter, r *http.Request) any {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		return fmt.Errorf("category parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	if _, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO faq (question, answer, faqCategories_idfaqCategories, users_idusers, language_idlanguage) VALUES (?, ?, ?, ?, ?)",
		sql.NullString{String: question, Valid: true},
		sql.NullString{String: answer, Valid: true},
		int32(category), uid, 1,
	); err != nil {
		return fmt.Errorf("insert faq fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
