package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveDigestTask updates notification digest settings.
type SaveDigestTask struct{ tasks.TaskString }

var saveDigestTask = &SaveDigestTask{TaskString: TaskSaveDigest}

var _ tasks.Task = (*SaveDigestTask)(nil)

func (SaveDigestTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		return common.UserError{ErrorMessage: "forbidden"}
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}


	hourStr := r.PostFormValue("daily_digest_hour")
	var hour *int
	if hourStr != "" && hourStr != "-1" {
		if h, err := strconv.Atoi(hourStr); err == nil && h >= 0 && h <= 23 {
			hour = &h
		}
	}

	weeklyDayStr := r.PostFormValue("weekly_digest_day")
	var weeklyDay *int
	if weeklyDayStr != "" && weeklyDayStr != "-1" {
		if d, err := strconv.Atoi(weeklyDayStr); err == nil && d >= 0 && d <= 6 {
			weeklyDay = &d
		}
	}

	weeklyHourStr := r.PostFormValue("weekly_digest_hour")
	var weeklyHour *int
	if weeklyHourStr != "" {
		if h, err := strconv.Atoi(weeklyHourStr); err == nil && h >= 0 && h <= 23 {
			weeklyHour = &h
		}
	}

	monthlyDayStr := r.PostFormValue("monthly_digest_day")
	var monthlyDay *int
	if monthlyDayStr != "" && monthlyDayStr != "-1" {
		if d, err := strconv.Atoi(monthlyDayStr); err == nil && d >= 1 && d <= 31 {
			monthlyDay = &d
		}
	}

	monthlyHourStr := r.PostFormValue("monthly_digest_hour")
	var monthlyHour *int
	if monthlyHourStr != "" {
		if h, err := strconv.Atoi(monthlyHourStr); err == nil && h >= 0 && h <= 23 {
			monthlyHour = &h
		}
	}

	markRead := r.PostFormValue("daily_digest_mark_read") == "on"

	if err := cd.SaveNotificationDigestPreferences(uid, hour, markRead, weeklyDay, weeklyHour, monthlyDay, monthlyHour); err != nil {
		log.Printf("save digest pref: %v", err)
		return fmt.Errorf("save digest pref fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: "/usr/notifications"}
}
