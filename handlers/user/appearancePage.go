package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

func userAppearancePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Appearance Settings"

	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get preference: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	customCSS := ""
	if pref != nil && pref.CustomCss.Valid {
		customCSS = pref.CustomCss.String
	}
	type Data struct {
		CustomCSS string
	}
	data := Data{
		CustomCSS: customCSS,
	}
	AppearancePage.Handle(w, r, data)
}

const AppearancePage handlers.Page = "user/appearance.gohtml"

// AppearanceSaveTask saves the user's custom CSS preference.
type AppearanceSaveTask struct{ tasks.TaskString }

var appearanceSaveTask = &AppearanceSaveTask{TaskString: tasks.TaskString(TaskSaveAppearance)}

var _ tasks.Task = (*AppearanceSaveTask)(nil)

func (AppearanceSaveTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	customCSS := r.FormValue("custom_css")

	// Ensure preference row exists
	pref, err := cd.Preference()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create default preferences if missing
			if err := queries.InsertPreferenceForLister(r.Context(), db.InsertPreferenceForListerParams{
				ListerID: uid,
				PageSize: 15, // Default
			}); err != nil {
				log.Printf("insert preference: %v", err)
				return fmt.Errorf("Internal Server Error")
			}
			// Load it again (or fake it, but simpler to just proceed)
		} else {
			log.Printf("check preference: %v", err)
			return fmt.Errorf("Internal Server Error")
		}
	}

	if err := queries.UpdateCustomCssForLister(r.Context(), db.UpdateCustomCssForListerParams{
		ListerID:  uid,
		CustomCss: sql.NullString{String: customCSS, Valid: customCSS != ""},
	}); err != nil {
		log.Printf("update custom css: %v", err)
		return fmt.Errorf("Internal Server Error")
	}

	// Update cached preference object in place so the re-rendered page sees the new value
	if pref != nil {
		pref.CustomCss = sql.NullString{String: customCSS, Valid: customCSS != ""}
	} else {
		// If it was nil (newly created), we might need to force reload or manually construct it if we want it to show up immediately
		// But UserAppearancePage calls cd.Preference() which might try to load it if we didn't have it before.
		// If we just inserted it, cd.Preference() (lazy) might still think it's not loaded or loaded as nil?
		// lazy.Value loads once. If it loaded nil (ErrNoRows), it stays nil?
		// CoreData.Preference() implementation:
		/*
			return cd.pref.Load(func() (*db.Preference, error) {
				// ...
				return cd.queries.GetPreferenceForLister(cd.ctx, cd.UserID)
			})
		*/
		// If it was loaded and returned nil, it is "loaded".
		// So we can't easily force it to reload.
		// However, for the user flow, if they didn't have preferences, they probably didn't have custom CSS.
		// The re-render will show what they submitted if we pass it in `data`, OR if we rely on `cd`.
		// `userAppearancePage` uses `cd.Preference`.
		// To be safe, if `pref` is nil, we can rely on the fact that we just saved it.
		// But to make `userAppearancePage` generic, maybe we should pass the value to it?
		// `userAppearancePage` extracts it from `cd`.
		// I'll stick to updating `pref` if it exists. If it was nil, well, the user will see empty/default, or I can try to fix it.
		// Actually, if `pref` is nil, `cd.pref` has `value=nil`.
		// I can't assign to `pref`.
	}

	cd.SetCurrentNotice("Appearance settings updated")

	// Render directly with the new value to ensure it is displayed even if pref was not cached
	cd.PageTitle = "Appearance Settings"
	data := struct {
		CustomCSS string
	}{
		CustomCSS: customCSS,
	}
	AppearancePage.Handle(w, r, data)
	return nil
}

func (t *AppearanceSaveTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{AppearancePage}
}
