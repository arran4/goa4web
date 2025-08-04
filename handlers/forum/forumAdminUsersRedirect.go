package forum

import (
	"net/http"

	"github.com/gorilla/mux"
)

// AdminUsersRedirect redirects forum user management to the global user admin page.
func AdminUsersRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/users", http.StatusTemporaryRedirect)
}

// AdminUserLevelsRedirect redirects to the global user profile page.
func AdminUserLevelsRedirect(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["user"]
	http.Redirect(w, r, "/admin/user/"+id, http.StatusTemporaryRedirect)
}
