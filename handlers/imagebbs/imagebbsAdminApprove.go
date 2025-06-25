package imagebbs

import (
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func AdminApprovePostPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := queries.ApproveImagePost(r.Context(), int32(pid)); err != nil {
		log.Printf("ApproveImagePost error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
