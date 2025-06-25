package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"

	"github.com/arran4/goa4web/runtimeconfig"
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

	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	pref, _ := r.Context().Value(common.KeyPreference).(*Preference)
	userLangs, _ := r.Context().Value(common.KeyLanguages).([]*Userlang)

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

	if err := templates.RenderTemplate(w, "langPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func saveUserLanguages(r *http.Request, queries *Queries, uid int32) error {
	// Clear existing language selections for the user.
	if _, err := queries.DB().ExecContext(r.Context(), "DELETE FROM userlang WHERE users_idusers = ?", uid); err != nil {
		return err
	}

	langs, err := queries.FetchLanguages(r.Context())
	if err != nil {
		return err
	}

	for _, l := range langs {
		if r.PostFormValue(fmt.Sprintf("language%d", l.Idlanguage)) != "" {
			if _, err := queries.DB().ExecContext(r.Context(), "INSERT INTO userlang (users_idusers, language_idlanguage) VALUES (?, ?)", uid, l.Idlanguage); err != nil {
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
				PageSize:           int32(runtimeconfig.AppRuntimeConfig.PageSizeDefault),
			})
		}
		return err
	}

	pref.LanguageIdlanguage = int32(langID)
	return queries.UpdatePreference(r.Context(), UpdatePreferenceParams{
		LanguageIdlanguage: pref.LanguageIdlanguage,
		UsersIdusers:       uid,
		PageSize:           pref.PageSize,
	})
}

func saveDefaultLanguage(r *http.Request, queries *Queries, uid int32) error {
	langID, _ := strconv.Atoi(r.PostFormValue("defaultLanguage"))
	pref, err := queries.GetPreferenceByUserID(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = queries.DB().ExecContext(r.Context(), "INSERT INTO preferences (language_idlanguage, users_idusers, page_size) VALUES (?, ?, ?)", langID, uid, runtimeconfig.AppRuntimeConfig.PageSizeDefault)
			return err
		}
		return err
	}
	_, err = queries.DB().ExecContext(r.Context(), "UPDATE preferences SET language_idlanguage=?, page_size=? WHERE users_idusers=?", langID, pref.PageSize, uid)
	return err
}

func userLangSaveLanguagesActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	if err := saveUserLanguages(r, queries, uid); err != nil {
		log.Printf("Save languages Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func userLangSaveLanguagePreferenceActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	if err := saveUserLanguagePreference(r, queries, uid); err != nil {
		log.Printf("Save language Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func userLangSaveDefaultLanguageActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	if err := saveDefaultLanguage(r, queries, uid); err != nil {
		log.Printf("Save language Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func userLangSaveAllActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*Queries)

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

	common.TaskDoneAutoRefreshPage(w, r)
}

// userLangSaveLanguageActionPage is kept for compatibility and forwards to
// userLangSaveLanguagePreferenceActionPage.
func userLangSaveLanguageActionPage(w http.ResponseWriter, r *http.Request) {
	userLangSaveLanguagePreferenceActionPage(w, r)
}
