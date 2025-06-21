package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		UserData        *User
		UserPreferences struct {
			EmailUpdates bool
		}
	}

	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	user, _ := r.Context().Value(ContextValues("user")).(*User)
	pref, _ := r.Context().Value(ContextValues("preference")).(*Preference)

	data := Data{
		CoreData: cd,
		UserData: user,
	}
	if pref != nil {
		data.UserPreferences.EmailUpdates = pref.Emailforumupdates.Valid && pref.Emailforumupdates.Bool
	}

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "userEmailPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userEmailSaveActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	updates := r.PostFormValue("emailupdates") != ""

	_, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetPreferenceByUserID Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	var execErr error
	if errors.Is(err, sql.ErrNoRows) {
    /// TODO use queries
		_, execErr = queries.db.ExecContext(r.Context(), "INSERT INTO preferences (emailforumupdates, users_idusers) VALUES (?, ?)", updates, uid)
	} else {
    /// TODO use queries
		_, execErr = queries.db.ExecContext(r.Context(), "UPDATE preferences SET emailforumupdates=? WHERE users_idusers=?", updates, uid)
	}
	if execErr != nil {
		log.Printf("Preference save Error: %s", execErr)
		http.Redirect(w, r, "?error="+execErr.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}

func userEmailTestActionPage(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(ContextValues("user")).(*User)
	if user == nil || !user.Email.Valid {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	if user != nil && user.Email.Valid {
		provider := getEmailProvider()
		if provider != nil {
			url := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
			if err := notifyChange(r.Context(), provider, user.Email.String, url); err != nil {
				log.Printf("notifyChange Error: %s", err)
			}
		}
	}

	taskDoneAutoRefreshPage(w, r)
}
