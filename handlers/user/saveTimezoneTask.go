package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveTimezoneTask updates the user's timezone preference.
type SaveTimezoneTask struct{ tasks.TaskString }

var saveTimezoneTask = &SaveTimezoneTask{TaskString: TaskSaveTimezone}

var _ tasks.Task = (*SaveTimezoneTask)(nil)

func (SaveTimezoneTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %v", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	tz := strings.TrimSpace(r.PostFormValue("timezone"))
	if tz != "" {
		if _, err := time.LoadLocation(tz); err != nil {
			return common.UserError{ErrorMessage: "invalid timezone"}
		}
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := queries.UpdateTimezoneForLister(r.Context(), db.UpdateTimezoneForListerParams{Timezone: sql.NullString{String: tz, Valid: tz != ""}, ListerID: uid}); err != nil {
		log.Printf("Save timezone Error: %v", err)
		return fmt.Errorf("save timezone fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
