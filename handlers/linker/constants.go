package linker

import corecommon "github.com/arran4/goa4web/core/common"
import handlers "github.com/arran4/goa4web/handlers"

// IndexItem exposes the navigation item type.
type IndexItem = corecommon.IndexItem

// CoreData exposes the handlers.CoreData type for handlers.
type CoreData = handlers.CoreData

// LinkerTopicName is the default name for the hidden linker forum.
const LinkerTopicName = "A LINKER TOPIC"

// LinkerTopicDescription describes the hidden linker forum.
const LinkerTopicDescription = "THIS IS A HIDDEN FORUM FOR A LINKER TOPIC"
