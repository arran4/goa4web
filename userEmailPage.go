package goa4web

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const errMailNotConfigured = "mail isn't configured" // shown when Test mail has no provider

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		UserData        *User
		UserPreferences struct{ EmailUpdates bool }
		Error           string
	}

	user, _ := r.Context().Value(ContextValues("user")).(*User)
	pref, _ := r.Context().Value(ContextValues("preference")).(*Preference)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		UserData: user,
		Error:    r.URL.Query().Get("error"),
	}
	if pref != nil && pref.Emailforumupdates.Valid {
		data.UserPreferences.EmailUpdates = pref.Emailforumupdates.Bool
	}

	if err := renderTemplate(w, r, "userEmailPage.gohtml", data); err != nil {
		log.Printf("user email page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userEmailSaveActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := GetSessionOrFail(w, r)
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

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	updates := r.PostFormValue("emailupdates") != ""

	_, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetPreferenceByUserID Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertEmailPreference(r.Context(), InsertEmailPreferenceParams{
				Emailforumupdates: sql.NullBool{Bool: updates, Valid: true},
				UsersIdusers:      uid,
			})
		}
	} else {
		err = queries.UpdateEmailForumUpdatesByUserID(r.Context(), UpdateEmailForumUpdatesByUserIDParams{
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
	user, _ := r.Context().Value(ContextValues("user")).(*User)
	if user == nil || !user.Email.Valid {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	base := "http://" + r.Host
	if appRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(appRuntimeConfig.HTTPHostname, "/")
	}
	pageURL := base + r.URL.Path
	provider := getEmailProvider()
	if provider == nil {
		q := url.QueryEscape(errMailNotConfigured)
		http.Redirect(w, r, "/usr/email?error="+q, http.StatusTemporaryRedirect)
		return
	}
	if err := notifyChange(r.Context(), provider, user.Email.String, pageURL); err != nil {
		log.Printf("notifyChange Error: %s", err)
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}
