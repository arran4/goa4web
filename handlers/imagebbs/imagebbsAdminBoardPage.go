package imagebbs

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ModifyBoardTask updates an existing board's settings.
type ModifyBoardTask struct{ tasks.TaskString }

var modifyBoardTask = &ModifyBoardTask{TaskString: TaskModifyBoard}

var _ tasks.Task = (*ModifyBoardTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*ModifyBoardTask)(nil)

func (ModifyBoardTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("imageBoardUpdateEmail")
}

func (ModifyBoardTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("imageBoardUpdateEmail")
	return &v
}

func (ModifyBoardTask) Action(w http.ResponseWriter, r *http.Request) any {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["board"])

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	err := queries.UpdateImageBoard(r.Context(), db.UpdateImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
		Idimageboard:           int32(bid),
	})
	if err != nil {
		return fmt.Errorf("update image board fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler("/admin/imagebbs/boards")
}
