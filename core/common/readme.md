# core/common

## Purpose

Package `common` contains foundational business logic and shared utilities for `common` that are used application-wide.

## Context and Use Cases (How and Why)

**Why it exists:** To house logic, constants, and utilities that are required universally across handlers, workers, and internal services.
**What this allows:** It prevents code duplication. For example, `CoreData` is defined here and passed everywhere to provide unified access to the database and configuration.
**How to use it:** Import the `core/*` package and invoke its exported utilities. Avoid adding dependencies from `core` to higher-level packages like `handlers` to prevent import cycles.

## Structure and Components

The primary files and their general responsibilities include:

- `coredata_auth.go`
- `privateforum.go`
- `thread_sideeffects_test.go`
- `coredata_imagebbs.go`
- `coredata_user.go`
- `download_image_test.go`
- `privateforum_check.go`
- `download_image.go`
- `link_provider_test.go`
- `role.go`
- `coredata_forum.go`
- `coredata_news.go`
- `highlight_test.go`
- `link_provider_tooltip_test.go`
- `coredata_labels.go`
- `link_provider.go`
- `privateforum_test.go`
- `search.go`
- `absolute_url_test.go`
- `coredata_admin.go`
- `datacache.go`
- `coredata_blogs.go`
- `coredata_misc.go`
- `coredata_request_test.go`
- `jsonld.go`
- `link_provider_favicon_test.go`
- `thread_sideeffects.go`
- `breadcrumb_private_title_test.go`
- `coredata_search.go`
- `signing.go`
- `coredata_images_test.go`
- `coredata_webauthn.go`
- `funcs.go`
- `url.go`
- `usererror.go`
- `encryption.go`
- `permissions.go`
- `privateforum_display_title_test.go`
- `privateforum_topic_labels_test.go`
- `testutil.go`
- `breadcrumb_private_test.go`
- `coredata_allroles_test.go`
- `coredata_misc_test.go`
- `coredata_read_markers.go`
- `coredata_webauthn_test.go`
- `coredata_writings.go`
- `errors.go`
- `link_provider_duration_test.go`
- `coredata.go`
- `faq_test.go`
- `opengraph_methods.go`
- `privateforum_labels_test.go`
- `privateforum_merge.go`
- `search_words_test.go`
- `testutil_test.go`
- `faq.go`
- `highlight.go`
- `pagination.go`
- `search_words.go`
- `sectionitemtype_test.go`
- `breadcrumb.go`

### Exported Types and Interfaces

- **`CreatePrivateTopicParams`**:
- **`PrivateTopicParticipant`**:
- **`StoreImageParams`**:
- **`JSONLDer`** (Interface): Defines a core contract for this module.
- **`LanguageCache`**:
  - Methods: `Load`, `Invalidate`
- **`IndexItem`**:
- **`NavigationProvider`** (Interface): Defines a core contract for this module.
- **`PageLink`**:
- **`AssociateEmailParams`**:
- **`CoreData`**:
  - Methods: `UserCredentials`, `VerifiedEmailsForUser`, `AssociateEmail`, `UserExists`, `CreateUserWithEmail`, `CreatePasswordReset`, `CreatePasswordResetForUser`, `VerifyPasswordReset`, `GetPrivateTopicDetails`, `GetPrivateTopicDisplayTitle`, `GetPrivateTopicParticipants`, `PrivateForumTopics`, `PrivateTopics`, `GrantPrivateForumTopic`, `UnreadPrivateThreads`, `UnreadPrivateThreadsCount`, `ImageBBSFeed`, `ImageBBSPoster`, `ImageBBSBoard`, `ImageBBSThread`, `ImageBBSThreadPosts`, `UserSettings`, `UserLanguages`, `UserEmails`, `AddUserEmail`, `SaveEmail`, `DeleteEmail`, `AddEmail`, `UserGallery`, `PublicProfile`, `PagedUsers`, `UserNotifications`, `DeleteSubscription`, `UpdateSubscriptions`, `SetUserLanguage`, `SetUserLanguages`, `SetTimezone`, `SaveProfile`, `UpdatePermissions`, `AllowPermission`, `DisallowPermission`, `SaveNotificationDigestPreferences`, `CheckAndFixPrivateForumInconsistencies`, `DownloadAndCacheImage`, `QueueRemoteImageCache`, `StartRemoteImageCacheFetch`, `ProcessPendingRemoteImageCacheEntries`, `RecordUploadedImageThumbnail`, `RecordUploadedImageDerivative`, `RecordCachedImageThumbnail`, `RecordDerivedImageCacheEntry`, `PrepareImageCacheEntryForServe`, `ImageCacheEntry`, `ForumCategories`, `ForumCategory`, `ForumThreadByID`, `ForumThread`, `ForumThreads`, `ForumTopicByID`, `ForumTopics`, `ForumThreadReplies`, `ForumComment`, `UpdateForumComment`, `EditForumComment`, `SubscribeTopic`, `UnsubscribeTopic`, `SubscribeThread`, `UnsubscribeThread`, `GrantForumCategory`, `RevokeForumCategory`, `GrantForumTopic`, `RevokeForumTopic`, `GrantForumThread`, `RevokeForumThread`, `ThreadInfo`, `CreateNewsReply`, `UpdateNewsReply`, `UpdateNewsPost`, `DeleteNewsPost`, `CreateNewsPost`, `SearchNews`, `AllowNewsUser`, `DisallowNewsUser`, `AddAnnouncement`, `DeleteAnnouncement`, `SystemGetNewsPost`, `PublicLabels`, `AddPublicLabel`, `RemovePublicLabel`, `AddAuthorLabel`, `RemoveAuthorLabel`, `SetAuthorLabels`, `SetPublicLabels`, `PrivateLabels`, `ClearPrivateLabelStatus`, `ClearUnreadForOthers`, `SetPrivateLabelStatus`, `AddPrivateLabel`, `RemovePrivateLabel`, `SetPrivateLabels`, `TopicPublicLabels`, `AddTopicPublicLabel`, `RemoveTopicPublicLabel`, `SetTopicPublicLabels`, `TopicPrivateLabels`, `AddTopicPrivateLabel`, `RemoveTopicPrivateLabel`, `SetTopicPrivateLabels`, `ThreadPublicLabels`, `AddThreadPublicLabel`, `RemoveThreadPublicLabel`, `AddThreadAuthorLabel`, `RemoveThreadAuthorLabel`, `SetThreadAuthorLabels`, `SetThreadPublicLabels`, `ThreadPrivateLabels`, `ClearThreadPrivateLabelStatus`, `ClearThreadUnreadForOthers`, `SetThreadPrivateLabelStatus`, `AddThreadPrivateLabel`, `RemoveThreadPrivateLabel`, `SetThreadPrivateLabels`, `WritingAuthorLabels`, `AddWritingAuthorLabel`, `RemoveWritingAuthorLabel`, `SetWritingAuthorLabels`, `WritingPrivateLabels`, `SetWritingPrivateLabels`, `ClearWritingUnreadForOthers`, `WritingLabels`, `NewsAuthorLabels`, `AddNewsAuthorLabel`, `RemoveNewsAuthorLabel`, `SetNewsAuthorLabels`, `NewsPrivateLabels`, `SetNewsPrivateLabels`, `NewsLabels`, `BlogAuthorLabels`, `AddBlogAuthorLabel`, `RemoveBlogAuthorLabel`, `SetBlogAuthorLabels`, `BlogPrivateLabels`, `SetBlogPrivateLabels`, `BlogLabels`, `CanSearch`, `AdminListUsers`, `AdminUserPendingPasswordResetCounts`, `AdminDashboardStats`, `AdminCommentsByUser`, `AdminListPasswordResets`, `AdminApprovePasswordReset`, `AdminDenyPasswordReset`, `BlogPost`, `BlogComments`, `BlogCategories`, `EditableBlogPost`, `BlogCommentThread`, `CreateBlogReply`, `UpdateBlogReply`, `BloggerProfile`, `BloggerPosts`, `AllBlogs`, `CreatePrivateTopic`, `UploadedImageByImageID`, `StoreImage`, `StoreSystemImage`, `HandleThreadUpdated`, `SearchLinker`, `SearchWritings`, `SearchBlogs`, `SearchForum`, `SearchComments`, `SearchCommentsNoResults`, `SearchCommentsEmptyWords`, `SearchLinkerItems`, `SearchLinkerNoResults`, `SearchLinkerEmptyWords`, `SearchWritingsResults`, `SearchWritingsNoResults`, `SearchWritingsEmptyWords`, `SearchBlogsResults`, `SearchBlogsNoResults`, `SearchBlogsEmptyWords`, `SignShareURL`, `SignShareURLQuery`, `SignImageURL`, `SignCacheURL`, `SignLinkURL`, `SignFeedURL`, `MapImageURL`, `MapFullImageURL`, `ThumbnailReferenceForImage`, `ThumbnailReferenceForCache`, `MapLinkURL`, `GetWebAuthnUser`, `GetWebAuthnUserByID`, `SavePasskey`, `UpdatePasskeyAfterLogin`, `Funcs`, `IsAllowedHost`, `SanitizeBackURL`, `EncryptData`, `DecryptData`, `HasGrant`, `SetThreadReadMarker`, `ThreadReadMarker`, `Article`, `ArticleComments`, `WritingCategories`, `EditableArticle`, `ArticleComment`, `UpdateArticleComment`, `WriterByUsername`, `WriterWritings`, `UpdateWritingReply`, `CreateWritingReply`, `UpdateWriting`, `CreateWriting`, `GrantWritingCategory`, `RevokeWritingCategory`, `CreateWritingCategory`, `ChangeWritingCategory`, `AbsoluteURL`, `AdminForumTopics`, `AdminLatestNews`, `AdminLatestNewsList`, `AdminLoginAttempts`, `AdminSessions`, `AdminLinkerItemByID`, `AllRoles`, `RoleByID`, `SelectedRole`, `Announcement`, `AnnouncementLoaded`, `ArchivedRequests`, `BlogEntryByID`, `Bloggers`, `BlogList`, `BlogListForSelectedAuthor`, `Bookmarks`, `CreateBookmark`, `SaveBookmark`, `IsAdmin`, `IsAdminMode`, `CanEditBlog`, `ShowReplyNews`, `ShowEditNews`, `CommentByID`, `CurrentBlog`, `CurrentBlogLoaded`, `CurrentComment`, `CurrentCommentLoaded`, `CurrentNewsPost`, `CurrentNewsPostLoaded`, `CurrentProfileBookmarkSize`, `CurrentProfileComments`, `CurrentProfileEmails`, `CurrentProfileGrants`, `CurrentProfileRoles`, `CurrentProfileStats`, `CurrentProfileUser`, `CurrentRequest`, `CurrentRequestComments`, `CurrentRequestUser`, `CurrentTopic`, `CurrentTopicLoaded`, `CurrentUser`, `CurrentUserLoaded`, `CurrentWriting`, `CurrentWritingLoaded`, `CustomQueries`, `DBRegistry`, `EmailRegistry`, `DefaultNotificationTemplate`, `EmailProvider`, `HTTPClient`, `Event`, `Publish`, `ExecuteSiteTemplate`, `ExternalLink`, `FAQCategories`, `HasAdminRole`, `HasContentWriterRole`, `HasRole`, `HasSubscription`, `ImageBoardPosts`, `ImageBoards`, `ImagePostByID`, `ImageURLMapper`, `Languages`, `RenameLanguage`, `DeleteLanguage`, `CreateLanguage`, `LatestNews`, `LatestNewsList`, `LatestWritings`, `LinkerCategories`, `LinkerCategoriesForUser`, `LinkerCategoryByID`, `LinkerCategoryCounts`, `CreateFAQCategory`, `LinkerItemsForUser`, `LinkerLinksByCategoryID`, `Marked`, `NewsAnnouncement`, `NewsAnnouncementWithErr`, `NewsPostByID`, `Notifications`, `PageSize`, `PendingRequests`, `Permissions`, `Preference`, `PreferredLanguageID`, `Location`, `LocalTime`, `FormatLocalTime`, `LocalTimeIn`, `FormatLocalTimeIn`, `PublicWritings`, `Queries`, `SelectedQuestionFromCategory`, `UpdateFAQQuestion`, `DeleteFAQCategory`, `DeleteFAQQuestion`, `RegisterExternalLinkClick`, `Role`, `SelectedAdminLinkerItem`, `SelectedAdminLinkerItemID`, `SelectedBoardPosts`, `SelectedBoardSubBoards`, `SelectedCategoryPublicWritings`, `SelectedLinkerCategory`, `SelectedLinkerItem`, `SelectedLinkerItemsForCurrentUser`, `SelectedThread`, `SelectedThreadComments`, `SelectedSectionThreadComments`, `SelectedThreadLoaded`, `SelectedThreadCanReply`, `SelectedNewsThreadCanReply`, `SelectedForumThreadCanReply`, `SelectedPrivateForumThreadCanReply`, `SelectedBlogThreadCanReply`, `SelectedImageBBSThreadCanReply`, `SelectedWritingThreadCanReply`, `SelectedLinkerThreadCanReply`, `CreateCommentInSectionForCommenter`, `CreateNewsCommentForCommenter`, `CreateForumCommentForCommenter`, `CreatePrivateForumCommentForCommenter`, `CreateBlogCommentForCommenter`, `CreateImageBBSCommentForCommenter`, `CreateWritingCommentForCommenter`, `CreateLinkerCommentForCommenter`, `CanEditComment`, `CommentEditing`, `CommentEditURL`, `CommentEditSaveURL`, `CommentAdminURL`, `SelectedCommentID`, `Session`, `GetSession`, `SessionManager`, `SetCurrentBlog`, `SetCurrentNewsPost`, `SetCurrentProfileUserID`, `CurrentProfileUserID`, `SetCurrentRequestID`, `CurrentRequestID`, `Offset`, `SetCurrentRoleID`, `SetCurrentSection`, `Section`, `SetCurrentNotificationTemplate`, `SetCurrentError`, `SetCurrentNotice`, `SetCurrentThreadAndTopic`, `SetCurrentWriting`, `SetCurrentExternalLinkID`, `SelectedExternalLink`, `EnsureExternalLink`, `GetExternalLink`, `UpdateExternalLinkMetadata`, `UpdateExternalLinkImageCache`, `SelectedBoardID`, `SelectedThreadID`, `SelectedImagePostID`, `SelectedRoleID`, `SetEvent`, `SetEventTask`, `SetSession`, `SubImageBoards`, `Subscribed`, `Subscriptions`, `CurrentError`, `CustomCSS`, `CurrentNotice`, `NotificationTemplateError`, `NotificationTemplateName`, `NotificationTemplateOverride`, `ThreadComments`, `SectionThreadComments`, `UnreadNotificationCount`, `UserByID`, `UserRoles`, `UserSubscriptions`, `VisibleWritingCategories`, `Writers`, `WritingByID`, `SortedCustomIndexGroups`, `SortedCustomIndexItems`, `LoadSelectionsFromRequest`, `HasModule`, `GenerateFeedURL`, `ValidateCodeImagesForUser`, `ValidateCodeImagesForThread`, `RecordThreadImages`, `ResolveExternalLink`, `ExternalLinkTargetURL`, `ExternalLinkRedirectURL`, `ExternalLinkReloadURL`, `MergePrivateTopicsWithSameParticipants`, `AllAnsweredFAQ`, `UpdateFAQCategory`, `CreateFAQQuestion`, `SearchWords`, `Breadcrumbs`
- **`PrivateForumInconsistency`**:
- **`UserError`**:
  - Methods: `Error`, `UserErrorMessage`, `Unwrap`
- **`CoreOption`**:
- **`PageNumberPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`Breadcrumb`**:
- **`ImageBBSPoster`**:
- **`NotFoundLink`**:
- **`DiscussionForumPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`ThreadUpdatedEvent`**:
- **`Goa4WebLinkProvider`**:
  - Methods: `MapImageURL`, `RenderLink`
- **`DataCache`**:
- **`ImageBBSThread`**:
- **`ImageBBSThreadPosts`**:
- **`Person`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`Article`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`OpenGraph`**:
  - Methods: `URLMeta`, `ImageMeta`, `SecureImageMeta`, `ImageWidthMeta`, `ImageHeightMeta`, `TwitterImageMeta`, `TypeMeta`, `ExpirationTimeMeta`, `PublishedTimeMeta`, `ModifiedTimeMeta`, `SiteNameMeta`, `UpdatedTimeMeta`, `JSONLDScript`
- **`SessionManager`** (Interface): Defines a core contract for this module.
- **`LatestWritingsOption`**:
- **`Pagination`** (Interface): Defines a core contract for this module.
- **`ImageBBSBoard`**:
- **`ThreadInfo`**:
- **`BlogPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`WebAuthnUser`**:
  - Methods: `WebAuthnID`, `WebAuthnName`, `WebAuthnDisplayName`, `WebAuthnIcon`, `WebAuthnCredentials`, `EstablishLegacyCredentialFlags`
- **`FAQ`**:
- **`CreateFAQQuestionParams`**:
- **`PrivateTopic`**:
- **`IndexGroup`**:
- **`AdminSection`**:
- **`MailProvider`** (Interface): Defines a core contract for this module.
- **`CategoryFAQs`**:
- **`NewsArticle`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`QuerierFake`**:
  - Methods: `SystemCheckGrant`, `SystemCheckRoleGrant`, `AdminListTopicsWithUserGrantsNoRoles`
- **`MockQuerier`**:
  - Methods: `GetExternalLink`
- **`Organization`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`MergeGroup`**:
- **`OffsetPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`

### Exported Functions

- `WithPrivateForumTopics`
- `TestHandleThreadUpdatedMarksThreadAndItemLabels`
- `TestDownloadAndCacheImageRecordsRemoteMetadata`
- `TestPrepareImageCacheEntryForServeRefreshesExpiredRemoteEntry`
- `TestQueueRemoteImageCacheCreatesPendingEntry`
- `TestDownloadExternalImageUsesOpenGraphImageFromHTML`
- `TestCreateCommentStartsImmediateRemoteImageFetch`
- `TestProcessPendingRemoteImageCacheEntriesRecordsRetryFailure`
- `TestPrepareImageCacheEntryForServeAllowsMissingMetadataWhenExpiryDisabled`
- `TestPrepareImageCacheEntryForServeRejectsExpiredRemoteEntryWithoutSource`
- `TestRenderLink_RoutesThroughGoto`
- `Allowed`
- `TestHighlightSearchTermsEscapesAndHighlights`
- `TestHighlightSearchTermsRespectsWordBoundaries`
- `TestHighlightSearchTermsWithoutWordsEscapesHTML`
- `TestRenderLink_Tooltips`
- `NewGoa4WebLinkProvider`
- `TestCoreData_PrivateForumTopics`
- `CanSearch`
- `TestAbsoluteURL`
- `TestLoadSelectionsFromRequest`
- `UnmarshalJSONLD`
- `TestRenderLink_Favicon`
- `TestPrivateForumBreadcrumbUsesDisplayTitle`
- `TestCreateCommentValidatesGalleryImages`
- `TestMapImageURLUsesDefaultThumbnailForLargeUploadedImage`
- `TestMapImageURLUsesDefaultThumbnailForLargeCachedImage`
- `TestMapImageURLUsesThumbnailForCachedImageWithoutMetadata`
- `TestRecordUploadedImageThumbnailLinksSourceImage`
- `TestStoreImageRecordsDefaultThumbnail`
- `TestSanitizeCodeImagesQueuesImageAliasGoogleRedirect`
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
- `TestGetPrivateTopicDisplayTitle`
- `TestCoreData_PrivateForumTopics_ShowsTopicLabels`
- `NewTestCoreData`
- `TestPrivateForumBreadcrumbBasePath`
- `TestAllRolesLazy`
- `TestCreatePrivateTopicUsesProvidedUsernames`
- `TestCreatePrivateTopicBuildsUsernamesWhenMissing`
- `TestMapLinkURL`
- `TestBackupEligiblePasskeyPersistence`
- `TestFormatDuration`
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
- `TestAllAnsweredFAQ_Categories`
- `TestCoreData_PrivateForumTopics_LabelsBug`
- `TestCoreData_PrivateForumTopics_UnreadNew`
- `TestCoreData_PrivateForumTopics_OwnThreadNotNew`
- `TestSearchWordsFromRequestCachesAndReturnsCopy`
- `TestQuerierFakeGrantStubs`
- `TestQuerierFakeTopicListing`
- `HighlightSearchTerms`
- `TestSectionItemType`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/common"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
