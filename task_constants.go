package main

// The following constants define the allowed values of the "task" form field.
// Each HTML form includes a hidden or submit input named "task" whose value
// identifies the intended action. When routes are registered the constants are
// passed to gorillamuxlogic's TaskMatcher so that only requests specifying the
// expected task reach a handler. Centralising these string values avoids typos
// between templates and route declarations.
const (
	// TaskAdd represents the "Add" action, commonly used when creating a new
	// record.
	TaskAdd = "Add"

	// TaskAddApproval adds an approval for a user or item.
	TaskAddApproval = "Add approval"

	// TaskAddOffsiteImage posts an image hosted elsewhere.
	TaskAddOffsiteImage = "Add offsite image"

	// TaskAnswer submits an answer in the FAQ admin interface.
	TaskAnswer = "Answer"

	// TaskApprove approves an item in moderation queues.
	TaskApprove = "Approve"

	// TaskAsk submits a new question to the FAQ system.
	TaskAsk = "Ask"

	// TaskCancel cancels the current operation and returns to the previous
	// page.
	TaskCancel = "Cancel"

	// TaskCreate indicates creation of an object, for instance a bookmark or
	// other record.
	TaskCreate = "Create"

	// TaskCreateCategory creates a new category entry.
	TaskCreateCategory = "Create Category"

	// TaskCreateLanguage creates a new language entry.
	TaskCreateLanguage = "Create Language"

	// TaskCreateThread creates a new forum thread.
	TaskCreateThread = "Create Thread"

	// TaskDelete removes an existing item.
	TaskDelete = "Delete"

	// TaskDeleteCategory removes a category.
	TaskDeleteCategory = "Delete Category"

	// TaskDeleteLanguage removes a language entry.
	TaskDeleteLanguage = "Delete Language"

	// TaskDeleteTopicRestriction deletes a topic restriction record.
	TaskDeleteTopicRestriction = "Delete topic restriction"

	// TaskDeleteUserApproval deletes a user's approval entry.
	TaskDeleteUserApproval = "Delete user approval"

	// TaskDeleteUserLevel deletes a user's access level.
	TaskDeleteUserLevel = "Delete user level"

	// TaskEdit modifies an existing item.
	TaskEdit = "Edit"

	// TaskEditReply edits a comment or reply.
	TaskEditReply = "Edit Reply"

	// TaskForumCategoryChange updates the name of a forum category.
	TaskForumCategoryChange = "Forum category change"

	// TaskForumCategoryCreate creates a new forum category.
	TaskForumCategoryCreate = "Forum category create"

	// TaskForumTopicChange updates the name of a forum topic.
	TaskForumTopicChange = "Forum topic change"

	// TaskForumTopicCreate creates a new forum topic.
	TaskForumTopicCreate = "Forum topic create"

	// TaskLogin performs a user login.
	TaskLogin = "Login"

	// TaskModifyCategory modifies an existing writing category.
	TaskModifyCategory = "Modify Category"

	// TaskModifyBoard modifies the settings of an image board.
	TaskModifyBoard = "Modify board"

	// TaskNewCategory creates a new writing category.
	TaskNewCategory = "New Category"

	// TaskNewPost creates a new news post.
	TaskNewPost = "New Post"

	// TaskNewBoard creates a new image board.
	TaskNewBoard = "New board"

	// TaskRegister registers a new user account.
	TaskRegister = "Register"

	// TaskRemakeBlogSearch rebuilds the blog search index.
	TaskRemakeBlogSearch = "Remake blog search"

	// TaskRemakeCommentsSearch rebuilds the comments search index.
	TaskRemakeCommentsSearch = "Remake comments search"

	// TaskRemakeLinkerSearch rebuilds the linker search index.
	TaskRemakeLinkerSearch = "Remake linker search"

	// TaskRemakeNewsSearch rebuilds the news search index.
	TaskRemakeNewsSearch = "Remake news search"

	// TaskRemakeStatisticInformationOnForumthread updates forum thread
	// statistics.
	TaskRemakeStatisticInformationOnForumthread = "Remake statistic information on forumthread"

	// TaskRemakeStatisticInformationOnForumtopic updates forum topic
	// statistics.
	TaskRemakeStatisticInformationOnForumtopic = "Remake statistic information on forumtopic"

	// TaskRemakeWritingSearch rebuilds the writing search index.
	TaskRemakeWritingSearch = "Remake writing search"

	// TaskRemoveRemove removes an item, typically from a list.
	TaskRemoveRemove = "Remove"

	// TaskRenameCategory renames a category.
	TaskRenameCategory = "Rename Category"

	// TaskRenameLanguage renames a language entry.
	TaskRenameLanguage = "Rename Language"

	// TaskRenameSection renames a permissions section.
	TaskRenameSection = "Rename section"

	// TaskReply posts a reply to a thread or comment.
	TaskReply = "Reply"

	// TaskSave persists changes for an item.
	TaskSave = "Save"

	// TaskSaveAll saves all changes in bulk.
	TaskSaveAll = "Save all"

	// TaskSaveLanguage saves updates to a single language.
	TaskSaveLanguage = "Save language"

	// TaskSaveLanguages saves multiple languages at once.
	TaskSaveLanguages = "Save languages"

	// TaskSearchBlogs triggers a blog search.
	TaskSearchBlogs = "Search blogs"

	// TaskSearchForum triggers a forum search.
	TaskSearchForum = "Search forum"

	// TaskSearchLinker triggers a linker search.
	TaskSearchLinker = "Search linker"

	// TaskSearchNews triggers a news search.
	TaskSearchNews = "Search news"

	// TaskSearchWritings triggers a writing search.
	TaskSearchWritings = "Search writings"

	// TaskSetTopicRestriction sets a new topic restriction.
	TaskSetTopicRestriction = "Set topic restriction"

	// TaskCopyTopicRestriction copies restriction levels from one topic to another.
	TaskCopyTopicRestriction = "Copy topic restriction"

	// TaskSetUserLevel sets a user's access level.
	TaskSetUserLevel = "Set user level"

	// TaskSubmitWriting submits a new writing.
	TaskSubmitWriting = "Submit writing"

	// TaskSuggest creates a suggestion in the linker.
	TaskSuggest = "Suggest"

	// TaskTestMail sends a test email to the current user.
	TaskTestMail = "Test mail"

	// TaskDismiss marks a notification as read.
	TaskDismiss = "Dismiss"

	// TaskUpdate updates an existing item.
	TaskUpdate = "Update"

	// TaskUpdateTopicRestriction updates an existing topic restriction.
	TaskUpdateTopicRestriction = "Update topic restriction"

	// TaskUpdateUserApproval updates a writing user's approval state.
	TaskUpdateUserApproval = "Update user approval"

	// TaskUpdateUserLevel updates a user's access level.
	TaskUpdateUserLevel = "Update user level"

	// TaskUpdateWriting updates an existing writing.
	TaskUpdateWriting = "Update writing"

	// TaskUserAllow grants a user a permission or level.
	TaskUserAllow = "User Allow"

	// TaskUserDisallow removes a user's permission or level.
	TaskUserDisallow = "User Disallow"

	// TaskUserDoNothing is used when no action should be taken on a user.
	TaskUserDoNothing = "User do nothing"

	// TaskUserResetPassword resets a user's password.
	TaskUserResetPassword = "Reset Password"

	// TaskAllow approves a news user level.
	TaskAllow = "allow"

	// TaskRemoveLower removes a news user level.
	TaskRemoveLower = "remove"

	// TaskWritingCategoryChange changes a writing category name.
	TaskWritingCategoryChange = "writing category change"

	// TaskWritingCategoryCreate creates a new writing category.
	TaskWritingCategoryCreate = "writing category create"
)
