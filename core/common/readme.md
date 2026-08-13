# core/common

## Purpose

Package `common` contains foundational business logic and shared utilities for `common` that are used application-wide.

## Structure and Components

The primary files and their general responsibilities include:

- `coredata_read_markers.go`
- `coredata_webauthn_test.go`
- `coredata_request_test.go`
- `highlight.go`
- `pagination.go`
- `privateforum_display_title_test.go`
- `privateforum_merge.go`
- `coredata_labels.go`
- `coredata_auth.go`
- `download_image_test.go`
- `funcs.go`
- `highlight_test.go`
- `permissions.go`
- `signing.go`
- `url.go`
- `usererror.go`
- `coredata_admin.go`
- `coredata_allroles_test.go`
- `coredata_misc.go`
- `opengraph_methods.go`
- `absolute_url_test.go`
- `coredata.go`
- `privateforum_topic_labels_test.go`
- `search_words_test.go`
- `thread_sideeffects.go`
- `coredata_imagebbs.go`
- `coredata_webauthn.go`
- `testutil_test.go`
- `coredata_writings.go`
- `privateforum_labels_test.go`
- `role.go`
- `coredata_user.go`
- `jsonld.go`
- `link_provider_duration_test.go`
- `sectionitemtype_test.go`
- `testutil.go`
- `breadcrumb_private_test.go`
- `link_provider_tooltip_test.go`
- `privateforum.go`
- `breadcrumb_private_title_test.go`
- `datacache.go`
- `download_image.go`
- `coredata_search.go`
- `errors.go`
- `link_provider_favicon_test.go`
- `privateforum_check.go`
- `thread_sideeffects_test.go`
- `breadcrumb.go`
- `faq.go`
- `search_words.go`
- `coredata_blogs.go`
- `coredata_misc_test.go`
- `privateforum_test.go`
- `coredata_images_test.go`
- `coredata_news.go`
- `encryption.go`
- `link_provider.go`
- `search.go`
- `coredata_forum.go`
- `faq_test.go`
- `link_provider_test.go`

### Exported Types and Interfaces

- **`DiscussionForumPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`PrivateTopic`**:
- **`ThreadInfo`**:
- **`CoreData`**:
  - Methods: `SetThreadReadMarker`, `ThreadReadMarker`, `MergePrivateTopicsWithSameParticipants`, `PublicLabels`, `AddPublicLabel`, `RemovePublicLabel`, `AddAuthorLabel`, `RemoveAuthorLabel`, `SetAuthorLabels`, `SetPublicLabels`, `PrivateLabels`, `ClearPrivateLabelStatus`, `ClearUnreadForOthers`, `SetPrivateLabelStatus`, `AddPrivateLabel`, `RemovePrivateLabel`, `SetPrivateLabels`, `TopicPublicLabels`, `AddTopicPublicLabel`, `RemoveTopicPublicLabel`, `SetTopicPublicLabels`, `TopicPrivateLabels`, `AddTopicPrivateLabel`, `RemoveTopicPrivateLabel`, `SetTopicPrivateLabels`, `ThreadPublicLabels`, `AddThreadPublicLabel`, `RemoveThreadPublicLabel`, `AddThreadAuthorLabel`, `RemoveThreadAuthorLabel`, `SetThreadAuthorLabels`, `SetThreadPublicLabels`, `ThreadPrivateLabels`, `ClearThreadPrivateLabelStatus`, `ClearThreadUnreadForOthers`, `SetThreadPrivateLabelStatus`, `AddThreadPrivateLabel`, `RemoveThreadPrivateLabel`, `SetThreadPrivateLabels`, `WritingAuthorLabels`, `AddWritingAuthorLabel`, `RemoveWritingAuthorLabel`, `SetWritingAuthorLabels`, `WritingPrivateLabels`, `SetWritingPrivateLabels`, `ClearWritingUnreadForOthers`, `WritingLabels`, `NewsAuthorLabels`, `AddNewsAuthorLabel`, `RemoveNewsAuthorLabel`, `SetNewsAuthorLabels`, `NewsPrivateLabels`, `SetNewsPrivateLabels`, `NewsLabels`, `BlogAuthorLabels`, `AddBlogAuthorLabel`, `RemoveBlogAuthorLabel`, `SetBlogAuthorLabels`, `BlogPrivateLabels`, `SetBlogPrivateLabels`, `BlogLabels`, `UserCredentials`, `VerifiedEmailsForUser`, `AssociateEmail`, `UserExists`, `CreateUserWithEmail`, `CreatePasswordReset`, `CreatePasswordResetForUser`, `VerifyPasswordReset`, `Funcs`, `HasGrant`, `SignShareURL`, `SignShareURLQuery`, `SignImageURL`, `SignCacheURL`, `SignLinkURL`, `SignFeedURL`, `MapImageURL`, `MapFullImageURL`, `ThumbnailReferenceForImage`, `ThumbnailReferenceForCache`, `MapLinkURL`, `IsAllowedHost`, `SanitizeBackURL`, `AdminListUsers`, `AdminUserPendingPasswordResetCounts`, `AdminDashboardStats`, `AdminCommentsByUser`, `AdminListPasswordResets`, `AdminApprovePasswordReset`, `AdminDenyPasswordReset`, `CreatePrivateTopic`, `UploadedImageByImageID`, `StoreImage`, `StoreSystemImage`, `AbsoluteURL`, `AdminForumTopics`, `AdminLatestNews`, `AdminLatestNewsList`, `AdminLoginAttempts`, `AdminSessions`, `AdminLinkerItemByID`, `AllRoles`, `RoleByID`, `SelectedRole`, `Announcement`, `AnnouncementLoaded`, `ArchivedRequests`, `BlogEntryByID`, `Bloggers`, `BlogList`, `BlogListForSelectedAuthor`, `Bookmarks`, `CreateBookmark`, `SaveBookmark`, `IsAdmin`, `IsAdminMode`, `CanEditBlog`, `ShowReplyNews`, `ShowEditNews`, `CommentByID`, `CurrentBlog`, `CurrentBlogLoaded`, `CurrentComment`, `CurrentCommentLoaded`, `CurrentNewsPost`, `CurrentNewsPostLoaded`, `CurrentProfileBookmarkSize`, `CurrentProfileComments`, `CurrentProfileEmails`, `CurrentProfileGrants`, `CurrentProfileRoles`, `CurrentProfileStats`, `CurrentProfileUser`, `CurrentRequest`, `CurrentRequestComments`, `CurrentRequestUser`, `CurrentTopic`, `CurrentTopicLoaded`, `CurrentUser`, `CurrentUserLoaded`, `CurrentWriting`, `CurrentWritingLoaded`, `CustomQueries`, `DBRegistry`, `EmailRegistry`, `DefaultNotificationTemplate`, `EmailProvider`, `HTTPClient`, `Event`, `Publish`, `ExecuteSiteTemplate`, `ExternalLink`, `FAQCategories`, `HasAdminRole`, `HasContentWriterRole`, `HasRole`, `HasSubscription`, `ImageBoardPosts`, `ImageBoards`, `ImagePostByID`, `ImageURLMapper`, `Languages`, `RenameLanguage`, `DeleteLanguage`, `CreateLanguage`, `LatestNews`, `LatestNewsList`, `LatestWritings`, `LinkerCategories`, `LinkerCategoriesForUser`, `LinkerCategoryByID`, `LinkerCategoryCounts`, `CreateFAQCategory`, `LinkerItemsForUser`, `LinkerLinksByCategoryID`, `Marked`, `NewsAnnouncement`, `NewsAnnouncementWithErr`, `NewsPostByID`, `Notifications`, `PageSize`, `PendingRequests`, `Permissions`, `Preference`, `PreferredLanguageID`, `Location`, `LocalTime`, `FormatLocalTime`, `LocalTimeIn`, `FormatLocalTimeIn`, `PublicWritings`, `Queries`, `SelectedQuestionFromCategory`, `UpdateFAQQuestion`, `DeleteFAQCategory`, `DeleteFAQQuestion`, `RegisterExternalLinkClick`, `Role`, `SelectedAdminLinkerItem`, `SelectedAdminLinkerItemID`, `SelectedBoardPosts`, `SelectedBoardSubBoards`, `SelectedCategoryPublicWritings`, `SelectedLinkerCategory`, `SelectedLinkerItem`, `SelectedLinkerItemsForCurrentUser`, `SelectedThread`, `SelectedThreadComments`, `SelectedSectionThreadComments`, `SelectedThreadLoaded`, `SelectedThreadCanReply`, `SelectedNewsThreadCanReply`, `SelectedForumThreadCanReply`, `SelectedPrivateForumThreadCanReply`, `SelectedBlogThreadCanReply`, `SelectedImageBBSThreadCanReply`, `SelectedWritingThreadCanReply`, `SelectedLinkerThreadCanReply`, `CreateCommentInSectionForCommenter`, `CreateNewsCommentForCommenter`, `CreateForumCommentForCommenter`, `CreatePrivateForumCommentForCommenter`, `CreateBlogCommentForCommenter`, `CreateImageBBSCommentForCommenter`, `CreateWritingCommentForCommenter`, `CreateLinkerCommentForCommenter`, `CanEditComment`, `CommentEditing`, `CommentEditURL`, `CommentEditSaveURL`, `CommentAdminURL`, `SelectedCommentID`, `Session`, `GetSession`, `SessionManager`, `SetCurrentBlog`, `SetCurrentNewsPost`, `SetCurrentProfileUserID`, `CurrentProfileUserID`, `SetCurrentRequestID`, `CurrentRequestID`, `Offset`, `SetCurrentRoleID`, `SetCurrentSection`, `Section`, `SetCurrentNotificationTemplate`, `SetCurrentError`, `SetCurrentNotice`, `SetCurrentThreadAndTopic`, `SetCurrentWriting`, `SetCurrentExternalLinkID`, `SelectedExternalLink`, `EnsureExternalLink`, `GetExternalLink`, `UpdateExternalLinkMetadata`, `UpdateExternalLinkImageCache`, `SelectedBoardID`, `SelectedThreadID`, `SelectedImagePostID`, `SelectedRoleID`, `SetEvent`, `SetEventTask`, `SetSession`, `SubImageBoards`, `Subscribed`, `Subscriptions`, `CurrentError`, `CustomCSS`, `CurrentNotice`, `NotificationTemplateError`, `NotificationTemplateName`, `NotificationTemplateOverride`, `ThreadComments`, `SectionThreadComments`, `UnreadNotificationCount`, `UserByID`, `UserRoles`, `UserSubscriptions`, `VisibleWritingCategories`, `Writers`, `WritingByID`, `SortedCustomIndexGroups`, `SortedCustomIndexItems`, `LoadSelectionsFromRequest`, `HasModule`, `GenerateFeedURL`, `ValidateCodeImagesForUser`, `ValidateCodeImagesForThread`, `RecordThreadImages`, `ResolveExternalLink`, `ExternalLinkTargetURL`, `ExternalLinkRedirectURL`, `ExternalLinkReloadURL`, `HandleThreadUpdated`, `ImageBBSFeed`, `ImageBBSPoster`, `ImageBBSBoard`, `ImageBBSThread`, `ImageBBSThreadPosts`, `GetWebAuthnUser`, `GetWebAuthnUserByID`, `SavePasskey`, `UpdatePasskeyAfterLogin`, `Article`, `ArticleComments`, `WritingCategories`, `EditableArticle`, `ArticleComment`, `UpdateArticleComment`, `WriterByUsername`, `WriterWritings`, `UpdateWritingReply`, `CreateWritingReply`, `UpdateWriting`, `CreateWriting`, `GrantWritingCategory`, `RevokeWritingCategory`, `CreateWritingCategory`, `ChangeWritingCategory`, `UserSettings`, `UserLanguages`, `UserEmails`, `AddUserEmail`, `SaveEmail`, `DeleteEmail`, `AddEmail`, `UserGallery`, `PublicProfile`, `PagedUsers`, `UserNotifications`, `DeleteSubscription`, `UpdateSubscriptions`, `SetUserLanguage`, `SetUserLanguages`, `SetTimezone`, `SaveProfile`, `UpdatePermissions`, `AllowPermission`, `DisallowPermission`, `SaveNotificationDigestPreferences`, `GetPrivateTopicDetails`, `GetPrivateTopicDisplayTitle`, `GetPrivateTopicParticipants`, `PrivateForumTopics`, `PrivateTopics`, `GrantPrivateForumTopic`, `UnreadPrivateThreads`, `UnreadPrivateThreadsCount`, `DownloadAndCacheImage`, `QueueRemoteImageCache`, `StartRemoteImageCacheFetch`, `ProcessPendingRemoteImageCacheEntries`, `RecordUploadedImageThumbnail`, `RecordUploadedImageDerivative`, `RecordCachedImageThumbnail`, `RecordDerivedImageCacheEntry`, `PrepareImageCacheEntryForServe`, `ImageCacheEntry`, `SearchLinker`, `SearchWritings`, `SearchBlogs`, `SearchForum`, `SearchComments`, `SearchCommentsNoResults`, `SearchCommentsEmptyWords`, `SearchLinkerItems`, `SearchLinkerNoResults`, `SearchLinkerEmptyWords`, `SearchWritingsResults`, `SearchWritingsNoResults`, `SearchWritingsEmptyWords`, `SearchBlogsResults`, `SearchBlogsNoResults`, `SearchBlogsEmptyWords`, `CheckAndFixPrivateForumInconsistencies`, `Breadcrumbs`, `AllAnsweredFAQ`, `UpdateFAQCategory`, `CreateFAQQuestion`, `SearchWords`, `BlogPost`, `BlogComments`, `BlogCategories`, `EditableBlogPost`, `BlogCommentThread`, `CreateBlogReply`, `UpdateBlogReply`, `BloggerProfile`, `BloggerPosts`, `AllBlogs`, `ThreadInfo`, `CreateNewsReply`, `UpdateNewsReply`, `UpdateNewsPost`, `DeleteNewsPost`, `CreateNewsPost`, `SearchNews`, `AllowNewsUser`, `DisallowNewsUser`, `AddAnnouncement`, `DeleteAnnouncement`, `SystemGetNewsPost`, `EncryptData`, `DecryptData`, `CanSearch`, `ForumCategories`, `ForumCategory`, `ForumThreadByID`, `ForumThread`, `ForumThreads`, `ForumTopicByID`, `ForumTopics`, `ForumThreadReplies`, `ForumComment`, `UpdateForumComment`, `EditForumComment`, `SubscribeTopic`, `UnsubscribeTopic`, `SubscribeThread`, `UnsubscribeThread`, `GrantForumCategory`, `RevokeForumCategory`, `GrantForumTopic`, `RevokeForumTopic`, `GrantForumThread`, `RevokeForumThread`
- **`PageNumberPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`AssociateEmailParams`**:
- **`StoreImageParams`**:
- **`CoreOption`**:
- **`LatestWritingsOption`**:
- **`Article`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`QuerierFake`**:
  - Methods: `SystemCheckGrant`, `SystemCheckRoleGrant`, `AdminListTopicsWithUserGrantsNoRoles`
- **`ImageBBSThread`**:
- **`WebAuthnUser`**:
  - Methods: `WebAuthnID`, `WebAuthnName`, `WebAuthnDisplayName`, `WebAuthnIcon`, `WebAuthnCredentials`, `EstablishLegacyCredentialFlags`
- **`CreatePrivateTopicParams`**:
- **`LanguageCache`**:
  - Methods: `Load`, `Invalidate`
- **`IndexGroup`**:
- **`SessionManager`** (Interface): Defines a core contract for this module.
- **`ImageBBSPoster`**:
- **`DataCache`**:
- **`OffsetPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`PrivateTopicParticipant`**:
- **`MailProvider`** (Interface): Defines a core contract for this module.
- **`ThreadUpdatedEvent`**:
- **`Person`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`PrivateForumInconsistency`**:
- **`PageLink`**:
- **`IndexItem`**:
- **`JSONLDer`** (Interface): Defines a core contract for this module.
- **`FAQ`**:
- **`CategoryFAQs`**:
- **`UserError`**:
  - Methods: `Error`, `UserErrorMessage`, `Unwrap`
- **`Organization`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`ImageBBSThreadPosts`**:
- **`OpenGraph`**:
  - Methods: `URLMeta`, `ImageMeta`, `SecureImageMeta`, `ImageWidthMeta`, `ImageHeightMeta`, `TwitterImageMeta`, `TypeMeta`, `ExpirationTimeMeta`, `PublishedTimeMeta`, `ModifiedTimeMeta`, `SiteNameMeta`, `UpdatedTimeMeta`, `JSONLDScript`
- **`NotFoundLink`**:
- **`NavigationProvider`** (Interface): Defines a core contract for this module.
- **`NewsArticle`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`Goa4WebLinkProvider`**:
  - Methods: `MapImageURL`, `RenderLink`
- **`Pagination`** (Interface): Defines a core contract for this module.
- **`MergeGroup`**:
- **`ImageBBSBoard`**:
- **`Breadcrumb`**:
- **`CreateFAQQuestionParams`**:
- **`MockQuerier`**:
  - Methods: `GetExternalLink`
- **`AdminSection`**:
- **`BlogPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`

### Exported Functions

- `TestBackupEligiblePasskeyPersistence`
- `TestLoadSelectionsFromRequest`
- `HighlightSearchTerms`
- `TestGetPrivateTopicDisplayTitle`
- `TestDownloadAndCacheImageRecordsRemoteMetadata`
- `TestPrepareImageCacheEntryForServeRefreshesExpiredRemoteEntry`
- `TestQueueRemoteImageCacheCreatesPendingEntry`
- `TestDownloadExternalImageUsesOpenGraphImageFromHTML`
- `TestCreateCommentStartsImmediateRemoteImageFetch`
- `TestProcessPendingRemoteImageCacheEntriesRecordsRetryFailure`
- `TestPrepareImageCacheEntryForServeAllowsMissingMetadataWhenExpiryDisabled`
- `TestPrepareImageCacheEntryForServeRejectsExpiredRemoteEntryWithoutSource`
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
- `TestHighlightSearchTermsEscapesAndHighlights`
- `TestHighlightSearchTermsRespectsWordBoundaries`
- `TestHighlightSearchTermsWithoutWordsEscapesHTML`
- `TestAllRolesLazy`
- `TestAbsoluteURL`
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
- `TestCoreData_PrivateForumTopics_ShowsTopicLabels`
- `TestSearchWordsFromRequestCachesAndReturnsCopy`
- `TestQuerierFakeGrantStubs`
- `TestQuerierFakeTopicListing`
- `TestCoreData_PrivateForumTopics_LabelsBug`
- `TestCoreData_PrivateForumTopics_UnreadNew`
- `TestCoreData_PrivateForumTopics_OwnThreadNotNew`
- `Allowed`
- `UnmarshalJSONLD`
- `TestFormatDuration`
- `TestSectionItemType`
- `NewTestCoreData`
- `TestPrivateForumBreadcrumbBasePath`
- `TestRenderLink_Tooltips`
- `WithPrivateForumTopics`
- `TestPrivateForumBreadcrumbUsesDisplayTitle`
- `TestRenderLink_Favicon`
- `TestHandleThreadUpdatedMarksThreadAndItemLabels`
- `TestCreatePrivateTopicUsesProvidedUsernames`
- `TestCreatePrivateTopicBuildsUsernamesWhenMissing`
- `TestMapLinkURL`
- `TestCoreData_PrivateForumTopics`
- `TestCreateCommentValidatesGalleryImages`
- `TestMapImageURLUsesDefaultThumbnailForLargeUploadedImage`
- `TestMapImageURLUsesDefaultThumbnailForLargeCachedImage`
- `TestMapImageURLUsesThumbnailForCachedImageWithoutMetadata`
- `TestRecordUploadedImageThumbnailLinksSourceImage`
- `TestStoreImageRecordsDefaultThumbnail`
- `TestSanitizeCodeImagesQueuesImageAliasGoogleRedirect`
- `NewGoa4WebLinkProvider`
- `CanSearch`
- `TestAllAnsweredFAQ_Categories`
- `TestRenderLink_RoutesThroughGoto`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/core/common"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
