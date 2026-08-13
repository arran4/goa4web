# internal/notifications

## Purpose

Package `notifications` provides internal, non-exported utilities and service integrations specific to `notifications`.

## Structure and Components

The primary files and their general responsibilities include:

- `digest_worker.go`
- `notifier.go`
- `self_notify_task_test.go`
- `subscriptionsinterfaces.go`
- `bus_worker.go`
- `digest_consumer.go`
- `email.go`
- `linker_queue_test.go`
- `templates_test.go`
- `update_email.go`
- `bus_worker_test.go`
- `dlq.go`
- `templates.go`
- `types.go`
- `digest_worker_test.go`
- `email_test.go`
- `notifications_test.go`
- `template_render_test.go`

### Exported Types and Interfaces

- **`Notifier`**:
  - Methods: `ScheduleDigest`, `ProcessDigestForTime`, `SendDigestToUser`, `NotifyAdmins`, `PurgeReadNotifications`, `BusWorker`, `RegisterSync`, `ProcessEvent`, `CreateEmailTemplateAndQueue`, `RenderEmailFromTemplates`
- **`EmailTemplates`**:
- **`EmailTemplateName`**:
  - Methods: `String`, `EmailTemplates`, `NotificationTemplate`, `RequiredTemplates`
- **`TestTask`**:
  - Methods: `Action`
- **`SubscriptionTarget`** (Interface): Defines a core contract for this module.
- **`NotificationTemplateName`**:
  - Methods: `String`, `NotificationTemplate`, `RequiredTemplates`
- **`SelfNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`DirectEmailNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`DigestConsumer`**:
  - Methods: `Run`
- **`DigestType`**:
- **`SelfEmailBroadcaster`** (Interface): Defines a core contract for this module.
- **`SubscribersNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`AutoSubscribeProvider`** (Interface): Defines a core contract for this module.
- **`TargetUsersNotificationProvider`** (Interface): Defines a core contract for this module.
- **`EmailData`**:
- **`EmailOption`**:
- **`TemplateEngine`** (Interface): Defines a core contract for this module.
- **`Option`**:
- **`AdminEmailTemplateProvider`** (Interface): Defines a core contract for this module.
- **`GrantsRequiredProvider`** (Interface): Defines a core contract for this module.
- **`NewTemplateEngine`** (Interface): Defines a core contract for this module.
- **`Target`**:
  - Methods: `SubscriptionTarget`
- **`GrantRequirement`**:

### Exported Functions

- `WithSilence`
- `WithQueries`
- `WithCustomQueries`
- `WithEmailProvider`
- `WithBus`
- `WithConfig`
- `New`
- `NewEmailTemplates`
- `NewDigestConsumer`
- `WithAdmin`
- `WithRecipient`
- `TestLinkerQueueNotifierMessages`
- `TestRenderNotificationUsesSequentialOverrides`
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
- `HTMLTemplatesNew`
- `TextTemplatesNew`
- `NotificationTemplateFilenameGenerator`
- `EmailTextTemplateFilenameGenerator`
- `EmailHTMLTemplateFilenameGenerator`
- `EmailSubjectTemplateFilenameGenerator`
- `TestNotificationDigestWorker_ScheduleDigest`
- `TestNotificationDigestWorker_SendDigest`
- `TestRenderEmailFromTemplates_AdminSubject`
- `TestNotificationsQueries`
- `TestNotifierNotifyAdmins`
- `TestNotifierInitialization`
- `TestAdminNotificationTemplate`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/notifications"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
