# internal/notifications

## Purpose

Package `notifications` provides internal, non-exported utilities and service integrations specific to `notifications`.

## Structure and Components

The primary files and their general responsibilities include:

- `digest_worker_test.go`
- `linker_queue_test.go`
- `templates.go`
- `update_email.go`
- `subscriptionsinterfaces.go`
- `notifier.go`
- `self_notify_task_test.go`
- `digest_worker.go`
- `dlq.go`
- `types.go`
- `template_render_test.go`
- `bus_worker.go`
- `bus_worker_test.go`
- `digest_consumer.go`
- `email.go`
- `email_test.go`
- `notifications_test.go`
- `templates_test.go`

### Exported Types

- `TemplateEngine`
- `NewTemplateEngine`
- `EmailTemplates`
- `EmailTemplateName`
- `NotificationTemplateName`
- `AdminEmailTemplateProvider`
- `SelfNotificationTemplateProvider`
- `SelfEmailBroadcaster`
- `DirectEmailNotificationTemplateProvider`
- `SubscribersNotificationTemplateProvider`
- `AutoSubscribeProvider`
- `TargetUsersNotificationProvider`
- `GrantsRequiredProvider`
- `Notifier`
- `Option`
- `DigestType`
- `SubscriptionTarget`
- `Target`
- `GrantRequirement`
- `TestTask`
- `DigestConsumer`
- `EmailData`
- `EmailOption`

### Exported Functions

- `TestNotificationDigestWorker_ScheduleDigest`
- `TestNotificationDigestWorker_SendDigest`
- `TestLinkerQueueNotifierMessages`
- `HTMLTemplatesNew`
- `TextTemplatesNew`
- `NotificationTemplateFilenameGenerator`
- `EmailTextTemplateFilenameGenerator`
- `EmailHTMLTemplateFilenameGenerator`
- `EmailSubjectTemplateFilenameGenerator`
- `GetUpdateEmailText`
- `NewEmailTemplates`
- `WithSilence`
- `WithQueries`
- `WithCustomQueries`
- `WithEmailProvider`
- `WithBus`
- `WithConfig`
- `New`
- `TestAdminNotificationTemplate`
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
- `NewDigestConsumer`
- `WithAdmin`
- `WithRecipient`
- `TestRenderEmailFromTemplates_AdminSubject`
- `TestNotificationsQueries`
- `TestNotifierNotifyAdmins`
- `TestNotifierInitialization`
- `TestRenderNotificationUsesSequentialOverrides`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/notifications"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
