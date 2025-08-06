package imagebbs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/gorilla/mux"
)

// DeleteBoardTask removes an image board.
type DeleteBoardTask struct{ tasks.TaskString }

var deleteBoardTask = &DeleteBoardTask{TaskString: TaskDeleteBoard}

var _ tasks.Task = (*DeleteBoardTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*DeleteBoardTask)(nil)

func (DeleteBoardTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("imageBoardDeleteEmail")
}

func (DeleteBoardTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("imageBoardDeleteEmail")
	return &v
}

func (DeleteBoardTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)
	if bid == 0 {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid board"))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := queries.AdminDeleteImageBoard(r.Context(), int32(bid)); err != nil {
		return fmt.Errorf("delete image board fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/imagebbs/boards"}
}
