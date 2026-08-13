# core/templates

## Purpose

Package `templates` contains foundational business logic and shared utilities for `templates` that are used application-wide.

## Why It Exists

To house logic, constants, and utilities that are required universally across handlers, workers, and internal services.

## What It Allows

It prevents code duplication. For example, `CoreData` is defined here and passed everywhere to provide unified access to the database and configuration state.

## Structure and Components

The primary files and their general responsibilities include:

- `passkeys_asset_test.go`
- `news_post_page_test.go`
- `templates.go`
- `verification.go`
- `article_page_test.go`
- `extract.go`
- `label.go`
- `no_empty_templates_test.go`
- `threadPage_labels_test.go`
- `asset_hash_test.go`
- `tableTopics_tags_test.go`
- `thread_new_page_test.go`
- `thread_page_test.go`
- `comment_test.go`
- `notification_open_template_test.go`

### Exported Types and Interfaces

- **`MockComment`**:
- **`MockCoreData`**:
  - Methods: `LocalTime`
- **`Option`** (Interface): Defines a core contract for this module.
- **`MissingImageData`**:
- **`TemplateSet`**:
- **`TopicLabel`**:
- **`MockUser`**:

### Exported Functions

- `SetDir`
- `WithDir`
- `WithSilence`
- `Asset`
- `GetAssetHash`
- `GetCompiledSiteTemplates`
- `GetCompiledNotificationTemplates`
- `GetCompiledEmailHtmlTemplates`
- `GetCompiledEmailTextTemplates`
- `GetAssetData`
- `GetMissingImageData`
- `GetMissingImageSVG`
- `ListSiteTemplateNames`
- `TemplateExists`
- `EmailTemplateExists`
- `NotificationTemplateExists`
- `AnyTemplateExists`
- `LoadAllTemplatesMap`
- `IsTemplateAvailable`
- `WriteToDir`
- `WriteTemplateSetsToDir`
- `ArchiveTemplates`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/core/templates"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
