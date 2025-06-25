package linker

import corecommon "github.com/arran4/goa4web/core/common"
import hcommon "github.com/arran4/goa4web/handlers/common"

// IndexItem exposes the navigation item type.
type IndexItem = corecommon.IndexItem

// CoreData exposes the common.CoreData type for handlers.
type CoreData = hcommon.CoreData

// LinkerTopicName is the default name for the hidden linker forum.
const LinkerTopicName = "A LINKER TOPIC"

// LinkerTopicDescription describes the hidden linker forum.
const LinkerTopicDescription = "THIS IS A HIDDEN FORUM FOR A LINKER TOPIC"
