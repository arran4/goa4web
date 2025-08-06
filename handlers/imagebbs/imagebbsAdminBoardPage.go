package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))
	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)

	queries := cd.Queries()

	boards, err := queries.AdminListBoards(r.Context(), db.AdminListBoardsParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("fetch boards %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	parents := make(map[int32]int32, len(boards))
	for _, b := range boards {
		parents[b.Idimageboard] = b.ImageboardIdimageboard
	}
	if path, loop := algorithms.WouldCreateLoop(parents, int32(bid), int32(parentBoardId)); loop {
		return common.UserError{ErrorMessage: fmt.Sprintf("invalid parent board: loop %v", path)}
	}

	err = queries.AdminUpdateImageBoard(r.Context(), db.AdminUpdateImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
		Idimageboard:           int32(bid),
	})
	if err != nil {
		return fmt.Errorf("update image board fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/imagebbs/boards"}
}

// AdminBoardPage shows a form to edit an existing board.
func AdminBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Board  *db.Imageboard
		Boards []*db.Imageboard
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Edit Image Board"

	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)
	if bid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	boards, err := cd.ImageBoards()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("imageBoards error: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	var board *db.Imageboard
	for _, b := range boards {
		if int(b.Idimageboard) == bid {
			board = b
			break
		}
	}
	if board == nil {
		http.NotFound(w, r)
		return
	}

	data := Data{Board: board, Boards: boards}

	handlers.TemplateHandler(w, r, "adminBoardPage.gohtml", data)
}
