package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func userPagingPage(w http.ResponseWriter, r *http.Request) {
	pref, _ := r.Context().Value(ContextValues("preference")).(*Preference)
	size := DefaultPageSize
	if pref != nil {
		size = int(pref.PageSize)
	}
	data := struct {
		*CoreData
		Size int
		Min  int
		Max  int
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Size:     size,
		Min:      appPaginationConfig.Min,
		Max:      appPaginationConfig.Max,
	}
	if err := renderTemplate(w, r, "userPagingPage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userPagingSaveActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/paging", http.StatusSeeOther)
		return
	}
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	size, _ := strconv.Atoi(r.FormValue("size"))
	if size < appPaginationConfig.Min {
		size = appPaginationConfig.Min
	}
	if size > appPaginationConfig.Max {
		size = appPaginationConfig.Max
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	pref, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertPreference(r.Context(), InsertPreferenceParams{
				LanguageIdlanguage: 0,
				UsersIdusers:       uid,
				PageSize:           int32(size),
			})
		}
	} else {
		pref.PageSize = int32(size)
		err = queries.UpdatePreference(r.Context(), UpdatePreferenceParams{
			LanguageIdlanguage: pref.LanguageIdlanguage,
			UsersIdusers:       uid,
			PageSize:           pref.PageSize,
		})
	}
	if err != nil {
		log.Printf("save paging: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/usr/paging", http.StatusSeeOther)
}
