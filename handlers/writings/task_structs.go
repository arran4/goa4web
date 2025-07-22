package writings

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

// SubmitWritingTask encapsulates creating a new writing.
type SubmitWritingTask struct{ tasks.TaskString }

var submitWritingTask = &SubmitWritingTask{TaskString: TaskSubmitWriting}

var _ tasks.Task = (*SubmitWritingTask)(nil)

// followers of an author should be alerted when new writing is submitted
var _ notif.SubscribersNotificationTemplateProvider = (*SubmitWritingTask)(nil)
var _ notif.GrantsRequiredProvider = (*SubmitWritingTask)(nil)

func (SubmitWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleAddPage(w, r) }
func (SubmitWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	ArticleAddActionPage(w, r)
	return nil
}

func (SubmitWritingTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("writingEmail")
}

func (SubmitWritingTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("writing")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider. The newly created article
// is referenced in evt.Data under the "target" key by the page handler.
func (SubmitWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

// ReplyTask posts a comment reply.
type ReplyTask struct{ tasks.TaskString }

// ReplyTask implements these interfaces so that when a user replies to a
// writing everyone following the discussion is automatically subscribed and
// receives a notification using the shared reply templates. This keeps readers
// informed when conversations continue.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)
var _ notif.GrantsRequiredProvider = (*ReplyTask)(nil)

// replying should notify anyone following the discussion
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)

// repliers expect to automatically follow further conversation
// ReplyTask notifies followers and auto-subscribes the author so replies aren't missed.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (ReplyTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider for replies.
func (ReplyTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

// AutoSubscribePath implements notif.AutoSubscribeProvider. It builds the
// subscription path for the writing's forum thread when that data is provided
// by the event.
func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskReply), evt.Path, nil
}

var _ searchworker.IndexedTask = ReplyTask{}
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	ArticleReplyActionPage(w, r)
	return nil
}

// EditReplyTask updates an existing comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

var _ tasks.Task = (*EditReplyTask)(nil)

// notify administrators when comments are edited so they can moderate discussions
// admins need to know when discussions change, notify them of edits
var _ notif.AdminEmailTemplateProvider = (*EditReplyTask)(nil)

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	ArticleCommentEditActionPage(w, r)
	return nil
}

func (EditReplyTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsCommentEditEmail")
}

func (EditReplyTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsCommentEditEmail")
	return &v
}

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

// CancelTask is only used to abort editing, implementing tasks.Task ensures it
// fits the routing interface even though no additional behaviour is required.
var _ tasks.Task = (*CancelTask)(nil)

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	ArticleCommentEditActionCancelPage(w, r)
	return nil
}

// UpdateWritingTask applies changes to an article.
type UpdateWritingTask struct{ tasks.TaskString }

var updateWritingTask = &UpdateWritingTask{TaskString: TaskUpdateWriting}

var _ tasks.Task = (*UpdateWritingTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*UpdateWritingTask)(nil)
var _ notif.GrantsRequiredProvider = (*UpdateWritingTask)(nil)

func (UpdateWritingTask) Page(w http.ResponseWriter, r *http.Request) { ArticleEditPage(w, r) }

func (UpdateWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	ArticleEditActionPage(w, r)
	return nil
}

func (UpdateWritingTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("writingUpdateEmail")
}

func (UpdateWritingTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("writing_update")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider for article updates.
func (UpdateWritingTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

// UserAllowTask grants a user a permission.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*UserAllowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserAllowTask)(nil)

func (UserAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	if r.URL.Path == "/admin/writings/users/roles" {
		AdminUserLevelsAllowActionPage(w, r)
		return nil
	}
	UsersPermissionsPermissionUserAllowPage(w, r)
	return nil
}

func (UserAllowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("setUserRoleEmail")
}

func (UserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}

// UserDisallowTask removes a user's permission.
type UserDisallowTask struct{ tasks.TaskString }

var userDisallowTask = &UserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*UserDisallowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserDisallowTask)(nil)

func (UserDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	if r.URL.Path == "/admin/writings/users/roles" {
		AdminUserLevelsRemoveActionPage(w, r)
		return nil
	}
	UsersPermissionsDisallowPage(w, r)
	return nil
}

func (UserDisallowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("deleteUserRoleEmail")
}

func (UserDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}

// WritingCategoryChangeTask modifies a category.
type WritingCategoryChangeTask struct{ tasks.TaskString }

var writingCategoryChangeTask = &WritingCategoryChangeTask{TaskString: TaskWritingCategoryChange}

var _ tasks.Task = (*WritingCategoryChangeTask)(nil)

func (WritingCategoryChangeTask) Action(w http.ResponseWriter, r *http.Request) any {
	AdminCategoriesModifyPage(w, r)
	return nil
}

// WritingCategoryCreateTask creates a new category.
type WritingCategoryCreateTask struct{ tasks.TaskString }

var writingCategoryCreateTask = &WritingCategoryCreateTask{TaskString: TaskWritingCategoryCreate}

var _ tasks.Task = (*WritingCategoryCreateTask)(nil)

func (WritingCategoryCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	AdminCategoriesCreatePage(w, r)
	return nil
}
