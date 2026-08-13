# internal/notifications

## Purpose

Package `notifications` provides internal, non-exported utilities and service integrations specific to `notifications`.

## Context and Use Cases (How and Why)

**Why it exists:** To encapsulate the logic necessary for this specific operational domain, ensuring modularity.
**What this allows:** It allows the system to remain decoupled. Code outside this package can rely on its exported API without worrying about its internal implementation details.
**How to use it:** Import the package and call its exported functions or instantiate its public interfaces.

## Structure and Components

The primary files and their general responsibilities include:

- `templates.go`
- `types.go`
- `digest_worker.go`
- `email_test.go`
- `self_notify_task_test.go`
- `bus_worker.go`
- `bus_worker_test.go`
- `digest_worker_test.go`
- `templates_test.go`
- `update_email.go`
- `digest_consumer.go`
- `notifications_test.go`
- `notifier.go`
- `dlq.go`
- `email.go`
- `linker_queue_test.go`
- `subscriptionsinterfaces.go`
- `template_render_test.go`

### Exported Types and Interfaces

- **`NewTemplateEngine`** (Interface): Defines a core contract for this module.
- **`DigestType`**:
- **`Notifier`**:
  - Methods: `ScheduleDigest`, `ProcessDigestForTime`, `SendDigestToUser`, `BusWorker`, `RegisterSync`, `ProcessEvent`, `NotifyAdmins`, `PurgeReadNotifications`, `CreateEmailTemplateAndQueue`, `RenderEmailFromTemplates`
- **`NotificationTemplateName`**:
  - Methods: `String`, `NotificationTemplate`, `RequiredTemplates`
- **`GrantsRequiredProvider`** (Interface): Defines a core contract for this module.
- **`SubscriptionTarget`** (Interface): Defines a core contract for this module.
- **`GrantRequirement`**:
- **`EmailTemplates`**:
- **`EmailTemplateName`**:
  - Methods: `String`, `EmailTemplates`, `NotificationTemplate`, `RequiredTemplates`
- **`AdminEmailTemplateProvider`** (Interface): Defines a core contract for this module.
- **`SelfNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`SubscribersNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`Target`**:
  - Methods: `SubscriptionTarget`
- **`EmailData`**:
- **`EmailOption`**:
- **`SelfEmailBroadcaster`** (Interface): Defines a core contract for this module.
- **`AutoSubscribeProvider`** (Interface): Defines a core contract for this module.
- **`TargetUsersNotificationProvider`** (Interface): Defines a core contract for this module.
- **`TemplateEngine`** (Interface): Defines a core contract for this module.
- **`TestTask`**:
  - Methods: `Action`
- **`DigestConsumer`**:
  - Methods: `Run`
- **`Option`**:
- **`DirectEmailNotificationTemplateProvider`** (Interface): Defines a core contract for this module.

### Exported Functions

- `HTMLTemplatesNew`
- `TextTemplatesNew`
- `NotificationTemplateFilenameGenerator`
- `EmailTextTemplateFilenameGenerator`
- `EmailHTMLTemplateFilenameGenerator`
- `EmailSubjectTemplateFilenameGenerator`
- `TestRenderEmailFromTemplates_AdminSubject`
- `TestBuildPatterns`
- `TestBuildPatternsAdditional`
- `TestExpandPatternSeparators`
- `TestCollectSubscribersQuery`
- `TestProcessEventDLQ`
- `TestProcessEventSubscribeSelf`
- `TestProcessEventNoAutoSubscribe`
- `TestProcessEventAdminNotify`
- `TestProcessEventWritingSubscribers`
- `TestProcessEventTargetUsers`
- `TestNotifySubscribersNews`
- `TestBusWorker`
- `TestProcessEventAutoSubscribe`
- `TestProcessEventAutoSubscribeMissingPreference`
- `TestProcessEventSelfNotifyWithUserIDTemplate`
- `TestNotificationDigestWorker_ScheduleDigest`
- `TestNotificationDigestWorker_SendDigest`
- `TestRenderNotificationUsesSequentialOverrides`
- `GetUpdateEmailText`
- `NewDigestConsumer`
- `TestNotificationsQueries`
- `TestNotifierNotifyAdmins`
- `TestNotifierInitialization`
- `WithSilence`
- `WithQueries`
- `WithCustomQueries`
- `WithEmailProvider`
- `WithBus`
- `WithConfig`
- `New`
- `WithAdmin`
- `WithRecipient`
- `TestLinkerQueueNotifierMessages`
- `NewEmailTemplates`
- `TestAdminNotificationTemplate`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/notifications"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
