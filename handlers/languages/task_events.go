package languages

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

var RenameLanguageTask = eventbus.BasicTaskEvent{
	EventName: "Rename Language",
	Match:     hcommon.TaskMatcher("Rename Language"),
	ActionH:   adminLanguagesRenamePage,
}

var DeleteLanguageTask = eventbus.BasicTaskEvent{
	EventName: "Delete Language",
	Match:     hcommon.TaskMatcher("Delete Language"),
	ActionH:   adminLanguagesDeletePage,
}

var CreateLanguageTask = eventbus.BasicTaskEvent{
	EventName: "Create Language",
	Match:     hcommon.TaskMatcher("Create Language"),
	ActionH:   adminLanguagesCreatePage,
}
