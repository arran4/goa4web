package writings

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// SubmitWritingTask encapsulates creating a new writing.
type SubmitWritingTask struct{ tasks.TaskString }

var submitWritingTask = &SubmitWritingTask{TaskString: TaskSubmitWriting}

func (SubmitWritingTask) Page(w http.ResponseWriter, r *http.Request)   { ArticleAddPage(w, r) }
func (SubmitWritingTask) Action(w http.ResponseWriter, r *http.Request) { ArticleAddActionPage(w, r) }

// ReplyTask posts a comment reply.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) { ArticleReplyActionPage(w, r) }

// EditReplyTask updates an existing comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	ArticleCommentEditActionPage(w, r)
}

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) {
	ArticleCommentEditActionCancelPage(w, r)
}

// UpdateWritingTask applies changes to an article.
type UpdateWritingTask struct{ tasks.TaskString }

var updateWritingTask = &UpdateWritingTask{TaskString: TaskUpdateWriting}

func (UpdateWritingTask) Action(w http.ResponseWriter, r *http.Request) { ArticleEditActionPage(w, r) }

// UserAllowTask grants a user a permission.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

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

func (WritingCategoryChangeTask) Action(w http.ResponseWriter, r *http.Request) {
	AdminCategoriesModifyPage(w, r)
}

// WritingCategoryCreateTask creates a new category.
type WritingCategoryCreateTask struct{ tasks.TaskString }

var writingCategoryCreateTask = &WritingCategoryCreateTask{TaskString: TaskWritingCategoryCreate}

func (WritingCategoryCreateTask) Action(w http.ResponseWriter, r *http.Request) {
	AdminCategoriesCreatePage(w, r)
}
