package user

import (
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"

	"github.com/arran4/goa4web/core"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	if data.CoreData.UserID == 0 {
		session, err := core.GetSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		redirectToLogin(w, r, session)
		return
	}

	common.TemplateHandler(w, r, "userPage", data)
}
