package forum

import corecommon "github.com/arran4/goa4web/core/common"
import handlers "github.com/arran4/goa4web/handlers"

// IndexItem exposes the navigation item type.
type IndexItem = corecommon.IndexItem

// CoreData exposes the handlers.CoreData type for handlers.
type CoreData = handlers.CoreData

// ForumTopicName is the default name for the hidden forum topic.
const ForumTopicName = "A FORUM TOPIC"

// ForumTopicDescription describes the hidden forum topic.
const ForumTopicDescription = "THIS IS A HIDDEN FORUM FOR A FORUM TOPIC"
