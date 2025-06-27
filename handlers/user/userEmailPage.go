package user

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/runtimeconfig"
)

const ErrMailNotConfigured = "mail isn't configured" // shown when Test mail has no provider

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		UserData        *db.User
		UserPreferences struct{ EmailUpdates bool }
		Error           string
	}

	user, _ := r.Context().Value(common.KeyUser).(*db.User)
	pref, _ := r.Context().Value(common.KeyPreference).(*db.Preference)

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		UserData: user,
		Error:    r.URL.Query().Get("error"),
	}
	if pref != nil && pref.Emailforumupdates.Valid {
		data.UserPreferences.EmailUpdates = pref.Emailforumupdates.Bool
	}

	if err := templates.RenderTemplate(w, "emailPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("user email page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userEmailSaveActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	updates := r.PostFormValue("emailupdates") != ""

	_, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetPreferenceByUserID Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertEmailPreference(r.Context(), db.InsertEmailPreferenceParams{
				Emailforumupdates: sql.NullBool{Bool: updates, Valid: true},
				UsersIdusers:      uid,
			})
		}
	} else {
		err = queries.UpdateEmailForumUpdatesByUserID(r.Context(), db.UpdateEmailForumUpdatesByUserIDParams{
			Emailforumupdates: sql.NullBool{Bool: updates, Valid: true},
			UsersIdusers:      uid,
		})
	}
	if err != nil {
		log.Printf("save email pref: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailTestActionPage(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(common.KeyUser).(*db.User)
	if user == nil || !user.Email.Valid {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	base := "http://" + r.Host
	if runtimeconfig.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(runtimeconfig.AppRuntimeConfig.HTTPHostname, "/")
	}
	pageURL := base + r.URL.Path
	provider := getEmailProvider()
	if provider == nil {
		q := url.QueryEscape(ErrMailNotConfigured)
		// TaskDoneAutoRefreshPage triggers a client-side refresh to the
		// same URL. By adjusting the query string we can show the error
		// without an HTTP redirect that would repeat the POST.
		r.URL.RawQuery = "error=" + q
		common.TaskDoneAutoRefreshPage(w, r)
		return
	}
	if err := notifyChange(r.Context(), provider, user.Email.String, pageURL); err != nil {
		log.Printf("notifyChange Error: %s", err)
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}
