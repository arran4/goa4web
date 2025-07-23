package writings

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func ArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
		Writing            *db.GetWritingByIdForUserDescendingByPublishedDateRow
		UserId             int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(cd.PreferredLanguageID(config.AppRuntimeConfig.DefaultLanguage)),
	}

	// article ID is validated by the RequireWritingAuthor middleware, so we
	// no longer need to parse it here.

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	writing := r.Context().Value(consts.KeyWriting).(*db.GetWritingByIdForUserDescendingByPublishedDateRow)
	data.Writing = writing

	languageRows, err := cd.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "articleEditPage.gohtml", data)
}
