package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

const AdminUserProfilePageTmpl tasks.Template = "admin/adminUserPage.gohtml"

type AdminUserProfilePage struct {
	UserID   int32
	UserName string
	Data     any
}

func (p *AdminUserProfilePage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	title := fmt.Sprintf("User %d", p.UserID)
	if p.UserName != "" {
		title = fmt.Sprintf("User %s", p.UserName)
	}
	return title, "", &AdminUserListPage{}
}

func (p *AdminUserProfilePage) PageTitle() string {
	if p.UserName != "" {
		return fmt.Sprintf("User %s", p.UserName)
	}
	return fmt.Sprintf("User %d", p.UserID)
}

func (p *AdminUserProfilePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminUserProfilePageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

var _ common.Page = (*AdminUserProfilePage)(nil)
var _ http.Handler = (*AdminUserProfilePage)(nil)

type AdminUserProfileTask struct{}

func (t *AdminUserProfileTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		return handlers.ErrNotFound
	}
	return &AdminUserProfilePage{
		UserID:   user.Idusers,
		UserName: user.Username.String,
		Data:     struct{}{},
	}
}

func adminUserAddCommentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	back := "/admin/user"
	if user != nil {
		back = fmt.Sprintf("/admin/user/%d", user.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	comment := r.PostFormValue("comment")
	if user == nil {
		data.Errors = append(data.Errors, "invalid user id")
	} else if strings.TrimSpace(comment) == "" {
		data.Errors = append(data.Errors, "empty comment")
	} else {
		if err := cd.Queries().InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: user.Idusers, Comment: comment}); err != nil {
			data.Errors = append(data.Errors, err.Error())
		} else {
			data.Messages = append(data.Messages, "comment added")
		}
	}
	RunTaskPageTmpl.Handle(w, r, data)
}
