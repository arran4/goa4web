package writings

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	searchworker "github.com/arran4/goa4web/workers/searchworker"
)

// SubmitWritingTask encapsulates creating a new writing.
type SubmitWritingTask struct{ tasks.TaskString }

var submitWritingTask = &SubmitWritingTask{TaskString: TaskSubmitWriting}

var _ tasks.Task = (*SubmitWritingTask)(nil)

// followers of an author should be alerted when new writing is submitted
var _ notif.SubscribersNotificationTemplateProvider = (*SubmitWritingTask)(nil)

func (SubmitWritingTask) Page(w http.ResponseWriter, r *http.Request)   { ArticleAddPage(w, r) }
func (SubmitWritingTask) Action(w http.ResponseWriter, r *http.Request) { ArticleAddActionPage(w, r) }

func (SubmitWritingTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("writingEmail")
}

func (SubmitWritingTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("writing")
	return &s
}

// ReplyTask posts a comment reply.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)

// replying should notify anyone following the discussion
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)

// repliers expect to automatically follow further conversation
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

func (ReplyTask) AutoSubscribePath(evt eventbus.Event) (string, string) {
	return string(TaskReply), evt.Path
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) { ArticleReplyActionPage(w, r) }

// EditReplyTask updates an existing comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

var _ tasks.Task = (*EditReplyTask)(nil)

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	ArticleCommentEditActionPage(w, r)
}

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

var _ tasks.Task = (*CancelTask)(nil)

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) {
	ArticleCommentEditActionCancelPage(w, r)
}

// UpdateWritingTask applies changes to an article.
type UpdateWritingTask struct{ tasks.TaskString }

var updateWritingTask = &UpdateWritingTask{TaskString: TaskUpdateWriting}

var _ tasks.Task = (*UpdateWritingTask)(nil)

func (UpdateWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleEditPage(w, r) }

func (UpdateWritingTask) Action(w http.ResponseWriter, r *http.Request) { ArticleEditActionPage(w, r) }

// UserAllowTask grants a user a permission.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*UserAllowTask)(nil)

func (UserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/writings/users/levels" {
		AdminUserLevelsAllowActionPage(w, r)
		return
	}
	UsersPermissionsPermissionUserAllowPage(w, r)
}

// UserDisallowTask removes a user's permission.
type UserDisallowTask struct{ tasks.TaskString }

var userDisallowTask = &UserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*UserDisallowTask)(nil)

func (UserDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/writings/users/levels" {
		AdminUserLevelsRemoveActionPage(w, r)
		return
	}
	UsersPermissionsDisallowPage(w, r)
}

// WritingCategoryChangeTask modifies a category.
type WritingCategoryChangeTask struct{ tasks.TaskString }

var writingCategoryChangeTask = &WritingCategoryChangeTask{TaskString: TaskWritingCategoryChange}

var _ tasks.Task = (*WritingCategoryChangeTask)(nil)

func (WritingCategoryChangeTask) Action(w http.ResponseWriter, r *http.Request) {
	AdminCategoriesModifyPage(w, r)
}

// WritingCategoryCreateTask creates a new category.
type WritingCategoryCreateTask struct{ tasks.TaskString }

var writingCategoryCreateTask = &WritingCategoryCreateTask{TaskString: TaskWritingCategoryCreate}

var _ tasks.Task = (*WritingCategoryCreateTask)(nil)

func (WritingCategoryCreateTask) Action(w http.ResponseWriter, r *http.Request) {
	AdminCategoriesCreatePage(w, r)
}
