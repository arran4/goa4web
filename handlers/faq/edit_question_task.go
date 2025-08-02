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

var editQuestionTask = &EditQuestionTask{TaskString: TaskEdit}
var _ tasks.Task = (*EditQuestionTask)(nil)

func (EditQuestionTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskEdit)(r, m)
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
		ViewerID:                     cd.UserID,
	}); err != nil {
		return fmt.Errorf("update faq question fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	_ = queries.InsertFAQRevisionForUser(r.Context(), db.InsertFAQRevisionForUserParams{
		FaqID:        int32(faq),
		UsersIdusers: uid,
		Question:     sql.NullString{String: question, Valid: true},
		Answer:       sql.NullString{String: answer, Valid: true},
		UserID:       sql.NullInt32{Int32: uid, Valid: true},
		ViewerID:     uid,
	})

	return nil
}
