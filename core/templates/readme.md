# core/templates

## Purpose

Package `templates` contains foundational business logic and shared utilities for `templates` that are used application-wide.

## Context and Use Cases (How and Why)

**Why it exists:** To house logic, constants, and utilities that are required universally across handlers, workers, and internal services.
**What this allows:** It prevents code duplication. For example, `CoreData` is defined here and passed everywhere to provide unified access to the database and configuration.
**How to use it:** Import the `core/*` package and invoke its exported utilities. Avoid adding dependencies from `core` to higher-level packages like `handlers` to prevent import cycles.

## Structure and Components

The primary files and their general responsibilities include:

- `tableTopics_tags_test.go`
- `thread_page_test.go`
- `verification.go`
- `article_page_test.go`
- `comment_test.go`
- `templates.go`
- `extract.go`
- `label.go`
- `no_empty_templates_test.go`
- `notification_open_template_test.go`
- `asset_hash_test.go`
- `news_post_page_test.go`
- `passkeys_asset_test.go`
- `threadPage_labels_test.go`
- `thread_new_page_test.go`

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

- `TestTableTopicsShowsLabels`
- `TestThreadPageLabelFormIncludesCSRF`
- `TestThreadPageDoesNotContainInlineMarkRead`
- `LoadAllTemplatesMap`
- `IsTemplateAvailable`
- `TestArticlePageLabelFormIncludesCSRF`
- `TestArticlePageReplyFormIncludesCSRF`
- `TestArticlePageDoesNotContainInlineMarkRead`
- `TestCommentTimestampSelfLink`
- `TestCommentUsernameBold`
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
- `WriteToDir`
- `WriteTemplateSetsToDir`
- `ArchiveTemplates`
- `TestNoEmptyTemplates`
- `TestNotificationOpenTemplateExists`
- `TestGetAssetHash`
- `TestNewsPostPageLabelFormIncludesCSRF`
- `TestNewsPostPageReplyFormIncludesCSRF`
- `TestNewsPostPageDoesNotContainInlineMarkRead`
- `TestPasskeysJavaScriptIsExternal`
- `TestThreadPageShowsDefaultPrivateLabels`
- `TestThreadNewPageJS`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/templates"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
