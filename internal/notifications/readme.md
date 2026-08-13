# internal/notifications

## Purpose

Package `notifications` provides internal, non-exported utilities and service integrations specific to `notifications`.

## Why It Exists

To encapsulate the logic necessary for this specific operational domain, ensuring modularity within the codebase.

## What It Allows

It allows the system to remain decoupled. Code outside this package can rely on its exported API without worrying about its internal implementation details.

## Structure and Components

The primary files and their general responsibilities include:

- `bus_worker_test.go`
- `digest_consumer.go`
- `linker_queue_test.go`
- `notifier.go`
- `subscriptionsinterfaces.go`
- `email_test.go`
- `templates.go`
- `types.go`
- `digest_worker_test.go`
- `email.go`
- `notifications_test.go`
- `self_notify_task_test.go`
- `template_render_test.go`
- `update_email.go`
- `bus_worker.go`
- `digest_worker.go`
- `dlq.go`
- `templates_test.go`

### Exported Types and Interfaces

- **`TemplateEngine`** (Interface): Defines a core contract for this module.
- **`TestTask`**:
  - Methods: `Action`
- **`EmailTemplates`**:
- **`NotificationTemplateName`**:
  - Methods: `String`, `NotificationTemplate`, `RequiredTemplates`
- **`AdminEmailTemplateProvider`** (Interface): Defines a core contract for this module.
- **`SelfNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`NewTemplateEngine`** (Interface): Defines a core contract for this module.
- **`EmailData`**:
- **`TargetUsersNotificationProvider`** (Interface): Defines a core contract for this module.
- **`GrantsRequiredProvider`** (Interface): Defines a core contract for this module.
- **`SubscriptionTarget`** (Interface): Defines a core contract for this module.
- **`DigestType`**:
- **`SubscribersNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`DigestConsumer`**:
  - Methods: `Run`
- **`SelfEmailBroadcaster`** (Interface): Defines a core contract for this module.
- **`DirectEmailNotificationTemplateProvider`** (Interface): Defines a core contract for this module.
- **`Target`**:
  - Methods: `SubscriptionTarget`
- **`GrantRequirement`**:
- **`EmailOption`**:
- **`Notifier`**:
  - Methods: `NotifyAdmins`, `PurgeReadNotifications`, `CreateEmailTemplateAndQueue`, `RenderEmailFromTemplates`, `BusWorker`, `RegisterSync`, `ProcessEvent`, `ScheduleDigest`, `ProcessDigestForTime`, `SendDigestToUser`
- **`Option`**:
- **`EmailTemplateName`**:
  - Methods: `String`, `EmailTemplates`, `NotificationTemplate`, `RequiredTemplates`
- **`AutoSubscribeProvider`** (Interface): Defines a core contract for this module.

### Exported Functions

- `NewDigestConsumer`
- `WithSilence`
- `WithQueries`
- `WithCustomQueries`
- `WithEmailProvider`
- `WithBus`
- `WithConfig`
- `New`
- `NewEmailTemplates`
- `HTMLTemplatesNew`
- `TextTemplatesNew`
- `NotificationTemplateFilenameGenerator`
- `EmailTextTemplateFilenameGenerator`
- `EmailHTMLTemplateFilenameGenerator`
- `EmailSubjectTemplateFilenameGenerator`
- `WithAdmin`
- `WithRecipient`
- `GetUpdateEmailText`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/internal/notifications"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
