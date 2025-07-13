package languages

import hcommon "github.com/arran4/goa4web/handlers/common"

var RenameLanguageTask = hcommon.NewTaskEvent("Rename Language")
var DeleteLanguageTask = hcommon.NewTaskEvent("Delete Language")
var CreateLanguageTask = hcommon.NewTaskEvent("Create Language")
