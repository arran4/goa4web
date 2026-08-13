# core/templates

## Purpose

Package `templates` contains foundational business logic and shared utilities for `templates` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `notification_open_template_test.go`
- `article_page_test.go`
- `verification.go`
- `tableTopics_tags_test.go`
- `templates.go`
- `thread_page_test.go`
- `asset_hash_test.go`
- `comment_test.go`
- `extract.go`
- `news_post_page_test.go`
- `passkeys_asset_test.go`
- `threadPage_labels_test.go`
- `thread_new_page_test.go`
- `label.go`
- `no_empty_templates_test.go`

### Exported Types and Interfaces

- **`TemplateSet`**:
- **`MockUser`**:
- **`MockComment`**:
- **`MockCoreData`**:
  - Methods: `LocalTime`
- **`TopicLabel`**:
- **`Option`** (Interface): Defines a core contract for this module.
- **`MissingImageData`**:

### Exported Functions

- `TestNotificationOpenTemplateExists`
- `TestArticlePageLabelFormIncludesCSRF`
- `TestArticlePageReplyFormIncludesCSRF`
- `TestArticlePageDoesNotContainInlineMarkRead`
- `LoadAllTemplatesMap`
- `IsTemplateAvailable`
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
- `TestThreadPageLabelFormIncludesCSRF`
- `TestThreadPageDoesNotContainInlineMarkRead`
- `TestGetAssetHash`
- `TestCommentTimestampSelfLink`
- `TestCommentUsernameBold`
- `WriteToDir`
- `WriteTemplateSetsToDir`
- `ArchiveTemplates`
- `TestNewsPostPageLabelFormIncludesCSRF`
- `TestNewsPostPageReplyFormIncludesCSRF`
- `TestNewsPostPageDoesNotContainInlineMarkRead`
- `TestPasskeysJavaScriptIsExternal`
- `TestThreadPageShowsDefaultPrivateLabels`
- `TestThreadNewPageJS`
- `TestNoEmptyTemplates`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/templates"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
