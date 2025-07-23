package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CommentEditActionCancelPage aborts editing a comment.
func CommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	http.Redirect(w, r, fmt.Sprintf("/linker/comments/%d", linkId), http.StatusTemporaryRedirect)
}
