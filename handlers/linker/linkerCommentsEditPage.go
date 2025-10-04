package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// CommentEditActionCancelPage aborts editing a comment.
func CommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	handlers.RedirectToGet(w, r, fmt.Sprintf("/linker/comments/%d", linkId))
}
