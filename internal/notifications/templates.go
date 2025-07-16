package notifications

import (
	_ "embed"
	hcommon "github.com/arran4/goa4web/internal/tasks"
	"strings"
)

var (
	//go:embed ../../core/templates/notifications/reply.txt
	replyTemplate string
	//go:embed ../../core/templates/notifications/thread.txt
	threadTemplate string
	//go:embed ../../core/templates/notifications/blog.txt
	blogTemplate string
	//go:embed ../../core/templates/notifications/writing.txt
	writingTemplate string
	//go:embed ../../core/templates/notifications/signup.txt
	signupTemplate string
	//go:embed ../../core/templates/notifications/ask.txt
	askTemplate string
	//go:embed ../../core/templates/notifications/set_user_level.txt
	setUserLevelTemplate string
	//go:embed ../../core/templates/notifications/update_user_level.txt
	updateUserLevelTemplate string
	//go:embed ../../core/templates/notifications/delete_user_level.txt
	deleteUserLevelTemplate string
	//go:embed ../../core/templates/notifications/set_topic_restriction.txt
	setTopicRestrictionTemplate string
	//go:embed ../../core/templates/notifications/update_topic_restriction.txt
	updateTopicRestrictionTemplate string
	//go:embed ../../core/templates/notifications/delete_topic_restriction.txt
	deleteTopicRestrictionTemplate string
	//go:embed ../../core/templates/notifications/copy_topic_restriction.txt
	copyTopicRestrictionTemplate string
	//go:embed ../../core/templates/notifications/password_reset.txt
	passwordResetTemplate string
)

var defaultTemplates = map[string]string{
	strings.ToLower(hcommon.TaskReply):                  replyTemplate,
	strings.ToLower(hcommon.TaskCreateThread):           threadTemplate,
	strings.ToLower(hcommon.TaskNewPost):                blogTemplate,
	strings.ToLower(hcommon.TaskSubmitWriting):          writingTemplate,
	strings.ToLower(hcommon.TaskRegister):               signupTemplate,
	strings.ToLower(hcommon.TaskAsk):                    askTemplate,
	strings.ToLower(hcommon.TaskSetUserLevel):           setUserLevelTemplate,
	strings.ToLower(hcommon.TaskUpdateUserLevel):        updateUserLevelTemplate,
	strings.ToLower(hcommon.TaskDeleteUserLevel):        deleteUserLevelTemplate,
	strings.ToLower(hcommon.TaskSetTopicRestriction):    setTopicRestrictionTemplate,
	strings.ToLower(hcommon.TaskUpdateTopicRestriction): updateTopicRestrictionTemplate,
	strings.ToLower(hcommon.TaskDeleteTopicRestriction): deleteTopicRestrictionTemplate,
	strings.ToLower(hcommon.TaskCopyTopicRestriction):   copyTopicRestrictionTemplate,
	strings.ToLower(hcommon.TaskUserResetPassword):      passwordResetTemplate,
}
