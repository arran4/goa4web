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
		LanguageOptions       []LanguageOption
		DefaultIsMultilingual bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Languages"
	queries := cd.Queries()

	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get preference: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	userLangs, err := queries.GetUserLanguages(r.Context(), cd.UserID)
	if err != nil {
		log.Printf("Error getting user languages: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	langs, err := cd.Languages()
	if err != nil {
		log.Printf("Error getting languages: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	selected := make(map[int32]bool)
	if len(userLangs) == 0 {
		for _, l := range langs {
			selected[l.ID] = true
		}
	} else {
		for _, ul := range userLangs {
			selected[ul.LanguageID] = true
		}
	}

	var opts []LanguageOption
	for _, l := range langs {
		opt := LanguageOption{ID: l.ID, Name: l.Nameof.String}
		if selected[l.ID] {
			opt.IsSelected = true
		}
		if pref != nil && pref.LanguageID.Valid && pref.LanguageID.Int32 == l.ID {
			opt.IsDefault = true
		}
		opts = append(opts, opt)
	}

	defaultIsMulti := pref == nil || !pref.LanguageID.Valid
	data := Data{
		LanguageOptions:       opts,
		DefaultIsMultilingual: defaultIsMulti,
	}

	handlers.TemplateHandler(w, r, "user/langPage.gohtml", data)
}

// updateLanguageSelections stores the languages selected by the user.
func updateLanguageSelections(r *http.Request, cd *common.CoreData, queries db.Querier, uid int32) error {
	// Clear existing language selections for the user.
	if err := queries.DeleteUserLanguagesForUser(r.Context(), uid); err != nil {
		return err
	}

	langs, err := cd.Languages()
	if err != nil {
		return err
	}

	for _, l := range langs {
		if r.PostFormValue(fmt.Sprintf("language%d", l.ID)) != "" {
			if err := queries.InsertUserLang(r.Context(), db.InsertUserLangParams{UsersIdusers: uid, LanguageID: l.ID}); err != nil {
				return err
			}
		}
	}
	return nil
}

// updateDefaultLanguage sets the user's preferred language.
func updateDefaultLanguage(r *http.Request, queries db.Querier, uid int32) error {
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
		return queries.InsertPreferenceForLister(r.Context(), db.InsertPreferenceForListerParams{
			LanguageID: sql.NullInt32{Int32: int32(langID), Valid: langID != 0},
			ListerID:   uid,
			PageSize:   int32(cd.Config.PageSizeDefault),
			Timezone:   sql.NullString{},
		})
	}

	pref.LanguageID = sql.NullInt32{Int32: int32(langID), Valid: langID != 0}
	return queries.UpdatePreferenceForLister(r.Context(), db.UpdatePreferenceForListerParams{
		LanguageID: pref.LanguageID,
		ListerID:   uid,
		PageSize:   pref.PageSize,
		Timezone:   pref.Timezone,
	})
}
