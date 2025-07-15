package user

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
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
)

func userLangPage(w http.ResponseWriter, r *http.Request) {
	type LanguageOption struct {
		ID         int32
		Name       string
		IsSelected bool
		IsDefault  bool
	}

	type Data struct {
		*common.CoreData
		LanguageOptions []LanguageOption
	}

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	pref, _ := cd.Preference()
	userLangs, _ := queries.GetUserLanguages(r.Context(), cd.UserID)

	langs, err := cd.Languages()
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
func saveUserLanguages(r *http.Request, cd *common.CoreData, queries *db.Queries, uid int32) error {
	// Clear existing language selections for the user.
	if _, err := queries.DB().ExecContext(r.Context(), "DELETE FROM user_language WHERE users_idusers = ?", uid); err != nil {
		return err
	}

	langs, err := cd.Languages()
	if err != nil {
		return err
	}

	for _, l := range langs {
		if r.PostFormValue(fmt.Sprintf("language%d", l.Idlanguage)) != "" {
			if _, err := queries.DB().ExecContext(r.Context(), "INSERT INTO user_language (users_idusers, language_idlanguage) VALUES (?, ?)", uid, l.Idlanguage); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveUserLanguagePreference(r *http.Request, queries *db.Queries, uid int32) error {
	langID, err := strconv.Atoi(r.PostFormValue("defaultLanguage"))
	if err != nil {
		return err
	}

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return queries.InsertPreference(r.Context(), db.InsertPreferenceParams{
			LanguageIdlanguage: int32(langID),
			UsersIdusers:       uid,
			PageSize:           int32(config.AppRuntimeConfig.PageSizeDefault),
		})
	}

	pref.LanguageIdlanguage = int32(langID)
	return queries.UpdatePreference(r.Context(), db.UpdatePreferenceParams{
		LanguageIdlanguage: pref.LanguageIdlanguage,
		UsersIdusers:       uid,
		PageSize:           pref.PageSize,
	})
}

func saveDefaultLanguage(r *http.Request, queries *db.Queries, uid int32) error {
	langID, _ := strconv.Atoi(r.PostFormValue("defaultLanguage"))

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = queries.DB().ExecContext(r.Context(), "INSERT INTO preferences (language_idlanguage, users_idusers, page_size) VALUES (?, ?, ?)", langID, uid, config.AppRuntimeConfig.PageSizeDefault)
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
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := saveUserLanguages(r, cd, queries, uid); err != nil {
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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := saveUserLanguagePreference(r, queries, uid); err != nil {
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
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := saveUserLanguages(r, cd, queries, uid); err != nil {
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
