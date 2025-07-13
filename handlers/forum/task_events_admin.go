package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// SetUserLevelTask updates a user's forum access level.
var SetUserLevelTask = eventbus.TaskEvent{
	Name:    hcommon.TaskSetUserLevel,
	Matcher: hcommon.TaskMatcher(hcommon.TaskSetUserLevel),
}

// UpdateUserLevelTask modifies a user's access level.
var UpdateUserLevelTask = eventbus.TaskEvent{
	Name:    hcommon.TaskUpdateUserLevel,
	Matcher: hcommon.TaskMatcher(hcommon.TaskUpdateUserLevel),
}

// DeleteUserLevelTask removes a user's access level.
var DeleteUserLevelTask = eventbus.TaskEvent{
	Name:    hcommon.TaskDeleteUserLevel,
	Matcher: hcommon.TaskMatcher(hcommon.TaskDeleteUserLevel),
}

// SetTopicRestrictionTask adds a topic restriction.
var SetTopicRestrictionTask = eventbus.TaskEvent{
	Name:    hcommon.TaskSetTopicRestriction,
	Matcher: hcommon.TaskMatcher(hcommon.TaskSetTopicRestriction),
}

// UpdateTopicRestrictionTask updates a topic restriction.
var UpdateTopicRestrictionTask = eventbus.TaskEvent{
	Name:    hcommon.TaskUpdateTopicRestriction,
	Matcher: hcommon.TaskMatcher(hcommon.TaskUpdateTopicRestriction),
}

// DeleteTopicRestrictionTask deletes a topic restriction.
var DeleteTopicRestrictionTask = eventbus.TaskEvent{
	Name:    hcommon.TaskDeleteTopicRestriction,
	Matcher: hcommon.TaskMatcher(hcommon.TaskDeleteTopicRestriction),
}

// CopyTopicRestrictionTask copies topic restrictions between topics.
var CopyTopicRestrictionTask = eventbus.TaskEvent{
	Name:    hcommon.TaskCopyTopicRestriction,
	Matcher: hcommon.TaskMatcher(hcommon.TaskCopyTopicRestriction),
}
