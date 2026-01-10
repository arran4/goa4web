package user

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/tasks"
)

type userTask struct {
}

const (
	UserPageTmpl handlers.Page = "user/page.gohtml"
)

func NewUserTask() tasks.Task {
	return &userTask{}
}

func (t *userTask) TemplatesRequired() []string {
	return []string{string(UserPageTmpl)}
}

func (t *userTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *userTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "User Preferences"

	if cd.UserID == 0 {
		session, err := core.GetSession(r)
		if err != nil {
			log.Printf("get session: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
		_ = middleware.RedirectToLogin(w, r, session)
		return
	}

	UserPageTmpl.Handle(w, r, struct{}{})
}
