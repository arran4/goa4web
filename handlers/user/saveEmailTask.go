package user

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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
		return nil
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	updates := r.PostFormValue("emailupdates") != ""
	auto := r.PostFormValue("autosubscribe") != ""

	_, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("preference load: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertEmailPreference(r.Context(), db.InsertEmailPreferenceParams{
				Emailforumupdates:    sql.NullBool{Bool: updates, Valid: true},
				AutoSubscribeReplies: auto,
				UsersIdusers:         uid,
			})
		}
	} else {
		err = queries.UpdateEmailForumUpdatesByUserID(r.Context(), db.UpdateEmailForumUpdatesByUserIDParams{
			Emailforumupdates: sql.NullBool{Bool: updates, Valid: true},
			UsersIdusers:      uid,
		})
		if err == nil {
			err = queries.UpdateAutoSubscribeRepliesByUserID(r.Context(), db.UpdateAutoSubscribeRepliesByUserIDParams{
				AutoSubscribeReplies: auto,
				UsersIdusers:         uid,
			})
		}
	}
	if err != nil {
		log.Printf("save email pref: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}

	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
	return nil
}
