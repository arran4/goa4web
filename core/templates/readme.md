# core/templates

## Purpose

Package `templates` contains foundational business logic and shared utilities for `templates` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `comment_test.go`
- `news_post_page_test.go`
- `threadPage_labels_test.go`
- `thread_new_page_test.go`
- `verification.go`
- `article_page_test.go`
- `label.go`
- `notification_open_template_test.go`
- `tableTopics_tags_test.go`
- `no_empty_templates_test.go`
- `templates.go`
- `asset_hash_test.go`
- `extract.go`
- `passkeys_asset_test.go`
- `thread_page_test.go`

### Exported Types

- `MockUser`
- `MockComment`
- `MockCoreData`
- `TopicLabel`
- `Option`
- `MissingImageData`
- `TemplateSet`

### Exported Functions

- `TestCommentTimestampSelfLink`
- `TestCommentUsernameBold`
- `TestNewsPostPageLabelFormIncludesCSRF`
- `TestNewsPostPageReplyFormIncludesCSRF`
- `TestNewsPostPageDoesNotContainInlineMarkRead`
- `TestThreadPageShowsDefaultPrivateLabels`
- `TestThreadNewPageJS`
- `LoadAllTemplatesMap`
- `IsTemplateAvailable`
- `TestArticlePageLabelFormIncludesCSRF`
- `TestArticlePageReplyFormIncludesCSRF`
- `TestArticlePageDoesNotContainInlineMarkRead`
- `TestNotificationOpenTemplateExists`
- `TestTableTopicsShowsLabels`
- `TestNoEmptyTemplates`
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
- `WriteToDir`
- `WriteTemplateSetsToDir`
- `ArchiveTemplates`
- `TestPasskeysJavaScriptIsExternal`
- `TestThreadPageLabelFormIncludesCSRF`
- `TestThreadPageDoesNotContainInlineMarkRead`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/templates"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
