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
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "userLangPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userLangSaveLanguagesActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session, _ := GetSession(r)
	uid, _ := session.Values["UID"].(int32)

	if err := saveUserLanguages(r, queries, uid); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/user/lang", http.StatusTemporaryRedirect)
}

func userLangSaveLanguageActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session, _ := GetSession(r)
	uid, _ := session.Values["UID"].(int32)

	if err := saveUserLanguagePreference(r, queries, uid); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/user/lang", http.StatusTemporaryRedirect)
}

func userLangSaveAllActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session, _ := GetSession(r)
	uid, _ := session.Values["UID"].(int32)

	if err := saveUserLanguages(r, queries, uid); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := saveUserLanguagePreference(r, queries, uid); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/user/lang", http.StatusTemporaryRedirect)
}

func saveUserLanguages(r *http.Request, queries *Queries, uid int32) error {
	langs, err := queries.FetchLanguages(r.Context())
	if err != nil {
		return err
	}

	if err := queries.DeleteUserLanguagesByUser(r.Context(), uid); err != nil {
		return err
	}

	for _, l := range langs {
		if r.PostFormValue(fmt.Sprintf("language%d", l.Idlanguage)) != "" {
			if err := queries.InsertUserLang(r.Context(), InsertUserLangParams{
				UsersIdusers:       uid,
				LanguageIdlanguage: l.Idlanguage,
			}); err != nil {
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
