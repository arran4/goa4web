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
		UserPreferences struct{ EmailUpdates bool }
	}

	user, _ := r.Context().Value(ContextValues("user")).(*User)
	pref, _ := r.Context().Value(ContextValues("preference")).(*Preference)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		UserData: user,
	}
	if pref != nil && pref.Emailforumupdates.Valid {
		data.UserPreferences.EmailUpdates = pref.Emailforumupdates.Bool
	}

	// Custom Index???

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "userEmailPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userEmailSaveActionPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	updates := r.PostFormValue("emailupdates") != ""

	_, err := queries.GetPreferenceByUserID(r.Context(), uid)
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

	http.Redirect(w, r, "/user/email", http.StatusSeeOther)
}

func userEmailTestActionPage(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(ContextValues("user")).(*User)
	if user == nil || !user.Email.Valid {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	url := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	if err := notifyChange(r.Context(), getEmailProvider(), user.Email.String, url); err != nil {
		log.Printf("send test mail: %v", err)
	}
	http.Redirect(w, r, "/user/email", http.StatusSeeOther)
}
