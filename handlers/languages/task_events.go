package languages

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var RenameLanguageTask = tasks.BasicTaskEvent{
	EventName:     "Rename Language",
	Match:         tasks.HasTask("Rename Language"),
	ActionHandler: adminLanguagesRenamePage,
}

var DeleteLanguageTask = tasks.BasicTaskEvent{
	EventName:     "Delete Language",
	Match:         tasks.HasTask("Delete Language"),
	ActionHandler: adminLanguagesDeletePage,
}

var CreateLanguageTask = tasks.BasicTaskEvent{
	EventName:     "Create Language",
	Match:         tasks.HasTask("Create Language"),
	ActionHandler: adminLanguagesCreatePage,
}
