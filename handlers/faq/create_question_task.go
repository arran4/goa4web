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

// CreateQuestionTask creates a new FAQ entry.
type CreateQuestionTask struct{ tasks.TaskString }

var createQuestionTask = &CreateQuestionTask{TaskString: TaskCreate}
var _ tasks.Task = (*CreateQuestionTask)(nil)

func (CreateQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreate)(r, m)
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

	res, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO faq (question, answer, faqCategories_idfaqCategories, users_idusers, language_idlanguage) VALUES (?, ?, ?, ?, ?)",
		sql.NullString{String: question, Valid: true},
		sql.NullString{String: answer, Valid: true},
		int32(category), uid, 1,
	)
	if err != nil {
		return fmt.Errorf("insert faq fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := res.LastInsertId()
	_ = queries.InsertFAQRevision(r.Context(), db.InsertFAQRevisionParams{
		FaqID:        int32(id),
		UsersIdusers: uid,
		Question:     sql.NullString{String: question, Valid: true},
		Answer:       sql.NullString{String: answer, Valid: true},
	})

	return nil
}
