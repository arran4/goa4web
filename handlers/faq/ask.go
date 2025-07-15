package faq

import (
	"database/sql"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func AskPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Languages          []*db.Language
		SelectedLanguageId int32
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	if err := templates.RenderTemplate(w, "askPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AskActionPage(w http.ResponseWriter, r *http.Request) {
	if err := common.ValidateForm(r, []string{"language", "text"}, []string{"language", "text"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if err := queries.CreateFAQQuestion(r.Context(), db.CreateFAQQuestionParams{
		Question: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(languageId),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData); ok {
		evt := cd.Event()
		if evt == nil {
			log.Printf("ask action: missing event")
			if corecommon.Version == "dev" {
				// TODO remove once TaskEventMiddleware always provides an event
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		} else {
			evt.Admin = true
			evt.Path = "/admin/faq"
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["question"] = text
		}
	}

	// The BusWorker sends notifications based on event metadata.
	// Setting Admin=true signals administrators should be alerted.

	http.Redirect(w, r, "/faq", http.StatusTemporaryRedirect)
}
