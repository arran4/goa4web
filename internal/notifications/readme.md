# internal/notifications

## Purpose

Package `notifications` provides internal, non-exported utilities and service integrations specific to `notifications`.

## Structure and Components

The primary files and their general responsibilities include:

- `subscriptionsinterfaces.go`
- `templates.go`
- `digest_worker.go`
- `digest_worker_test.go`
- `dlq.go`
- `email.go`
- `notifier.go`
- `self_notify_task_test.go`
- `template_render_test.go`
- `templates_test.go`
- `digest_consumer.go`
- `notifications_test.go`
- `types.go`
- `bus_worker.go`
- `update_email.go`
- `bus_worker_test.go`
- `email_test.go`
- `linker_queue_test.go`

### Exported Types and Interfaces

- **`GrantsRequiredProvider`** (Interface): Defines a core contract for this module.
- **`EmailData`**:
- **`EmailOption`**:
- **`Option`**:
- **`TestTask`**:
  - Methods: `Action`
- **`EmailTemplates`**:
- **`SelfNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`SelfEmailBroadcaster`** (Interface): Defines a core contract for this module.
- **`TargetUsersNotificationProvider`** (Interface): Defines a core contract for this module.
- **`DigestType`**:
- **`GrantRequirement`**:
- **`NotificationTemplateName`**:
  - Methods: `String`, `NotificationTemplate`, `RequiredTemplates`
- **`AutoSubscribeProvider`** (Interface): Defines a core contract for this module.
- **`TemplateEngine`** (Interface): Defines a core contract for this module.
- **`NewTemplateEngine`** (Interface): Defines a core contract for this module.
- **`Notifier`**:
  - Methods: `ScheduleDigest`, `ProcessDigestForTime`, `SendDigestToUser`, `CreateEmailTemplateAndQueue`, `RenderEmailFromTemplates`, `NotifyAdmins`, `PurgeReadNotifications`, `BusWorker`, `RegisterSync`, `ProcessEvent`
- **`Target`**:
  - Methods: `SubscriptionTarget`
- **`EmailTemplateName`**:
  - Methods: `String`, `EmailTemplates`, `NotificationTemplate`, `RequiredTemplates`
- **`DirectEmailNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`DigestConsumer`**:
  - Methods: `Run`
- **`SubscriptionTarget`** (Interface): Defines a core contract for this module.
- **`AdminEmailTemplateProvider`** (Interface): Defines a core contract for this module.
- **`SubscribersNotificationTemplateProvider`** (Interface): Defines a core contract for this module.

### Exported Functions

- `NewEmailTemplates`
- `HTMLTemplatesNew`
- `TextTemplatesNew`
- `NotificationTemplateFilenameGenerator`
- `EmailTextTemplateFilenameGenerator`
- `EmailHTMLTemplateFilenameGenerator`
- `EmailSubjectTemplateFilenameGenerator`
- `TestNotificationDigestWorker_ScheduleDigest`
- `TestNotificationDigestWorker_SendDigest`
- `WithAdmin`
- `WithRecipient`
- `WithSilence`
- `WithQueries`
- `WithCustomQueries`
- `WithEmailProvider`
- `WithBus`
- `WithConfig`
- `New`
- `TestAdminNotificationTemplate`
- `TestRenderNotificationUsesSequentialOverrides`
- `NewDigestConsumer`
- `TestNotificationsQueries`
- `TestNotifierNotifyAdmins`
- `TestNotifierInitialization`
- `GetUpdateEmailText`
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
- `TestRenderEmailFromTemplates_AdminSubject`
- `TestLinkerQueueNotifierMessages`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/notifications"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
