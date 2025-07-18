package forum

import corecommon "github.com/arran4/goa4web/core/common"
import hcommon "github.com/arran4/goa4web/handlers/common"

// IndexItem exposes the navigation item type.
type IndexItem = corecommon.IndexItem

// CoreData exposes the common.CoreData type for handlers.
type CoreData = hcommon.CoreData

// ForumTopicName is the default name for the hidden forum topic.
const ForumTopicName = "A FORUM TOPIC"

// ForumTopicDescription describes the hidden forum topic.
const ForumTopicDescription = "THIS IS A HIDDEN FORUM FOR A FORUM TOPIC"
