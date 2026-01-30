package admin

import "github.com/arran4/goa4web/internal/tasks"

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's HasTask so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskAdd represents the "Add" action, commonly used when creating a new
	// record.
	TaskAdd tasks.TaskString = "Add"

	// TaskAddApproval adds an approval for a user or item.
	TaskAddApproval tasks.TaskString = "Add approval"

	// TaskAddEmail adds a new email to the user.
	TaskAddEmail tasks.TaskString = "Add Email"

	// TaskDeleteEmail deletes an email from the user.
	TaskDeleteEmail tasks.TaskString = "Delete Email"

	// TaskVerifyEmail verifies a user's email.
	TaskVerifyEmail tasks.TaskString = "Verify Email"

	// TaskUnverifyEmail unverifies a user's email.
	TaskUnverifyEmail tasks.TaskString = "Unverify Email"

	// TaskResendVerification resends the verification email.
	TaskResendVerification tasks.TaskString = "Resend Verification"

	// TaskUploadImage uploads an image file to the image board.
	TaskUploadImage tasks.TaskString = "Upload image"

	// TaskAnswer submits an answer in the FAQ admin interface.
	TaskAnswer tasks.TaskString = "Answer"

	// TaskApprove approves an item in moderation queues.
	TaskApprove tasks.TaskString = "Approve"

	// TaskApplySubscriptionTemplate applies a subscription template to a role.
	TaskApplySubscriptionTemplate tasks.TaskString = "Apply Template"

	// TaskAccept accepts a queued request.
	TaskAccept tasks.TaskString = "Accept"

	// TaskReject rejects a queued request.
	TaskReject tasks.TaskString = "Reject"

	// TaskQuery asks for more details about a queued request.
	TaskQuery tasks.TaskString = "Query"

	// TaskAsk submits a new question to the FAQ system.
	TaskAsk tasks.TaskString = "Ask"

	// TaskCancel cancels the current operation and returns to the previous
	// page.
	TaskCancel tasks.TaskString = "Cancel"

	// TaskCreate indicates creation of an object, for instance a bookmark or
	// other record.
	TaskCreate tasks.TaskString = "Create"

	// TaskCreateCategory creates a new category entry.
	TaskCreateCategory tasks.TaskString = "Create Category"

	// TaskCreateLanguage creates a new language entry.
	TaskCreateLanguage tasks.TaskString = "Create Language"

	// TaskCreateThread creates a new forum thread.
	TaskCreateThread tasks.TaskString = "Create Thread"

	// TaskDelete removes an existing item.
	TaskDelete tasks.TaskString = "Delete"

	// TaskImageCacheList refreshes the image cache list.
	TaskImageCacheList tasks.TaskString = "List cache"

	// TaskImageCachePrune prunes the image cache to the configured max size.
	TaskImageCachePrune tasks.TaskString = "Prune cache"

	// TaskImageCacheDelete deletes selected image cache entries.
	TaskImageCacheDelete tasks.TaskString = "Delete cache"

	// TaskDeactivate moves an item to the deactivated store.
	TaskDeactivate tasks.TaskString = "Deactivate"

	// TaskDBBackup creates and downloads a database backup.
	TaskDBBackup tasks.TaskString = "Database backup"

	// TaskDBRestore restores the database from a backup.
	TaskDBRestore tasks.TaskString = "Database restore"

	// TaskActivate restores a previously deactivated item.
	TaskActivate tasks.TaskString = "Activate"

	// TaskDeleteCategory removes a category.
	TaskDeleteCategory tasks.TaskString = "Delete Category"

	// TaskDeleteTopicRestriction deletes a topic restriction record.
	TaskDeleteTopicRestriction tasks.TaskString = "Delete topic restriction"

	// TaskDeleteUserApproval deletes a user's approval entry.
	TaskDeleteUserApproval tasks.TaskString = "Delete user approval"

	// TaskRevokeRole revokes a role from a user.
	TaskRevokeRole tasks.TaskString = "Revoke role"

	// TaskEdit modifies an existing item.
	TaskEdit tasks.TaskString = "Edit"

	// TaskEditReply edits a comment or reply.
	TaskEditReply tasks.TaskString = "Edit Reply"

	// TaskExportTemplates exports embedded templates.
	TaskExportTemplates tasks.TaskString = "Export Templates"

	// TaskForumCategoryChange updates the name of a forum category.
	TaskForumCategoryChange tasks.TaskString = "Forum category change"

	// TaskForumCategoryCreate creates a new forum category.
	TaskForumCategoryCreate tasks.TaskString = "Forum category create"

	// TaskForumTopicChange updates the name of a forum topic.
	TaskForumTopicChange tasks.TaskString = "Forum topic change"

	// TaskForumTopicCreate creates a new forum topic.
	TaskForumTopicCreate tasks.TaskString = "Forum topic create"

	// TaskForumTopicDelete removes a forum topic.
	TaskForumTopicDelete tasks.TaskString = "Forum topic delete"

	// TaskForumTopicConvertPrivate converts a forum topic to a private forum.
	TaskForumTopicConvertPrivate tasks.TaskString = "Forum topic convert private"

	// TaskForumThreadDelete removes a forum thread.
	TaskForumThreadDelete tasks.TaskString = "Forum thread delete"

	// TaskLogin performs a user login.
	TaskLogin tasks.TaskString = "Login"

	// TaskModifyCategory modifies an existing writing category.
	TaskModifyCategory tasks.TaskString = "Modify Category"

	// TaskModifyBoard modifies the settings of an image board.
	TaskModifyBoard tasks.TaskString = "Modify board"

	// TaskDeleteBoard deletes an existing image board.
	TaskDeleteBoard tasks.TaskString = "Delete board"

	// TaskNewCategory creates a new writing category.
	TaskNewCategory tasks.TaskString = "New Category"

	// TaskNewPost creates a new news post.
	TaskNewPost tasks.TaskString = "New Post"

	// TaskNewBoard creates a new image board.
	TaskNewBoard tasks.TaskString = "New board"

	// TaskRegister registers a new user account.
	TaskRegister tasks.TaskString = "Register"

	// TaskRemakeBlogSearch rebuilds the blog search index.
	TaskRemakeBlogSearch tasks.TaskString = "Remake blog search"

	// TaskRemakeCommentsSearch rebuilds the comments search index.
	TaskRemakeCommentsSearch tasks.TaskString = "Remake comments search"

	// TaskRemakeLinkerSearch rebuilds the linker search index.
	TaskRemakeLinkerSearch tasks.TaskString = "Remake linker search"

	// TaskRemakeNewsSearch rebuilds the news search index.
	TaskRemakeNewsSearch tasks.TaskString = "Remake news search"

	// TaskRemakeStatisticInformationOnForumthread updates forum thread
	// statistics.
	TaskRemakeStatisticInformationOnForumthread tasks.TaskString = "Remake statistic information on forumthread"

	// TaskRemakeStatisticInformationOnForumtopic updates forum topic
	// statistics.
	TaskRemakeStatisticInformationOnForumtopic tasks.TaskString = "Remake statistic information on forumtopic"

	// TaskRemakeWritingSearch rebuilds the writing search index.
	TaskRemakeWritingSearch tasks.TaskString = "Remake writing search"

	// TaskRemakeImageSearch rebuilds the image search index.
	TaskRemakeImageSearch tasks.TaskString = "Remake image search"

	// TaskRemoveRemove removes an item, typically from a list.
	TaskRemoveRemove tasks.TaskString = "Remove"

	// TaskRenameCategory renames a category.
	TaskRenameCategory tasks.TaskString = "Rename Category"

	// TaskReply posts a reply to a thread or comment.
	TaskReply tasks.TaskString = "Reply"

	// TaskSave persists changes for an item.
	TaskSave tasks.TaskString = "Save"

	// TaskSaveAll saves all changes in bulk.
	TaskSaveAll tasks.TaskString = "Save all"

	// TaskSaveLanguage saves updates to a single language.
	TaskSaveLanguage tasks.TaskString = "Save language"

	// TaskSaveLanguages saves multiple languages at once.
	TaskSaveLanguages tasks.TaskString = "Save languages"

	// TaskSearchBlogs triggers a blog search.
	TaskSearchBlogs tasks.TaskString = "Search blogs"

	// TaskSearchForum triggers a forum search.
	TaskSearchForum tasks.TaskString = "Search forum"

	// TaskSearchLinker triggers a linker search.
	TaskSearchLinker tasks.TaskString = "Search linker"

	// TaskSearchNews triggers a news search.
	TaskSearchNews tasks.TaskString = "Search news"

	// TaskSearchWritings triggers a writing search.
	TaskSearchWritings tasks.TaskString = "Search writings"

	// TaskSetTopicRestriction sets a new topic restriction.
	TaskSetTopicRestriction tasks.TaskString = "Set topic restriction"

	// TaskCopyTopicRestriction copies restriction levels from one topic to another.
	TaskCopyTopicRestriction tasks.TaskString = "Copy topic restriction"

	// TaskGrantRole grants a role to a user.
	TaskGrantRole tasks.TaskString = "Grant role"

	// TaskSubmitWriting submits a new writing.
	TaskSubmitWriting tasks.TaskString = "Submit writing"

	// TaskSuggest creates a suggestion in the linker.
	TaskSuggest tasks.TaskString = "Suggest"

	// TaskTestMail sends a test email to the current user.
	TaskTestMail tasks.TaskString = "Test mail"

	// TaskResend attempts to send queued emails immediately.
	TaskResend tasks.TaskString = "Resend"

	// TaskResendFilteredQueue retries all filtered queued emails.
	TaskResendFilteredQueue tasks.TaskString = "Retry all filtered"

	// TaskRetry queues a previously sent email for another attempt.
	TaskRetry tasks.TaskString = "Retry"

	// TaskDismiss marks a notification as read.
	TaskDismiss tasks.TaskString = "Dismiss"

	// TaskMarkUnread marks a notification as unread.
	TaskMarkUnread tasks.TaskString = "Mark unread"

	// TaskToggleRead toggles a notification's read state.
	TaskToggleRead tasks.TaskString = "Toggle read"

	// TaskDeleteNotification deletes selected notifications.
	TaskDeleteNotification tasks.TaskString = "Delete notification"

	// TaskPurgeSelected removes the selected notifications.
	TaskPurgeSelected tasks.TaskString = "Purge selected"

	// TaskPurgeRead removes old read notifications.
	TaskPurgeRead tasks.TaskString = "Purge read"

	// TaskClearExpiredPasswordResets removes expired or verified password resets.
	TaskClearExpiredPasswordResets tasks.TaskString = "Clear expired password resets"

	// TaskClearUserPasswordResets removes password reset requests for a user.
	TaskClearUserPasswordResets tasks.TaskString = "Clear user password resets"

	// TaskUpdate updates an existing item.
	TaskUpdate tasks.TaskString = "Update"

	// TaskBulkImport applies a bulk import action.
	TaskBulkImport tasks.TaskString = "Bulk import"

	// TaskUpdateTopicRestriction updates an existing topic restriction.
	TaskUpdateTopicRestriction tasks.TaskString = "Update topic restriction"

	// TaskUpdateUserApproval updates a writing user's approval state.
	TaskUpdateUserApproval tasks.TaskString = "Update user approval"

	// TaskUpdateRole updates an existing user role grant.
	TaskUpdateRole tasks.TaskString = "Update role"

	// TaskBulkApprove approves multiple queued items at once.
	TaskBulkApprove tasks.TaskString = "Bulk Approve"

	// TaskBulkDelete removes multiple queued items at once.
	TaskBulkDelete tasks.TaskString = "Bulk Delete"

	// TaskDeleteFilteredQueue deletes all filtered queued emails.
	TaskDeleteFilteredQueue tasks.TaskString = "Delete all filtered"

	// TaskUpdateWriting updates an existing writing.
	TaskUpdateWriting tasks.TaskString = "Update writing"

	// TaskUserAllow grants a user a role.
	TaskUserAllow tasks.TaskString = "User Allow"

	// TaskUpdatePermission modifies an existing user permission.
	TaskUpdatePermission tasks.TaskString = "Update permission"

	// TaskUserDisallow removes a user's role.
	TaskUserDisallow tasks.TaskString = "User Disallow"

	// TaskUsersAllow grants multiple users a role.
	TaskUsersAllow tasks.TaskString = "Users Allow"

	// TaskUsersDisallow removes multiple user roles.
	TaskUsersDisallow tasks.TaskString = "Users Disallow"

	// TaskUserForcePasswordChange resets a user's password.
	TaskUserForcePasswordChange tasks.TaskString = "Force Password Change"

	// TaskUserSendResetEmail sends a reset password email to the user.
	TaskUserSendResetEmail tasks.TaskString = "Send Reset Email"

	// TaskUserGenerateResetLink generates a reset password link for the user.
	TaskUserGenerateResetLink tasks.TaskString = "Generate Reset Link"

	// TaskPasswordVerify verifies a password reset code.
	TaskPasswordVerify tasks.TaskString = "Password Verify"

	// TaskUserEmailVerification verifies a user's email address.
	TaskUserEmailVerification tasks.TaskString = "Email Verification"

	// TaskWritingCategoryChange changes a writing category name.
	TaskWritingCategoryChange tasks.TaskString = "writing category change"

	// TaskWritingCategoryCreate creates a new writing category.
	TaskWritingCategoryCreate tasks.TaskString = "writing category create"

	// TaskNotify sends a custom notification to users.
	TaskNotify tasks.TaskString = "Notify"

	// TaskPurge removes old records.
	TaskPurge tasks.TaskString = "Purge"

	// TaskToggleRolePublicProfile toggles whether a role allows public profiles.
	TaskToggleRolePublicProfile tasks.TaskString = "Toggle role public profile"

	// TaskRoleGrantCreate adds a grant to a role.
	TaskRoleGrantCreate tasks.TaskString = "Create grant"

	// TaskGrantBulkCreate adds grants to multiple users or roles at once.
	TaskGrantBulkCreate tasks.TaskString = "Create grants"

	// TaskRoleGrantDelete removes a grant from a role.
	TaskRoleGrantDelete tasks.TaskString = "Delete grant"

	// TaskRoleGrantUpdate updates grants for a role.
	TaskRoleGrantUpdate tasks.TaskString = "Update grants"

	// TaskApplyRoleTemplate applies a role template.
	TaskApplyRoleTemplate tasks.TaskString = "Apply role template"

	// TaskGrantUpdateActive updates a grant's active state.
	TaskGrantUpdateActive tasks.TaskString = "Update grant"
)
