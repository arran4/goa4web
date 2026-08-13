# core/templates

## Purpose

Package `templates` contains foundational business logic and shared utilities for `templates` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `news_post_page_test.go`
- `notification_open_template_test.go`
- `tableTopics_tags_test.go`
- `templates.go`
- `asset_hash_test.go`
- `no_empty_templates_test.go`
- `passkeys_asset_test.go`
- `verification.go`
- `thread_new_page_test.go`
- `label.go`
- `threadPage_labels_test.go`
- `extract.go`
- `thread_page_test.go`
- `article_page_test.go`
- `comment_test.go`

### Exported Types and Interfaces

- **`Option`** (Interface): Defines a core contract for this module.
- **`MissingImageData`**:
- **`MockUser`**:
- **`MockComment`**:
- **`MockCoreData`**:
  - Methods: `LocalTime`
- **`TopicLabel`**:
- **`TemplateSet`**:

### Exported Functions

- `TestNewsPostPageLabelFormIncludesCSRF`
- `TestNewsPostPageReplyFormIncludesCSRF`
- `TestNewsPostPageDoesNotContainInlineMarkRead`
- `TestNotificationOpenTemplateExists`
- `TestTableTopicsShowsLabels`
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
- `TestGetAssetHash`
- `TestNoEmptyTemplates`
- `TestPasskeysJavaScriptIsExternal`
- `LoadAllTemplatesMap`
- `IsTemplateAvailable`
- `TestThreadNewPageJS`
- `TestThreadPageShowsDefaultPrivateLabels`
- `WriteToDir`
- `WriteTemplateSetsToDir`
- `ArchiveTemplates`
- `TestThreadPageLabelFormIncludesCSRF`
- `TestThreadPageDoesNotContainInlineMarkRead`
- `TestArticlePageLabelFormIncludesCSRF`
- `TestArticlePageReplyFormIncludesCSRF`
- `TestArticlePageDoesNotContainInlineMarkRead`
- `TestCommentTimestampSelfLink`
- `TestCommentUsernameBold`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/templates"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
