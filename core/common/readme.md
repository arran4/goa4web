# core/common

## Purpose

Package `common` contains foundational business logic and shared utilities for `common` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `opengraph_methods.go`
- `thread_sideeffects_test.go`
- `privateforum_labels_test.go`
- `role.go`
- `breadcrumb.go`
- `coredata_auth.go`
- `highlight_test.go`
- `coredata_allroles_test.go`
- `coredata_read_markers.go`
- `download_image.go`
- `link_provider_favicon_test.go`
- `permissions.go`
- `privateforum_merge.go`
- `privateforum_topic_labels_test.go`
- `coredata_images_test.go`
- `breadcrumb_private_test.go`
- `coredata_misc.go`
- `coredata_user.go`
- `funcs.go`
- `search_words.go`
- `coredata_request_test.go`
- `coredata_webauthn_test.go`
- `sectionitemtype_test.go`
- `usererror.go`
- `errors.go`
- `coredata_imagebbs.go`
- `coredata_search.go`
- `coredata_admin.go`
- `url.go`
- `breadcrumb_private_title_test.go`
- `coredata.go`
- `datacache.go`
- `highlight.go`
- `privateforum_check.go`
- `testutil_test.go`
- `search.go`
- `coredata_blogs.go`
- `link_provider.go`
- `link_provider_duration_test.go`
- `link_provider_test.go`
- `pagination.go`
- `absolute_url_test.go`
- `coredata_news.go`
- `coredata_writings.go`
- `faq.go`
- `search_words_test.go`
- `privateforum_display_title_test.go`
- `privateforum_test.go`
- `thread_sideeffects.go`
- `testutil.go`
- `encryption.go`
- `faq_test.go`
- `link_provider_tooltip_test.go`
- `privateforum.go`
- `signing.go`
- `coredata_forum.go`
- `coredata_webauthn.go`
- `jsonld.go`
- `coredata_misc_test.go`
- `coredata_labels.go`
- `download_image_test.go`

### Exported Types

- `Breadcrumb`
- `AssociateEmailParams`
- `MergeGroup`
- `CreatePrivateTopicParams`
- `PrivateTopicParticipant`
- `StoreImageParams`
- `UserError`
- `ImageBBSPoster`
- `ImageBBSBoard`
- `ImageBBSThread`
- `ImageBBSThreadPosts`
- `LanguageCache`
- `IndexGroup`
- `IndexItem`
- `AdminSection`
- `OpenGraph`
- `NotFoundLink`
- `SessionManager`
- `MailProvider`
- `NavigationProvider`
- `CoreData`
- `CoreOption`
- `LatestWritingsOption`
- `DataCache`
- `PrivateForumInconsistency`
- `Goa4WebLinkProvider`
- `MockQuerier`
- `PageLink`
- `Pagination`
- `OffsetPagination`
- `PageNumberPagination`
- `ThreadInfo`
- `FAQ`
- `CategoryFAQs`
- `CreateFAQQuestionParams`
- `ThreadUpdatedEvent`
- `QuerierFake`
- `PrivateTopic`
- `WebAuthnUser`
- `JSONLDer`
- `Person`
- `Organization`
- `Article`
- `NewsArticle`
- `BlogPosting`
- `DiscussionForumPosting`

### Exported Functions

- `TestHandleThreadUpdatedMarksThreadAndItemLabels`
- `TestCoreData_PrivateForumTopics_LabelsBug`
- `TestCoreData_PrivateForumTopics_UnreadNew`
- `TestCoreData_PrivateForumTopics_OwnThreadNotNew`
- `Allowed`
- `TestHighlightSearchTermsEscapesAndHighlights`
- `TestHighlightSearchTermsRespectsWordBoundaries`
- `TestHighlightSearchTermsWithoutWordsEscapesHTML`
- `TestAllRolesLazy`
- `TestRenderLink_Favicon`
- `TestCoreData_PrivateForumTopics_ShowsTopicLabels`
- `TestCreateCommentValidatesGalleryImages`
- `TestMapImageURLUsesDefaultThumbnailForLargeUploadedImage`
- `TestMapImageURLUsesDefaultThumbnailForLargeCachedImage`
- `TestMapImageURLUsesThumbnailForCachedImageWithoutMetadata`
- `TestRecordUploadedImageThumbnailLinksSourceImage`
- `TestStoreImageRecordsDefaultThumbnail`
- `TestSanitizeCodeImagesQueuesImageAliasGoogleRedirect`
- `TestPrivateForumBreadcrumbBasePath`
- `GetTemplateFuncs`
- `A4Code2String`
- `TopicTitleOrDefault`
- `TopicDescriptionOrDefault`
- `FirstLine`
- `Left`
- `TruncateWords`
- `ToString`
- `ToInt32`
- `Add`
- `Seq`
- `Dict`
- `ToJSON`
- `TimeAgo`
- `Since`
- `TestLoadSelectionsFromRequest`
- `TestBackupEligiblePasskeyPersistence`
- `TestSectionItemType`
- `TestPrivateForumBreadcrumbUsesDisplayTitle`
- `NewLanguageCache`
- `WithImageURLMapper`
- `WithLanguageCache`
- `WithHTTPClient`
- `WithSession`
- `WithSessionManager`
- `WithEvent`
- `WithEventBus`
- `WithAbsoluteURLBase`
- `WithPreference`
- `WithUserRoles`
- `WithPermissions`
- `WithGrants`
- `WithConfig`
- `WithSiteTitle`
- `WithWebAuthn`
- `WithImageSignKey`
- `WithShareSignKey`
- `WithLinkSignKey`
- `WithFeedSignKey`
- `WithTasksRegistry`
- `WithDLQRegistry`
- `WithDBRegistry`
- `WithEmailRegistry`
- `WithNavRegistry`
- `WithRouterModules`
- `WithCustomQueries`
- `WithOffset`
- `NewCoreData`
- `WithEmailProvider`
- `ContainsItem`
- `WithWritingsOffset`
- `WithWritingsLimit`
- `WithSilence`
- `WithTrustedProxies`
- `HighlightSearchTerms`
- `TestQuerierFakeGrantStubs`
- `TestQuerierFakeTopicListing`
- `CanSearch`
- `NewGoa4WebLinkProvider`
- `TestFormatDuration`
- `TestRenderLink_RoutesThroughGoto`
- `TestAbsoluteURL`
- `TestSearchWordsFromRequestCachesAndReturnsCopy`
- `TestGetPrivateTopicDisplayTitle`
- `TestCoreData_PrivateForumTopics`
- `NewTestCoreData`
- `TestAllAnsweredFAQ_Categories`
- `TestRenderLink_Tooltips`
- `WithPrivateForumTopics`
- `UnmarshalJSONLD`
- `TestCreatePrivateTopicUsesProvidedUsernames`
- `TestCreatePrivateTopicBuildsUsernamesWhenMissing`
- `TestMapLinkURL`
- `TestDownloadAndCacheImageRecordsRemoteMetadata`
- `TestPrepareImageCacheEntryForServeRefreshesExpiredRemoteEntry`
- `TestQueueRemoteImageCacheCreatesPendingEntry`
- `TestDownloadExternalImageUsesOpenGraphImageFromHTML`
- `TestCreateCommentStartsImmediateRemoteImageFetch`
- `TestProcessPendingRemoteImageCacheEntriesRecordsRetryFailure`
- `TestPrepareImageCacheEntryForServeAllowsMissingMetadataWhenExpiryDisabled`
- `TestPrepareImageCacheEntryForServeRejectsExpiredRemoteEntryWithoutSource`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/common"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
