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

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"

	"github.com/arran4/goa4web/runtimeconfig"
)

// ErrMailNotConfigured is returned when test mail has no provider configured.
var ErrMailNotConfigured = errors.New("mail isn't configured")

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		UserData        *db.User
		UserPreferences struct {
			EmailUpdates         bool
			AutoSubscribeReplies bool
		}
		Error string
	}

	user, _ := r.Context().Value(common.KeyUser).(*db.User)
	pref, _ := r.Context().Value(common.KeyPreference).(*db.Preference)

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		UserData: user,
		Error:    r.URL.Query().Get("error"),
	}
	if pref != nil {
		if pref.Emailforumupdates.Valid {
			data.UserPreferences.EmailUpdates = pref.Emailforumupdates.Bool
		}
		data.UserPreferences.AutoSubscribeReplies = pref.AutoSubscribeReplies
	} else {
		data.UserPreferences.AutoSubscribeReplies = true
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
	auto := r.PostFormValue("autosubscribe") != ""

	_, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetPreferenceByUserID Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertEmailPreference(r.Context(), db.InsertEmailPreferenceParams{
				Emailforumupdates:    sql.NullBool{Bool: updates, Valid: true},
				AutoSubscribeReplies: auto,
				UsersIdusers:         uid,
			})
		}
	} else {
		err = queries.UpdateEmailForumUpdatesByUserID(r.Context(), db.UpdateEmailForumUpdatesByUserIDParams{
			Emailforumupdates: sql.NullBool{Bool: updates, Valid: true},
			UsersIdusers:      uid,
		})
		if err == nil {
			err = queries.UpdateAutoSubscribeRepliesByUserID(r.Context(), db.UpdateAutoSubscribeRepliesByUserIDParams{
				AutoSubscribeReplies: auto,
				UsersIdusers:         uid,
			})
		}
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
	provider := email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig)
	if provider == nil {
		q := url.QueryEscape(ErrMailNotConfigured.Error())
		// Display the error without redirecting so the POST isn't repeated.
		r.URL.RawQuery = "error=" + q
		common.TaskErrorAcknowledgementPage(w, r)
		return
	}
	if err := emailutil.NotifyChange(r.Context(), provider, user.Idusers, user.Email.String, pageURL, "update", nil); err != nil {
		log.Printf("notifyChange Error: %s", err)
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}
