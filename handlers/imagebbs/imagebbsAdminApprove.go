package imagebbs

import (
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ApprovePostTask marks a post as approved.
type ApprovePostTask struct{ tasks.TaskString }

var approvePostTask = &ApprovePostTask{TaskString: TaskApprove}

func (ApprovePostTask) Action(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := queries.ApproveImagePost(r.Context(), int32(pid)); err != nil {
		log.Printf("ApproveImagePost error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
