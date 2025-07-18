package linker

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

func SuggestPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories         []*db.LinkerCategory
		Languages          []*db.Language
		SelectedLanguageId int
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(common.KeyCoreData).(*common.CoreData),
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
	}

	categoryRows, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "suggestPage.gohtml", data)
}

type SuggestTask struct{ tasks.TaskString }

var suggestTask = SuggestTask{TaskString: TaskSuggest}

func (SuggestTask) Page(w http.ResponseWriter, r *http.Request) { SuggestPage(w, r) }

func (SuggestTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	if err := queries.CreateLinkerQueuedItem(r.Context(), db.CreateLinkerQueuedItemParams{
		UsersIdusers:     uid,
		LinkerCategoryID: int32(category),
		Title:            sql.NullString{Valid: true, String: title},
		Url:              sql.NullString{Valid: true, String: url},
		Description:      sql.NullString{Valid: true, String: description},
	}); err != nil {
		log.Printf("createLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
