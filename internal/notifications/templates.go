package notifications

import (
	_ "embed"
	"strings"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

var (
	//go:embed templates/reply.txt
	replyTemplate string
	//go:embed templates/thread.txt
	threadTemplate string
	//go:embed templates/blog.txt
	blogTemplate string
	//go:embed templates/writing.txt
	writingTemplate string
	//go:embed templates/signup.txt
	signupTemplate string
	//go:embed templates/ask.txt
	askTemplate string
	//go:embed templates/set_user_level.txt
	setUserLevelTemplate string
	//go:embed templates/update_user_level.txt
	updateUserLevelTemplate string
	//go:embed templates/delete_user_level.txt
	deleteUserLevelTemplate string
	//go:embed templates/set_topic_restriction.txt
	setTopicRestrictionTemplate string
	//go:embed templates/update_topic_restriction.txt
	updateTopicRestrictionTemplate string
	//go:embed templates/delete_topic_restriction.txt
	deleteTopicRestrictionTemplate string
	//go:embed templates/copy_topic_restriction.txt
	copyTopicRestrictionTemplate string
	//go:embed templates/password_reset.txt
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
