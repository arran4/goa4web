package imagebbs

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// NewBoardTask creates a new image board.
type NewBoardTask struct{ tasks.TaskString }

var newBoardTask = &NewBoardTask{TaskString: TaskNewBoard}

var _ tasks.Task = (*NewBoardTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*NewBoardTask)(nil)

func (NewBoardTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationImageBoardNewEmail")
}

func (NewBoardTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationImageBoardNewEmail")
	return &v
}

func AdminNewBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Boards []*db.Imageboard
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	boardRows, err := data.CoreData.ImageBoards()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllImageBoards Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Boards = boardRows

	handlers.TemplateHandler(w, r, "adminNewBoardPage.gohtml", data)
}

func (NewBoardTask) Action(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	err := queries.CreateImageBoard(r.Context(), db.CreateImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
	})
	if err != nil {
		log.Printf("Error: createImageBoard: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/admin/imagebbs/boards", http.StatusTemporaryRedirect)
}
