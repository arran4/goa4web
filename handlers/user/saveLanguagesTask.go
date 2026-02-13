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

// SaveLanguagesTask stores multiple language selections.
type SaveLanguagesTask struct{ tasks.TaskString }

var saveLanguagesTask = &SaveLanguagesTask{TaskString: tasks.TaskString(TaskSaveLanguages)}

var _ tasks.Task = (*SaveLanguagesTask)(nil)

func (SaveLanguagesTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := updateLanguageSelections(r, cd, queries, uid); err != nil {
		log.Printf("Save languages Error: %s", err)
		return fmt.Errorf("save languages fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
