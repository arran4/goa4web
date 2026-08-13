# core/common

## Purpose

Package `common` contains foundational business logic and shared utilities for `common` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `signing.go`
- `breadcrumb_private_test.go`
- `coredata.go`
- `coredata_user.go`
- `highlight_test.go`
- `breadcrumb_private_title_test.go`
- `coredata_admin.go`
- `download_image_test.go`
- `privateforum_labels_test.go`
- `search.go`
- `coredata_request_test.go`
- `privateforum_check.go`
- `coredata_imagebbs.go`
- `coredata_misc_test.go`
- `download_image.go`
- `privateforum_display_title_test.go`
- `sectionitemtype_test.go`
- `encryption.go`
- `coredata_auth.go`
- `coredata_misc.go`
- `coredata_webauthn.go`
- `coredata_webauthn_test.go`
- `opengraph_methods.go`
- `privateforum_test.go`
- `role.go`
- `coredata_labels.go`
- `coredata_news.go`
- `privateforum_merge.go`
- `thread_sideeffects.go`
- `thread_sideeffects_test.go`
- `usererror.go`
- `breadcrumb.go`
- `coredata_images_test.go`
- `coredata_read_markers.go`
- `testutil.go`
- `coredata_allroles_test.go`
- `link_provider.go`
- `link_provider_tooltip_test.go`
- `search_words_test.go`
- `absolute_url_test.go`
- `coredata_search.go`
- `highlight.go`
- `link_provider_duration_test.go`
- `url.go`
- `link_provider_test.go`
- `pagination.go`
- `privateforum.go`
- `coredata_forum.go`
- `faq.go`
- `search_words.go`
- `testutil_test.go`
- `coredata_writings.go`
- `funcs.go`
- `jsonld.go`
- `faq_test.go`
- `link_provider_favicon_test.go`
- `datacache.go`
- `permissions.go`
- `coredata_blogs.go`
- `errors.go`
- `privateforum_topic_labels_test.go`

### Exported Types and Interfaces

- **`LanguageCache`**:
  - Methods: `Load`, `Invalidate`
- **`AdminSection`**:
- **`SessionManager`** (Interface): Defines a core contract for this module.
- **`LatestWritingsOption`**:
- **`CreatePrivateTopicParams`**:
- **`ThreadUpdatedEvent`**:
- **`UserError`**:
  - Methods: `Error`, `UserErrorMessage`, `Unwrap`
- **`CategoryFAQs`**:
- **`CoreData`**:
  - Methods: `SignShareURL`, `SignShareURLQuery`, `SignImageURL`, `SignCacheURL`, `SignLinkURL`, `SignFeedURL`, `MapImageURL`, `MapFullImageURL`, `ThumbnailReferenceForImage`, `ThumbnailReferenceForCache`, `MapLinkURL`, `AbsoluteURL`, `AdminForumTopics`, `AdminLatestNews`, `AdminLatestNewsList`, `AdminLoginAttempts`, `AdminSessions`, `AdminLinkerItemByID`, `AllRoles`, `RoleByID`, `SelectedRole`, `Announcement`, `AnnouncementLoaded`, `ArchivedRequests`, `BlogEntryByID`, `Bloggers`, `BlogList`, `BlogListForSelectedAuthor`, `Bookmarks`, `CreateBookmark`, `SaveBookmark`, `IsAdmin`, `IsAdminMode`, `CanEditBlog`, `ShowReplyNews`, `ShowEditNews`, `CommentByID`, `CurrentBlog`, `CurrentBlogLoaded`, `CurrentComment`, `CurrentCommentLoaded`, `CurrentNewsPost`, `CurrentNewsPostLoaded`, `CurrentProfileBookmarkSize`, `CurrentProfileComments`, `CurrentProfileEmails`, `CurrentProfileGrants`, `CurrentProfileRoles`, `CurrentProfileStats`, `CurrentProfileUser`, `CurrentRequest`, `CurrentRequestComments`, `CurrentRequestUser`, `CurrentTopic`, `CurrentTopicLoaded`, `CurrentUser`, `CurrentUserLoaded`, `CurrentWriting`, `CurrentWritingLoaded`, `CustomQueries`, `DBRegistry`, `EmailRegistry`, `DefaultNotificationTemplate`, `EmailProvider`, `HTTPClient`, `Event`, `Publish`, `ExecuteSiteTemplate`, `ExternalLink`, `FAQCategories`, `HasAdminRole`, `HasContentWriterRole`, `HasRole`, `HasSubscription`, `ImageBoardPosts`, `ImageBoards`, `ImagePostByID`, `ImageURLMapper`, `Languages`, `RenameLanguage`, `DeleteLanguage`, `CreateLanguage`, `LatestNews`, `LatestNewsList`, `LatestWritings`, `LinkerCategories`, `LinkerCategoriesForUser`, `LinkerCategoryByID`, `LinkerCategoryCounts`, `CreateFAQCategory`, `LinkerItemsForUser`, `LinkerLinksByCategoryID`, `Marked`, `NewsAnnouncement`, `NewsAnnouncementWithErr`, `NewsPostByID`, `Notifications`, `PageSize`, `PendingRequests`, `Permissions`, `Preference`, `PreferredLanguageID`, `Location`, `LocalTime`, `FormatLocalTime`, `LocalTimeIn`, `FormatLocalTimeIn`, `PublicWritings`, `Queries`, `SelectedQuestionFromCategory`, `UpdateFAQQuestion`, `DeleteFAQCategory`, `DeleteFAQQuestion`, `RegisterExternalLinkClick`, `Role`, `SelectedAdminLinkerItem`, `SelectedAdminLinkerItemID`, `SelectedBoardPosts`, `SelectedBoardSubBoards`, `SelectedCategoryPublicWritings`, `SelectedLinkerCategory`, `SelectedLinkerItem`, `SelectedLinkerItemsForCurrentUser`, `SelectedThread`, `SelectedThreadComments`, `SelectedSectionThreadComments`, `SelectedThreadLoaded`, `SelectedThreadCanReply`, `SelectedNewsThreadCanReply`, `SelectedForumThreadCanReply`, `SelectedPrivateForumThreadCanReply`, `SelectedBlogThreadCanReply`, `SelectedImageBBSThreadCanReply`, `SelectedWritingThreadCanReply`, `SelectedLinkerThreadCanReply`, `CreateCommentInSectionForCommenter`, `CreateNewsCommentForCommenter`, `CreateForumCommentForCommenter`, `CreatePrivateForumCommentForCommenter`, `CreateBlogCommentForCommenter`, `CreateImageBBSCommentForCommenter`, `CreateWritingCommentForCommenter`, `CreateLinkerCommentForCommenter`, `CanEditComment`, `CommentEditing`, `CommentEditURL`, `CommentEditSaveURL`, `CommentAdminURL`, `SelectedCommentID`, `Session`, `GetSession`, `SessionManager`, `SetCurrentBlog`, `SetCurrentNewsPost`, `SetCurrentProfileUserID`, `CurrentProfileUserID`, `SetCurrentRequestID`, `CurrentRequestID`, `Offset`, `SetCurrentRoleID`, `SetCurrentSection`, `Section`, `SetCurrentNotificationTemplate`, `SetCurrentError`, `SetCurrentNotice`, `SetCurrentThreadAndTopic`, `SetCurrentWriting`, `SetCurrentExternalLinkID`, `SelectedExternalLink`, `EnsureExternalLink`, `GetExternalLink`, `UpdateExternalLinkMetadata`, `UpdateExternalLinkImageCache`, `SelectedBoardID`, `SelectedThreadID`, `SelectedImagePostID`, `SelectedRoleID`, `SetEvent`, `SetEventTask`, `SetSession`, `SubImageBoards`, `Subscribed`, `Subscriptions`, `CurrentError`, `CustomCSS`, `CurrentNotice`, `NotificationTemplateError`, `NotificationTemplateName`, `NotificationTemplateOverride`, `ThreadComments`, `SectionThreadComments`, `UnreadNotificationCount`, `UserByID`, `UserRoles`, `UserSubscriptions`, `VisibleWritingCategories`, `Writers`, `WritingByID`, `SortedCustomIndexGroups`, `SortedCustomIndexItems`, `LoadSelectionsFromRequest`, `HasModule`, `GenerateFeedURL`, `ValidateCodeImagesForUser`, `ValidateCodeImagesForThread`, `RecordThreadImages`, `ResolveExternalLink`, `ExternalLinkTargetURL`, `ExternalLinkRedirectURL`, `ExternalLinkReloadURL`, `UserSettings`, `UserLanguages`, `UserEmails`, `AddUserEmail`, `SaveEmail`, `DeleteEmail`, `AddEmail`, `UserGallery`, `PublicProfile`, `PagedUsers`, `UserNotifications`, `DeleteSubscription`, `UpdateSubscriptions`, `SetUserLanguage`, `SetUserLanguages`, `SetTimezone`, `SaveProfile`, `UpdatePermissions`, `AllowPermission`, `DisallowPermission`, `SaveNotificationDigestPreferences`, `AdminListUsers`, `AdminUserPendingPasswordResetCounts`, `AdminDashboardStats`, `AdminCommentsByUser`, `AdminListPasswordResets`, `AdminApprovePasswordReset`, `AdminDenyPasswordReset`, `CanSearch`, `CheckAndFixPrivateForumInconsistencies`, `ImageBBSFeed`, `ImageBBSPoster`, `ImageBBSBoard`, `ImageBBSThread`, `ImageBBSThreadPosts`, `DownloadAndCacheImage`, `QueueRemoteImageCache`, `StartRemoteImageCacheFetch`, `ProcessPendingRemoteImageCacheEntries`, `RecordUploadedImageThumbnail`, `RecordUploadedImageDerivative`, `RecordCachedImageThumbnail`, `RecordDerivedImageCacheEntry`, `PrepareImageCacheEntryForServe`, `ImageCacheEntry`, `EncryptData`, `DecryptData`, `UserCredentials`, `VerifiedEmailsForUser`, `AssociateEmail`, `UserExists`, `CreateUserWithEmail`, `CreatePasswordReset`, `CreatePasswordResetForUser`, `VerifyPasswordReset`, `CreatePrivateTopic`, `UploadedImageByImageID`, `StoreImage`, `StoreSystemImage`, `GetWebAuthnUser`, `GetWebAuthnUserByID`, `SavePasskey`, `UpdatePasskeyAfterLogin`, `PublicLabels`, `AddPublicLabel`, `RemovePublicLabel`, `AddAuthorLabel`, `RemoveAuthorLabel`, `SetAuthorLabels`, `SetPublicLabels`, `PrivateLabels`, `ClearPrivateLabelStatus`, `ClearUnreadForOthers`, `SetPrivateLabelStatus`, `AddPrivateLabel`, `RemovePrivateLabel`, `SetPrivateLabels`, `TopicPublicLabels`, `AddTopicPublicLabel`, `RemoveTopicPublicLabel`, `SetTopicPublicLabels`, `TopicPrivateLabels`, `AddTopicPrivateLabel`, `RemoveTopicPrivateLabel`, `SetTopicPrivateLabels`, `ThreadPublicLabels`, `AddThreadPublicLabel`, `RemoveThreadPublicLabel`, `AddThreadAuthorLabel`, `RemoveThreadAuthorLabel`, `SetThreadAuthorLabels`, `SetThreadPublicLabels`, `ThreadPrivateLabels`, `ClearThreadPrivateLabelStatus`, `ClearThreadUnreadForOthers`, `SetThreadPrivateLabelStatus`, `AddThreadPrivateLabel`, `RemoveThreadPrivateLabel`, `SetThreadPrivateLabels`, `WritingAuthorLabels`, `AddWritingAuthorLabel`, `RemoveWritingAuthorLabel`, `SetWritingAuthorLabels`, `WritingPrivateLabels`, `SetWritingPrivateLabels`, `ClearWritingUnreadForOthers`, `WritingLabels`, `NewsAuthorLabels`, `AddNewsAuthorLabel`, `RemoveNewsAuthorLabel`, `SetNewsAuthorLabels`, `NewsPrivateLabels`, `SetNewsPrivateLabels`, `NewsLabels`, `BlogAuthorLabels`, `AddBlogAuthorLabel`, `RemoveBlogAuthorLabel`, `SetBlogAuthorLabels`, `BlogPrivateLabels`, `SetBlogPrivateLabels`, `BlogLabels`, `ThreadInfo`, `CreateNewsReply`, `UpdateNewsReply`, `UpdateNewsPost`, `DeleteNewsPost`, `CreateNewsPost`, `SearchNews`, `AllowNewsUser`, `DisallowNewsUser`, `AddAnnouncement`, `DeleteAnnouncement`, `SystemGetNewsPost`, `MergePrivateTopicsWithSameParticipants`, `HandleThreadUpdated`, `Breadcrumbs`, `SetThreadReadMarker`, `ThreadReadMarker`, `SearchLinker`, `SearchWritings`, `SearchBlogs`, `SearchForum`, `SearchComments`, `SearchCommentsNoResults`, `SearchCommentsEmptyWords`, `SearchLinkerItems`, `SearchLinkerNoResults`, `SearchLinkerEmptyWords`, `SearchWritingsResults`, `SearchWritingsNoResults`, `SearchWritingsEmptyWords`, `SearchBlogsResults`, `SearchBlogsNoResults`, `SearchBlogsEmptyWords`, `IsAllowedHost`, `SanitizeBackURL`, `GetPrivateTopicDetails`, `GetPrivateTopicDisplayTitle`, `GetPrivateTopicParticipants`, `PrivateForumTopics`, `PrivateTopics`, `GrantPrivateForumTopic`, `UnreadPrivateThreads`, `UnreadPrivateThreadsCount`, `ForumCategories`, `ForumCategory`, `ForumThreadByID`, `ForumThread`, `ForumThreads`, `ForumTopicByID`, `ForumTopics`, `ForumThreadReplies`, `ForumComment`, `UpdateForumComment`, `EditForumComment`, `SubscribeTopic`, `UnsubscribeTopic`, `SubscribeThread`, `UnsubscribeThread`, `GrantForumCategory`, `RevokeForumCategory`, `GrantForumTopic`, `RevokeForumTopic`, `GrantForumThread`, `RevokeForumThread`, `AllAnsweredFAQ`, `UpdateFAQCategory`, `CreateFAQQuestion`, `SearchWords`, `Article`, `ArticleComments`, `WritingCategories`, `EditableArticle`, `ArticleComment`, `UpdateArticleComment`, `WriterByUsername`, `WriterWritings`, `UpdateWritingReply`, `CreateWritingReply`, `UpdateWriting`, `CreateWriting`, `GrantWritingCategory`, `RevokeWritingCategory`, `CreateWritingCategory`, `ChangeWritingCategory`, `Funcs`, `HasGrant`, `BlogPost`, `BlogComments`, `BlogCategories`, `EditableBlogPost`, `BlogCommentThread`, `CreateBlogReply`, `UpdateBlogReply`, `BloggerProfile`, `BloggerPosts`, `AllBlogs`
- **`StoreImageParams`**:
- **`ThreadInfo`**:
- **`Pagination`** (Interface): Defines a core contract for this module.
- **`OffsetPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`Organization`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`DataCache`**:
- **`OpenGraph`**:
  - Methods: `URLMeta`, `ImageMeta`, `SecureImageMeta`, `ImageWidthMeta`, `ImageHeightMeta`, `TwitterImageMeta`, `TypeMeta`, `ExpirationTimeMeta`, `PublishedTimeMeta`, `ModifiedTimeMeta`, `SiteNameMeta`, `UpdatedTimeMeta`, `JSONLDScript`
- **`CoreOption`**:
- **`PrivateForumInconsistency`**:
- **`ImageBBSThread`**:
- **`AssociateEmailParams`**:
- **`MergeGroup`**:
- **`QuerierFake`**:
  - Methods: `SystemCheckGrant`, `SystemCheckRoleGrant`, `AdminListTopicsWithUserGrantsNoRoles`
- **`CreateFAQQuestionParams`**:
- **`NavigationProvider`** (Interface): Defines a core contract for this module.
- **`PageLink`**:
- **`Person`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`NotFoundLink`**:
- **`ImageBBSThreadPosts`**:
- **`PrivateTopicParticipant`**:
- **`MockQuerier`**:
  - Methods: `GetExternalLink`
- **`JSONLDer`** (Interface): Defines a core contract for this module.
- **`NewsArticle`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`BlogPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`IndexItem`**:
- **`ImageBBSPoster`**:
- **`ImageBBSBoard`**:
- **`FAQ`**:
- **`PageNumberPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`PrivateTopic`**:
- **`Article`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`DiscussionForumPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`IndexGroup`**:
- **`MailProvider`** (Interface): Defines a core contract for this module.
- **`WebAuthnUser`**:
  - Methods: `WebAuthnID`, `WebAuthnName`, `WebAuthnDisplayName`, `WebAuthnIcon`, `WebAuthnCredentials`, `EstablishLegacyCredentialFlags`
- **`Breadcrumb`**:
- **`Goa4WebLinkProvider`**:
  - Methods: `MapImageURL`, `RenderLink`

### Exported Functions

- `TestPrivateForumBreadcrumbBasePath`
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
- `TestHighlightSearchTermsEscapesAndHighlights`
- `TestHighlightSearchTermsRespectsWordBoundaries`
- `TestHighlightSearchTermsWithoutWordsEscapesHTML`
- `TestPrivateForumBreadcrumbUsesDisplayTitle`
- `TestDownloadAndCacheImageRecordsRemoteMetadata`
- `TestPrepareImageCacheEntryForServeRefreshesExpiredRemoteEntry`
- `TestQueueRemoteImageCacheCreatesPendingEntry`
- `TestDownloadExternalImageUsesOpenGraphImageFromHTML`
- `TestCreateCommentStartsImmediateRemoteImageFetch`
- `TestProcessPendingRemoteImageCacheEntriesRecordsRetryFailure`
- `TestPrepareImageCacheEntryForServeAllowsMissingMetadataWhenExpiryDisabled`
- `TestPrepareImageCacheEntryForServeRejectsExpiredRemoteEntryWithoutSource`
- `TestCoreData_PrivateForumTopics_LabelsBug`
- `TestCoreData_PrivateForumTopics_UnreadNew`
- `TestCoreData_PrivateForumTopics_OwnThreadNotNew`
- `CanSearch`
- `TestLoadSelectionsFromRequest`
- `TestCreatePrivateTopicUsesProvidedUsernames`
- `TestCreatePrivateTopicBuildsUsernamesWhenMissing`
- `TestMapLinkURL`
- `TestGetPrivateTopicDisplayTitle`
- `TestSectionItemType`
- `TestBackupEligiblePasskeyPersistence`
- `TestCoreData_PrivateForumTopics`
- `Allowed`
- `TestHandleThreadUpdatedMarksThreadAndItemLabels`
- `TestCreateCommentValidatesGalleryImages`
- `TestMapImageURLUsesDefaultThumbnailForLargeUploadedImage`
- `TestMapImageURLUsesDefaultThumbnailForLargeCachedImage`
- `TestMapImageURLUsesThumbnailForCachedImageWithoutMetadata`
- `TestRecordUploadedImageThumbnailLinksSourceImage`
- `TestStoreImageRecordsDefaultThumbnail`
- `TestSanitizeCodeImagesQueuesImageAliasGoogleRedirect`
- `NewTestCoreData`
- `TestAllRolesLazy`
- `NewGoa4WebLinkProvider`
- `TestRenderLink_Tooltips`
- `TestSearchWordsFromRequestCachesAndReturnsCopy`
- `TestAbsoluteURL`
- `HighlightSearchTerms`
- `TestFormatDuration`
- `TestRenderLink_RoutesThroughGoto`
- `WithPrivateForumTopics`
- `TestQuerierFakeGrantStubs`
- `TestQuerierFakeTopicListing`
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
- `UnmarshalJSONLD`
- `TestAllAnsweredFAQ_Categories`
- `TestRenderLink_Favicon`
- `TestCoreData_PrivateForumTopics_ShowsTopicLabels`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/common"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
