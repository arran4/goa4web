package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
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
		LanguageOptions       []LanguageOption
		DefaultIsMultilingual bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userLangs, err := queries.GetUserLanguages(r.Context(), cd.UserID)
	if err != nil {
		log.Printf("Error getting user languages: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	langs, err := cd.Languages()
	if err != nil {
		log.Printf("Error getting languages: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	selected := make(map[int32]bool)
	if len(userLangs) == 0 {
		for _, l := range langs {
			selected[l.Idlanguage] = true
		}
	} else {
		for _, ul := range userLangs {
			selected[ul.LanguageIdlanguage] = true
		}
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

	defaultIsMulti := pref == nil || pref.LanguageIdlanguage == 0
	data := Data{
		CoreData:              cd,
		LanguageOptions:       opts,
		DefaultIsMultilingual: defaultIsMulti,
	}

	handlers.TemplateHandler(w, r, "langPage.gohtml", data)
}

// updateLanguageSelections stores the languages selected by the user.
func updateLanguageSelections(r *http.Request, cd *common.CoreData, queries *db.Queries, uid int32) error {
	// Clear existing language selections for the user.
	if err := queries.DeleteUserLanguagesByUser(r.Context(), uid); err != nil {
		return err
	}

	langs, err := cd.Languages()
	if err != nil {
		return err
	}

	for _, l := range langs {
		if r.PostFormValue(fmt.Sprintf("language%d", l.Idlanguage)) != "" {
			if err := queries.InsertUserLang(r.Context(), db.InsertUserLangParams{UsersIdusers: uid, LanguageIdlanguage: l.Idlanguage}); err != nil {
				return err
			}
		}
	}
	return nil
}

// updateDefaultLanguage sets the user's preferred language.
func updateDefaultLanguage(r *http.Request, queries *db.Queries, uid int32) error {
	langID, err := strconv.Atoi(r.PostFormValue("defaultLanguage"))
	if err != nil {
		return err
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return queries.InsertPreference(r.Context(), db.InsertPreferenceParams{
			LanguageIdlanguage: int32(langID),
			UsersIdusers:       uid,
			PageSize:           int32(cd.Config.PageSizeDefault),
		})
	}

	pref.LanguageIdlanguage = int32(langID)
	return queries.UpdatePreference(r.Context(), db.UpdatePreferenceParams{
		LanguageIdlanguage: pref.LanguageIdlanguage,
		UsersIdusers:       uid,
		PageSize:           pref.PageSize,
	})
}
