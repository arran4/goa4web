package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core"
)

func SuggestPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories         []*db.LinkerCategory
		Languages          []*db.Language
		SelectedLanguageId int
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Suggest Link"
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}

	uid := data.CoreData.UserID
	categoryRows, err := queries.GetAllLinkerCategoriesForUser(r.Context(), db.GetAllLinkerCategoriesForUserParams{
		ViewerID:     uid,
		ViewerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
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
var _ tasks.Task = (*SuggestTask)(nil)

func (SuggestTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return nil
	}

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	if err := queries.CreateLinkerQueuedItemForWriter(r.Context(), db.CreateLinkerQueuedItemForWriterParams{
		WriterID:         uid,
		LinkerCategoryID: int32(category),
		Title:            sql.NullString{Valid: true, String: title},
		Url:              sql.NullString{Valid: true, String: url},
		Description:      sql.NullString{Valid: true, String: description},
		GrantCategoryID:  sql.NullInt32{Int32: int32(category), Valid: true},
		GranteeID:        sql.NullInt32{Int32: uid, Valid: true},
	}); err != nil {
		return fmt.Errorf("create linker queued item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
