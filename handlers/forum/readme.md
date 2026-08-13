# handlers/forum

## Purpose

Package `forum` handles HTTP requests for the `forum` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

### Exported Types and Interfaces

- **`ThreadDeleteTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`AddTopicPublicLabelTask`**:
  - Methods: `Action`
- **`TopicGrantGroup`**:
- **`RemoveAuthorLabelTask`**:
  - Methods: `Action`
- **`ForumcategoryPlus`**:
- **`RemakeThreadStatsTask`**:
- **`TopicGrantDeleteTask`**:
  - Methods: `Action`
- **`TopicDeleteTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`CategoryCreateTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`RemovePublicLabelTask`**:
  - Methods: `Action`
- **`CategoryTree`**:
  - Methods: `PruneEmpty`, `CategoryRoots`
- **`RemakeTopicStatsTask`**:
- **`RemoveTopicPublicLabelTask`**:
  - Methods: `Action`
- **`MarkThreadReadTask`**:
  - Methods: `Matcher`, `Action`
- **`SetLabelsTask`**:
  - Methods: `Action`
- **`DeleteCategoryTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`TopicChangeTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`TopicGrantUpdateTask`**:
  - Methods: `Action`
- **`AdminTopicDisplay`**:
- **`AddAuthorLabelTask`**:
  - Methods: `Action`
- **`ForumtopicPlus`**:
- **`CategoryChangeTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`CategoryGrantCreateTask`**:
  - Methods: `Action`
- **`TopicCreateTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`
- **`AddPrivateLabelTask`**:
  - Methods: `Action`
- **`TopicGrantCreateTask`**:
  - Methods: `Action`
- **`MockEmailProvider`**:
  - Methods: `Send`, `TestConfig`
- **`ReplyTask`**:
  - Methods: `IndexType`, `IndexData`, `SubscribedEmailTemplate`, `SubscribedInternalNotificationTemplate`, `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `AutoSubscribePath`, `AutoSubscribeGrants`, `Action`
- **`CreateThreadTask`**:
  - Methods: `IndexType`, `IndexData`, `SubscribedEmailTemplate`, `SubscribedInternalNotificationTemplate`, `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `AutoSubscribePath`, `AutoSubscribeGrants`, `Page`, `Action`
- **`Handlers`**:
- **`SetTopicLabelsTask`**:
  - Methods: `Action`
- **`CreateTopicPageForm`**:
- **`AddPublicLabelTask`**:
  - Methods: `Action`
- **`CategoryGrantDeleteTask`**:
  - Methods: `Action`
- **`RemovePrivateLabelTask`**:
  - Methods: `Action`

### Exported Functions

- `AdminCategoryEditPage`
- `AdminCategoryEditSubmit`
- `AdminCategoryDeletePage`
- `TestRequireThreadAndTopicTrue`
- `TestRequireThreadAndTopicFalse`
- `TestRequireThreadAndTopicError`
- `TestCategoryRoute`
- `SubscribeTopicPage`
- `TestThreadDelete`
- `TestHappyPathAdminTopicEditTemplateDeleteTaskValue`
- `AdminCategoryCreatePage`
- `AdminCategoryCreateSubmit`
- `TestForumTemplatesExist`
- `TestHappyPathCreateThreadLabels`
- `TestHappyPathForumReply`
- `TestForumAutoSubscribeTasks`
- `QuoteApi`
- `QuoteSelectionApi`
- `TopicFeed`
- `TopicRssPage`
- `TopicAtomPage`
- `TopicsPageWithBasePath`
- `TopicsPage`
- `AdminForumFlaggedPostsPage`
- `BasePathMiddleware`
- `AdminTopicPage`
- `AdminTopicEditFormPage`
- `UserCanCreateThread`
- `UserCanCreateTopic`
- `UserCanLabelTopic`
- `ThreadDelete`
- `UnsubscribeTopicPage`
- `CustomForumIndex`
- `ForumCustomIndexItems`
- `TestCustomForumIndexWriteReply`
- `TestCustomForumIndexMarkReadLinks`
- `TestCustomForumIndexHidesMarkReadWhenClear`
- `TestCustomForumIndexWriteReplyDenied`
- `TestCustomForumIndexCreateThread`
- `TestCustomForumIndexAdminEditLink`
- `TestCustomForumIndexCreateThreadDenied`
- `TestCustomForumIndexSubscribeLink`
- `TestCustomForumIndexUnsubscribeLink`
- `ThreadPageWithBasePath`
- `ThreadPage`
- `TestForumTopicFeed`
- `TestHappyPathCreateThreadNotificationLink`
- `TestSubscribeTopicTaskAction`
- `TestUnsubscribeTopicTaskAction`
- `TestPagesExist`
- `RegisterRoutes`
- `Register`
- `RegisterAdminRoutes`
- `TestForumAdminTemplatesExist`
- `TopicThreadReplyCancelPage`
- `TestMarkThreadReadTaskRedirect`
- `TestMarkThreadReadTaskRefererFallback`
- `TestSetLabelsTaskAddsInverseLabels`
- `TestSetLabelsTaskUpdatesSpecialLabels`
- `TestMarkThreadReadTaskRedirectWithThread`
- `TestBuildTopicGrantGroupsIncludesAllRoles`
- `TestTopicThreadReplyCancel_BasePath`
- `TestAdminCategoryCreateSubmitSuccess`
- `TestAdminCategoryCreateSubmitValidationError`
- `TestAdminCategoryEditSubmitSuccess`
- `TestAdminCategoryEditSubmitMissingCategory`
- `TestAdminCategoryEditSubmitValidationError`
- `TestSharedTopicPreviewPage_GuestRedirectsToLogin`
- `ManageTopicLabelsPage`
- `TestHappyPathCategoryTreePruneEmpty`
- `CreateTopicPageWithPostTask`
- `AdminThreadsPage`
- `AdminThreadDeletePage`
- `AdminThreadDeleteConfirmPage`
- `AdminThreadPage`
- `ThreadNewCancelPage`
- `TopicThreadCommentEditActionPage`
- `TopicThreadCommentEditActionCancelPage`
- `SharedThreadPreviewPage`
- `SharedTopicPreviewPage`
- `TestForumPageHandlers`
- `TestThreadPageTitle`
- `TestTopicPageTitle`
- `RequireThreadAndTopic`
- `AdminForumModeratorLogsPage`
- `AdminForumWordListPage`
- `RegisterTasks`
- `TestQuoteApi`
- `TestQuoteSelectionApiGroupsConsecutiveSelectionsByUser`
- `AdminCategoryPage`
- `AdminTopicGrantsPage`
- `New`
- `TestManageTopicLabelsPage`
- `TestCustomForumIndex_Author_NewStatus`
- `AdminCategoryGrantsPage`
- `Page`
- `AdminTopicsPage`
- `AdminTopicEditPage`
- `AdminTopicCreatePage`
- `AdminTopicDeleteConfirmPage`
- `AdminTopicDeletePage`
- `TestHappyPathForumReplyRedirect`
- `TestUserCanCreateThread_Allowed`
- `TestUserCanCreateThread_Denied`
- `TestUserCanCreateThread_Error`
- `TestUserCanCreateTopic_Allowed`
- `TestUserCanCreateTopic_Denied`
- `TestUserCanCreateTopic_Error`
- `AdminForumPage`
- `AdminForumRemakeForumThreadPage`
- `AdminForumRemakeForumTopicPage`
- `NewCategoryTree`
- `NewCategoryTreeUnpruned`
- `TestHappyPathCreateThreadTaskAutoSubscribeGrants`
- `TestHappyPathReplyTaskAutoSubscribeGrants`
- `AdminCategoriesPage`
- `TestForumReplyErrorRetainsText`

## Usage

Handlers are registered during server initialization. They are not typically called directly by other Go code. To add a new endpoint, implement an `http.HandlerFunc` or implement `tasks.Task` for the admin framework, and map it in the router initialization.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.
