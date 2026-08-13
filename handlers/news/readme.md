# handlers/news

## Purpose

Package `news` handles HTTP requests for the `news` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

### Exported Types and Interfaces

- **`DeleteNewsPostTask`**:
  - Methods: `Action`
- **`AnnouncementAddTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`
- **`AnnouncementDeleteTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`
- **`EditTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Page`, `Action`
- **`CancelTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`
- **`NewPostTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `SubscribedEmailTemplate`, `SubscribedInternalNotificationTemplate`, `RequiredTemplates`, `AutoSubscribePath`, `AutoSubscribeGrants`, `Action`
- **`ReplyTask`**:
  - Methods: `IndexType`, `IndexData`, `SubscribedEmailTemplate`, `SubscribedInternalNotificationTemplate`, `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `AutoSubscribePath`, `AutoSubscribeGrants`, `Action`
- **`EditReplyTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`
- **`UserDisallowTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`
- **`NewsTestStub`**:
  - Methods: `GetForumThreadIdByNewsPostId`
- **`UserAllowTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`
- **`SetLabelsTask`**:
  - Methods: `Action`
- **`MarkReadTask`**:
  - Methods: `Matcher`, `Action`

### Exported Functions

- `TestMarkReadTaskAjax`
- `NewNewsTask`
- `RegisterAdminRoutes`
- `TestNewsPostPageLabelBars`
- `TestNewsPostPagePrivateLabelsOnce`
- `TestMarkReadTaskRedirect`
- `EnforceNewsPostAccess`
- `NewNewsPostTask`
- `NewsRssPage`
- `RegisterRoutes`
- `Register`
- `TestHappyPathEditTaskTemplatesRequiredExist`
- `TestCustomNewsIndexRoles`
- `SearchResultNewsActionPage`
- `SharedPreviewPage`
- `RegisterTasks`
- `AdminNewsPage`
- `AdminNewsPostPage`
- `AdminNewsDeleteConfirmPage`
- `TestNewsReply`
- `TestHappyPathEditRouteRegistered`
- `MatchCanPostNews`
- `TestPreviewRoute`
- `TestPreviewHandler`
- `TestHappyPathNewsTasksTemplatesRequiredExist`
- `TestNewsSearchFiltersUnauthorized`
- `TestNewsCustomIndexItems_FetchAuthor`
- `CanEditNewsPost`
- `CanPostNews`
- `TestHappyPathNewsAutoSubscribeTasks`
- `TestNewsPostNewActionPage_InvalidForms`
- `TestNewsPostEditActionPage_InvalidForms`
- `PreviewPage`
- `TestHappyPathNewsReplyAutoSubscribe`
- `TestHappyPathNewsPostAutoSubscribe`
- `NewsGeneralIndexItems`
- `NewsPageSpecificItems`
- `NewsCustomIndexItems`
- `NewsPageHandler`
- `CustomNewsIndex`
- `TestNewsRssPage`
- `TestHappyPathNewsTemplatesExist`
- `NewsCreatePageHandler`
- `TestNewsListingDismissLink`
- `NewsPostPageHandler`

## Usage

Handlers are registered during server initialization. They are not typically called directly by other Go code. To add a new endpoint, implement an `http.HandlerFunc` or implement `tasks.Task` for the admin framework, and map it in the router initialization.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.
