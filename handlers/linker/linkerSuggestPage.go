package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core"
)

func SuggestPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories         []*db.LinkerCategory
		Languages          []*db.Language
		SelectedLanguageId int
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Suggest Link"
	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}

	uid := cd.UserID
	categoryRows, err := queries.GetAllLinkerCategoriesForUser(r.Context(), db.GetAllLinkerCategoriesForUserParams{
		ViewerID:     uid,
		ViewerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			//http.Redirect(w, r, "?error="+err.Error(), http.StatusSeeOther)
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}

	data.Categories = categoryRows

	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	LinkerSuggestPageTmpl.Handle(w, r, data)
}

const LinkerSuggestPageTmpl handlers.Page = "linker/suggestPage.gohtml"

type SuggestTask struct{ tasks.TaskString }

var suggestTask = SuggestTask{TaskString: TaskSuggest}
var _ tasks.Task = (*SuggestTask)(nil)

func (SuggestTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

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
		WriterID:        uid,
		CategoryID:      sql.NullInt32{Int32: int32(category), Valid: category != 0},
		Title:           sql.NullString{Valid: true, String: title},
		Url:             sql.NullString{Valid: true, String: url},
		Description:     sql.NullString{Valid: true, String: description},
		Timezone:        sql.NullString{String: cd.Location().String(), Valid: true},
		GrantCategoryID: sql.NullInt32{Int32: int32(category), Valid: category != 0},
		GranteeID:       sql.NullInt32{Int32: uid, Valid: true},
	}); err != nil {
		return fmt.Errorf("create linker queued item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
