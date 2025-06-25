package goa4web

import (
	"github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func imagebbsAdminApprovePostPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := queries.ApproveImagePost(r.Context(), int32(pid)); err != nil {
		log.Printf("ApproveImagePost error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
