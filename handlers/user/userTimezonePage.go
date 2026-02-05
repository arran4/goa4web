package user

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func userTimezonePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Timezone"

	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get preference: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	tz := ""
	if pref != nil && pref.Timezone.Valid {
		tz = pref.Timezone.String
	}
	type Data struct {
		Timezone  string
		Timezones []string
	}
	data := Data{
		Timezone:  tz,
		Timezones: getAvailableTimezones(),
	}
	UserTimezonePage.Handle(w, r, data)
}

const UserTimezonePage tasks.Template = "user/timezonePage.gohtml"
