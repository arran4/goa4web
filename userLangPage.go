package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func userLangPage(w http.ResponseWriter, r *http.Request) {
	type LanguageOption struct {
		ID         int32
		Name       string
		IsSelected bool
		IsDefault  bool
	}

	type Data struct {
		*CoreData
		LanguageOptions []LanguageOption
	}

	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	pref, _ := r.Context().Value(ContextValues("preference")).(*Preference)
	userLangs, _ := r.Context().Value(ContextValues("languages")).([]*Userlang)

	langs, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	selected := make(map[int32]bool)
	for _, ul := range userLangs {
		selected[ul.LanguageIdlanguage] = true
	}

	var opts []LanguageOption
	for _, l := range langs {
		opt := LanguageOption{ID: l.Idlanguage, Name: l.Nameof.String}
		if selected[l.Idlanguage] {
			opt.IsSelected = true
		}
		if pref != nil && pref.LanguageIdlanguage == l.Idlanguage {
			opt.IsDefault = true
		}
		opts = append(opts, opt)
	}

	data := Data{
		CoreData:        cd,
		LanguageOptions: opts,
	}

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "userLangPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func saveUserLanguages(r *http.Request, queries *Queries, uid int32) error {
	langs, err := queries.FetchLanguages(r.Context())
	if err != nil {
		return err
	}
	// Clear existing language selections for the user.
	if _, err := queries.db.ExecContext(r.Context(), "DELETE FROM userlang WHERE users_idusers = ?", uid); err != nil {
		return err
	}
	for _, l := range langs {
		if r.PostFormValue(fmt.Sprintf("language%d", l.Idlanguage)) != "" {
			// TODO use queries
			if _, err := queries.db.ExecContext(r.Context(), "INSERT INTO userlang (users_idusers, language_idlanguage) VALUES (?, ?)", uid, l.Idlanguage); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveUserLanguagePreference(r *http.Request, queries *Queries, uid int32) error {
	langID, err := strconv.Atoi(r.PostFormValue("defaultLanguage"))
	if err != nil {
		return err
	}

	pref, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return queries.InsertPreference(r.Context(), InsertPreferenceParams{
				LanguageIdlanguage: int32(langID),
				UsersIdusers:       uid,
			})
		}
		return err
	}

	pref.LanguageIdlanguage = int32(langID)
	return queries.UpdatePreference(r.Context(), UpdatePreferenceParams{
		LanguageIdlanguage: pref.LanguageIdlanguage,
		UsersIdusers:       uid,
	})
}

func saveDefaultLanguage(r *http.Request, queries *Queries, uid int32) error {
	langID, _ := strconv.Atoi(r.PostFormValue("defaultLanguage"))
	_, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		_, err = queries.db.ExecContext(r.Context(), "INSERT INTO preferences (language_idlanguage, users_idusers) VALUES (?, ?)", langID, uid)
	} else {
		_, err = queries.db.ExecContext(r.Context(), "UPDATE preferences SET language_idlanguage=? WHERE users_idusers=?", langID, uid)
	}
	return err
}

func userLangSaveLanguagesActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := saveUserLanguages(r, queries, uid); err != nil {
		log.Printf("Save languages Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/user/lang", http.StatusTemporaryRedirect)
}

func userLangSaveLanguageActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := saveUserLanguagePreference(r, queries, uid); err != nil {
		log.Printf("Save language Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}

func userLangSaveDefaultLanguageActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := saveDefaultLanguage(r, queries, uid); err != nil {
		log.Printf("Save language Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/user/lang", http.StatusTemporaryRedirect)
}

func userLangSaveAllActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := saveUserLanguages(r, queries, uid); err != nil {
		log.Printf("Save languages Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := saveDefaultLanguage(r, queries, uid); err != nil {
		log.Printf("Save language Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/user/lang", http.StatusTemporaryRedirect)
}
