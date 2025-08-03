package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveEmailTask updates email notification preferences.
type SaveEmailTask struct{ tasks.TaskString }

var saveEmailTask = &SaveEmailTask{TaskString: tasks.TaskString(TaskSaveAll)}

var _ tasks.Task = (*SaveEmailTask)(nil)

func (SaveEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		return common.UserError{ErrorMessage: "forbidden"}
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	updates := r.PostFormValue("emailupdates") != ""
	auto := r.PostFormValue("autosubscribe") != ""

	_, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("preference load: %v", err)
		return fmt.Errorf("preference load fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertEmailPreferenceForLister(r.Context(), db.InsertEmailPreferenceForListerParams{
				EmailForumUpdates:    sql.NullBool{Bool: updates, Valid: true},
				AutoSubscribeReplies: auto,
				ListerID:             uid,
			})
		}
	} else {
		err = queries.UpdateEmailForumUpdatesForLister(r.Context(), db.UpdateEmailForumUpdatesForListerParams{
			EmailForumUpdates: sql.NullBool{Bool: updates, Valid: true},
			ListerID:          uid,
		})
		if err == nil {
			err = queries.UpdateAutoSubscribeRepliesForLister(r.Context(), db.UpdateAutoSubscribeRepliesForListerParams{
				AutoSubscribeReplies: auto,
				ListerID:             uid,
			})
		}
	}
	if err != nil {
		log.Printf("save email pref: %v", err)
		return fmt.Errorf("save email pref fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}
