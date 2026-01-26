package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// NewBoardTask creates a new image board.
type NewBoardTask struct{ tasks.TaskString }

var newBoardTask = &NewBoardTask{TaskString: TaskNewBoard}

var _ tasks.Task = (*NewBoardTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*NewBoardTask)(nil)
var _ tasks.EmailTemplatesRequired = (*NewBoardTask)(nil)

func (NewBoardTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationImageBoardNew.EmailTemplates(), true
}

func (NewBoardTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationImageBoardNew.NotificationTemplate()
	return &v
}

func (NewBoardTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationImageBoardNew.RequiredTemplates()
}

func AdminNewBoardPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/imagebbs/boards", http.StatusTemporaryRedirect)
}

func (NewBoardTask) Action(w http.ResponseWriter, r *http.Request) any {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	boards, err := queries.AdminListBoards(r.Context(), db.AdminListBoardsParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("fetch boards %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	parents := make(map[int32]int32, len(boards))
	for _, b := range boards {
		if b.ImageboardIdimageboard.Valid {
			parents[b.Idimageboard] = b.ImageboardIdimageboard.Int32
		} else {
			parents[b.Idimageboard] = 0
		}
	}
	if path, loop := algorithms.WouldCreateLoop(parents, 0, int32(parentBoardId)); loop {
		return common.UserError{ErrorMessage: fmt.Sprintf("invalid parent board: loop %v", path)}
	}

	err = queries.AdminCreateImageBoard(r.Context(), db.AdminCreateImageBoardParams{
		ImageboardIdimageboard: sql.NullInt32{Int32: int32(parentBoardId), Valid: parentBoardId != 0},
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
	})
	if err != nil {
		return fmt.Errorf("create image board fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/imagebbs/boards"}
}
