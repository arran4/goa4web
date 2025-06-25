package blogs

import corecommon "github.com/arran4/goa4web/core/common"
import hcommon "github.com/arran4/goa4web/handlers/common"

// IndexItem exposes the navigation item type.
type IndexItem = corecommon.IndexItem

// CoreData exposes the common.CoreData type for handlers.
type CoreData = hcommon.CoreData

// BloggerTopicName is the default name for the hidden blogger forum.
const BloggerTopicName = "A BLOGGER TOPIC"

// BloggerTopicDescription describes the hidden blogger forum.
const BloggerTopicDescription = "THIS IS A HIDDEN FORUM FOR A BLOGGER TOPIC"
