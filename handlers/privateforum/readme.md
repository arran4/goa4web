# handlers/privateforum

## Purpose

Package `privateforum` handles HTTP requests for the `privateforum` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

### Exported Types and Interfaces

- **`UserExistsResponse`**:
- **`PrivateTopicCreateTask`**:
  - Methods: `Action`, `AutoSubscribePath`, `AutoSubscribeGrants`
- **`QuerierProxier`**:
  - Methods: `SystemGetUserByUsername`

### Exported Functions

- `APIListTopics`
- `APICreateTopic`
- `APIListThreads`
- `APIShowComments`
- `APIPostComment`
- `TopicRssPage`
- `TopicAtomPage`
- `TopicFeedHandler`
- `DisablePrivateForumCaching`
- `EnforcePrivateForumTopicSeeAccess`
- `TestHappyPathPrivateForumTasksTemplatesRequiredExist`
- `TestPrivateRoute`
- `TestHappyPathTopicPage_Prefix`
- `UserExistsAPI`
- `TestStartGroupDiscussionPage_RouterGrantFailure`
- `TestHappyPathPrivateTopicCreateTaskAutoSubscribe`
- `TestHappyPathPrivateTopicCreateTaskAutoSubscribePath`
- `TestPrivateLabelRoutes`
- `RegisterTasks`
- `TopicEditPage`
- `TopicEditSubmit`
- `RegisterRoutes`
- `Register`
- `SharedThreadPreviewPage`
- `SharedTopicPreviewPage`
- `TopicCancelAlias`
- `ThreadPage`
- `TestDisablePrivateForumCaching`
- `TestHappyPathPagesExist`
- `UnreadThreadsPage`
- `TestUnhappyPathPage_NoAccess`
- `TestHappyPathPage_Access`
- `TestHappyPathPage_SeeNoCreate`
- `TestHappyPathPage_AdminLinks`
- `NewPrivateForumTask`
- `TopicPage`
- `TestUserExistsAPI`
- `PrivateForumPage`
- `StartGroupDiscussionPage`

## Usage

Handlers are registered during server initialization. They are not typically called directly by other Go code. To add a new endpoint, implement an `http.HandlerFunc` or implement `tasks.Task` for the admin framework, and map it in the router initialization.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.
