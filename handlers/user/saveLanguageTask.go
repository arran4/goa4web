package user

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// SaveLanguageTask updates the user's default language preference.
type SaveLanguageTask struct{ tasks.TaskString }

var saveLanguageTask = &SaveLanguageTask{TaskString: tasks.TaskString(TaskSaveLanguage)}

var _ tasks.Task = (*SaveLanguageTask)(nil)

func (SaveLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := updateDefaultLanguage(r, queries, uid); err != nil {
		log.Printf("Save language Error: %s", err)
		return fmt.Errorf("save language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
