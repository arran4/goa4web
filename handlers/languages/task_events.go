package languages

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

type RenameLanguageTask struct{ tasks.TaskString }

var _ tasks.Task = (*RenameLanguageTask)(nil)

var renameLanguageTask = &RenameLanguageTask{TaskString: tasks.TaskString("Rename Language")}

func (RenameLanguageTask) Action(w http.ResponseWriter, r *http.Request) {
	adminLanguagesRenamePage(w, r)
}

type DeleteLanguageTask struct{ tasks.TaskString }

var _ tasks.Task = (*DeleteLanguageTask)(nil)

var deleteLanguageTask = &DeleteLanguageTask{TaskString: tasks.TaskString("Delete Language")}

func (DeleteLanguageTask) Action(w http.ResponseWriter, r *http.Request) {
	adminLanguagesDeletePage(w, r)
}

type CreateLanguageTask struct{ tasks.TaskString }

var _ tasks.Task = (*CreateLanguageTask)(nil)

var createLanguageTask = &CreateLanguageTask{TaskString: tasks.TaskString("Create Language")}

func (CreateLanguageTask) Action(w http.ResponseWriter, r *http.Request) {
	adminLanguagesCreatePage(w, r)
}
