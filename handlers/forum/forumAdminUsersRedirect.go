package forum

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// AdminUsersRedirect redirects forum user management to the global user admin page.
func AdminUsersRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/users", http.StatusTemporaryRedirect)
}

// AdminUserLevelsRedirect redirects to the global user profile page.
func AdminUserLevelsRedirect(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cpu := cd.CurrentProfileUser()
	http.Redirect(w, r, "/admin/user/"+strconv.Itoa(int(cpu.Idusers)), http.StatusTemporaryRedirect)
}
