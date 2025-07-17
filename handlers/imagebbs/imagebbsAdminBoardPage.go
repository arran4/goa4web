package imagebbs

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ModifyBoardTask updates an existing board's settings.
type ModifyBoardTask struct{ tasks.TaskString }

var modifyBoardTask = &ModifyBoardTask{TaskString: TaskModifyBoard}

func (ModifyBoardTask) Action(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["board"])

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	err := queries.UpdateImageBoard(r.Context(), db.UpdateImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
		Idimageboard:           int32(bid),
	})
	if err != nil {
		log.Printf("Error: createImageBoard: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/admin/imagebbs/boards", http.StatusTemporaryRedirect)
}
