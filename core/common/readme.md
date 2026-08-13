# core/common

## Purpose

Package `common` contains foundational business logic and shared utilities for `common` that are used application-wide.

## Why It Exists

To house logic, constants, and utilities that are required universally across handlers, workers, and internal services.

## What It Allows

It prevents code duplication. For example, `CoreData` is defined here and passed everywhere to provide unified access to the database and configuration state.

## Structure and Components

The primary files and their general responsibilities include:

- `coredata_forum.go`
- `coredata_images_test.go`
- `coredata_misc_test.go`
- `thread_sideeffects.go`
- `coredata.go`
- `coredata_writings.go`
- `search.go`
- `coredata_read_markers.go`
- `role.go`
- `testutil_test.go`
- `coredata_blogs.go`
- `datacache.go`
- `highlight_test.go`
- `link_provider_duration_test.go`
- `privateforum_display_title_test.go`
- `privateforum_labels_test.go`
- `thread_sideeffects_test.go`
- `breadcrumb.go`
- `link_provider_tooltip_test.go`
- `permissions.go`
- `search_words_test.go`
- `breadcrumb_private_test.go`
- `coredata_imagebbs.go`
- `coredata_webauthn_test.go`
- `coredata_user.go`
- `privateforum_topic_labels_test.go`
- `coredata_labels.go`
- `coredata_misc.go`
- `link_provider_test.go`
- `coredata_search.go`
- `download_image.go`
- `faq_test.go`
- `pagination.go`
- `encryption.go`
- `coredata_news.go`
- `download_image_test.go`
- `privateforum.go`
- `url.go`
- `absolute_url_test.go`
- `highlight.go`
- `jsonld.go`
- `usererror.go`
- `coredata_webauthn.go`
- `privateforum_check.go`
- `breadcrumb_private_title_test.go`
- `privateforum_test.go`
- `coredata_request_test.go`
- `funcs.go`
- `link_provider_favicon_test.go`
- `privateforum_merge.go`
- `sectionitemtype_test.go`
- `coredata_auth.go`
- `opengraph_methods.go`
- `search_words.go`
- `coredata_admin.go`
- `faq.go`
- `link_provider.go`
- `testutil.go`
- `signing.go`
- `errors.go`
- `coredata_allroles_test.go`

### Exported Types and Interfaces

- **`OffsetPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`ThreadInfo`**:
- **`DiscussionForumPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`UserError`**:
  - Methods: `Error`, `UserErrorMessage`, `Unwrap`
- **`PrivateForumInconsistency`**:
- **`Goa4WebLinkProvider`**:
  - Methods: `MapImageURL`, `RenderLink`
- **`CoreData`**:
  - Methods: `ForumCategories`, `ForumCategory`, `ForumThreadByID`, `ForumThread`, `ForumThreads`, `ForumTopicByID`, `ForumTopics`, `ForumThreadReplies`, `ForumComment`, `UpdateForumComment`, `EditForumComment`, `SubscribeTopic`, `UnsubscribeTopic`, `SubscribeThread`, `UnsubscribeThread`, `GrantForumCategory`, `RevokeForumCategory`, `GrantForumTopic`, `RevokeForumTopic`, `GrantForumThread`, `RevokeForumThread`, `HandleThreadUpdated`, `AbsoluteURL`, `AdminForumTopics`, `AdminLatestNews`, `AdminLatestNewsList`, `AdminLoginAttempts`, `AdminSessions`, `AdminLinkerItemByID`, `AllRoles`, `RoleByID`, `SelectedRole`, `Announcement`, `AnnouncementLoaded`, `ArchivedRequests`, `BlogEntryByID`, `Bloggers`, `BlogList`, `BlogListForSelectedAuthor`, `Bookmarks`, `CreateBookmark`, `SaveBookmark`, `IsAdmin`, `IsAdminMode`, `CanEditBlog`, `ShowReplyNews`, `ShowEditNews`, `CommentByID`, `CurrentBlog`, `CurrentBlogLoaded`, `CurrentComment`, `CurrentCommentLoaded`, `CurrentNewsPost`, `CurrentNewsPostLoaded`, `CurrentProfileBookmarkSize`, `CurrentProfileComments`, `CurrentProfileEmails`, `CurrentProfileGrants`, `CurrentProfileRoles`, `CurrentProfileStats`, `CurrentProfileUser`, `CurrentRequest`, `CurrentRequestComments`, `CurrentRequestUser`, `CurrentTopic`, `CurrentTopicLoaded`, `CurrentUser`, `CurrentUserLoaded`, `CurrentWriting`, `CurrentWritingLoaded`, `CustomQueries`, `DBRegistry`, `EmailRegistry`, `DefaultNotificationTemplate`, `EmailProvider`, `HTTPClient`, `Event`, `Publish`, `ExecuteSiteTemplate`, `ExternalLink`, `FAQCategories`, `HasAdminRole`, `HasContentWriterRole`, `HasRole`, `HasSubscription`, `ImageBoardPosts`, `ImageBoards`, `ImagePostByID`, `ImageURLMapper`, `Languages`, `RenameLanguage`, `DeleteLanguage`, `CreateLanguage`, `LatestNews`, `LatestNewsList`, `LatestWritings`, `LinkerCategories`, `LinkerCategoriesForUser`, `LinkerCategoryByID`, `LinkerCategoryCounts`, `CreateFAQCategory`, `LinkerItemsForUser`, `LinkerLinksByCategoryID`, `Marked`, `NewsAnnouncement`, `NewsAnnouncementWithErr`, `NewsPostByID`, `Notifications`, `PageSize`, `PendingRequests`, `Permissions`, `Preference`, `PreferredLanguageID`, `Location`, `LocalTime`, `FormatLocalTime`, `LocalTimeIn`, `FormatLocalTimeIn`, `PublicWritings`, `Queries`, `SelectedQuestionFromCategory`, `UpdateFAQQuestion`, `DeleteFAQCategory`, `DeleteFAQQuestion`, `RegisterExternalLinkClick`, `Role`, `SelectedAdminLinkerItem`, `SelectedAdminLinkerItemID`, `SelectedBoardPosts`, `SelectedBoardSubBoards`, `SelectedCategoryPublicWritings`, `SelectedLinkerCategory`, `SelectedLinkerItem`, `SelectedLinkerItemsForCurrentUser`, `SelectedThread`, `SelectedThreadComments`, `SelectedSectionThreadComments`, `SelectedThreadLoaded`, `SelectedThreadCanReply`, `SelectedNewsThreadCanReply`, `SelectedForumThreadCanReply`, `SelectedPrivateForumThreadCanReply`, `SelectedBlogThreadCanReply`, `SelectedImageBBSThreadCanReply`, `SelectedWritingThreadCanReply`, `SelectedLinkerThreadCanReply`, `CreateCommentInSectionForCommenter`, `CreateNewsCommentForCommenter`, `CreateForumCommentForCommenter`, `CreatePrivateForumCommentForCommenter`, `CreateBlogCommentForCommenter`, `CreateImageBBSCommentForCommenter`, `CreateWritingCommentForCommenter`, `CreateLinkerCommentForCommenter`, `CanEditComment`, `CommentEditing`, `CommentEditURL`, `CommentEditSaveURL`, `CommentAdminURL`, `SelectedCommentID`, `Session`, `GetSession`, `SessionManager`, `SetCurrentBlog`, `SetCurrentNewsPost`, `SetCurrentProfileUserID`, `CurrentProfileUserID`, `SetCurrentRequestID`, `CurrentRequestID`, `Offset`, `SetCurrentRoleID`, `SetCurrentSection`, `Section`, `SetCurrentNotificationTemplate`, `SetCurrentError`, `SetCurrentNotice`, `SetCurrentThreadAndTopic`, `SetCurrentWriting`, `SetCurrentExternalLinkID`, `SelectedExternalLink`, `EnsureExternalLink`, `GetExternalLink`, `UpdateExternalLinkMetadata`, `UpdateExternalLinkImageCache`, `SelectedBoardID`, `SelectedThreadID`, `SelectedImagePostID`, `SelectedRoleID`, `SetEvent`, `SetEventTask`, `SetSession`, `SubImageBoards`, `Subscribed`, `Subscriptions`, `CurrentError`, `CustomCSS`, `CurrentNotice`, `NotificationTemplateError`, `NotificationTemplateName`, `NotificationTemplateOverride`, `ThreadComments`, `SectionThreadComments`, `UnreadNotificationCount`, `UserByID`, `UserRoles`, `UserSubscriptions`, `VisibleWritingCategories`, `Writers`, `WritingByID`, `SortedCustomIndexGroups`, `SortedCustomIndexItems`, `LoadSelectionsFromRequest`, `HasModule`, `GenerateFeedURL`, `ValidateCodeImagesForUser`, `ValidateCodeImagesForThread`, `RecordThreadImages`, `ResolveExternalLink`, `ExternalLinkTargetURL`, `ExternalLinkRedirectURL`, `ExternalLinkReloadURL`, `Article`, `ArticleComments`, `WritingCategories`, `EditableArticle`, `ArticleComment`, `UpdateArticleComment`, `WriterByUsername`, `WriterWritings`, `UpdateWritingReply`, `CreateWritingReply`, `UpdateWriting`, `CreateWriting`, `GrantWritingCategory`, `RevokeWritingCategory`, `CreateWritingCategory`, `ChangeWritingCategory`, `CanSearch`, `SetThreadReadMarker`, `ThreadReadMarker`, `BlogPost`, `BlogComments`, `BlogCategories`, `EditableBlogPost`, `BlogCommentThread`, `CreateBlogReply`, `UpdateBlogReply`, `BloggerProfile`, `BloggerPosts`, `AllBlogs`, `Breadcrumbs`, `HasGrant`, `ImageBBSFeed`, `ImageBBSPoster`, `ImageBBSBoard`, `ImageBBSThread`, `ImageBBSThreadPosts`, `UserSettings`, `UserLanguages`, `UserEmails`, `AddUserEmail`, `SaveEmail`, `DeleteEmail`, `AddEmail`, `UserGallery`, `PublicProfile`, `PagedUsers`, `UserNotifications`, `DeleteSubscription`, `UpdateSubscriptions`, `SetUserLanguage`, `SetUserLanguages`, `SetTimezone`, `SaveProfile`, `UpdatePermissions`, `AllowPermission`, `DisallowPermission`, `SaveNotificationDigestPreferences`, `PublicLabels`, `AddPublicLabel`, `RemovePublicLabel`, `AddAuthorLabel`, `RemoveAuthorLabel`, `SetAuthorLabels`, `SetPublicLabels`, `PrivateLabels`, `ClearPrivateLabelStatus`, `ClearUnreadForOthers`, `SetPrivateLabelStatus`, `AddPrivateLabel`, `RemovePrivateLabel`, `SetPrivateLabels`, `TopicPublicLabels`, `AddTopicPublicLabel`, `RemoveTopicPublicLabel`, `SetTopicPublicLabels`, `TopicPrivateLabels`, `AddTopicPrivateLabel`, `RemoveTopicPrivateLabel`, `SetTopicPrivateLabels`, `ThreadPublicLabels`, `AddThreadPublicLabel`, `RemoveThreadPublicLabel`, `AddThreadAuthorLabel`, `RemoveThreadAuthorLabel`, `SetThreadAuthorLabels`, `SetThreadPublicLabels`, `ThreadPrivateLabels`, `ClearThreadPrivateLabelStatus`, `ClearThreadUnreadForOthers`, `SetThreadPrivateLabelStatus`, `AddThreadPrivateLabel`, `RemoveThreadPrivateLabel`, `SetThreadPrivateLabels`, `WritingAuthorLabels`, `AddWritingAuthorLabel`, `RemoveWritingAuthorLabel`, `SetWritingAuthorLabels`, `WritingPrivateLabels`, `SetWritingPrivateLabels`, `ClearWritingUnreadForOthers`, `WritingLabels`, `NewsAuthorLabels`, `AddNewsAuthorLabel`, `RemoveNewsAuthorLabel`, `SetNewsAuthorLabels`, `NewsPrivateLabels`, `SetNewsPrivateLabels`, `NewsLabels`, `BlogAuthorLabels`, `AddBlogAuthorLabel`, `RemoveBlogAuthorLabel`, `SetBlogAuthorLabels`, `BlogPrivateLabels`, `SetBlogPrivateLabels`, `BlogLabels`, `CreatePrivateTopic`, `UploadedImageByImageID`, `StoreImage`, `StoreSystemImage`, `SearchLinker`, `SearchWritings`, `SearchBlogs`, `SearchForum`, `SearchComments`, `SearchCommentsNoResults`, `SearchCommentsEmptyWords`, `SearchLinkerItems`, `SearchLinkerNoResults`, `SearchLinkerEmptyWords`, `SearchWritingsResults`, `SearchWritingsNoResults`, `SearchWritingsEmptyWords`, `SearchBlogsResults`, `SearchBlogsNoResults`, `SearchBlogsEmptyWords`, `DownloadAndCacheImage`, `QueueRemoteImageCache`, `StartRemoteImageCacheFetch`, `ProcessPendingRemoteImageCacheEntries`, `RecordUploadedImageThumbnail`, `RecordUploadedImageDerivative`, `RecordCachedImageThumbnail`, `RecordDerivedImageCacheEntry`, `PrepareImageCacheEntryForServe`, `ImageCacheEntry`, `EncryptData`, `DecryptData`, `ThreadInfo`, `CreateNewsReply`, `UpdateNewsReply`, `UpdateNewsPost`, `DeleteNewsPost`, `CreateNewsPost`, `SearchNews`, `AllowNewsUser`, `DisallowNewsUser`, `AddAnnouncement`, `DeleteAnnouncement`, `SystemGetNewsPost`, `GetPrivateTopicDetails`, `GetPrivateTopicDisplayTitle`, `GetPrivateTopicParticipants`, `PrivateForumTopics`, `PrivateTopics`, `GrantPrivateForumTopic`, `UnreadPrivateThreads`, `UnreadPrivateThreadsCount`, `IsAllowedHost`, `SanitizeBackURL`, `GetWebAuthnUser`, `GetWebAuthnUserByID`, `SavePasskey`, `UpdatePasskeyAfterLogin`, `CheckAndFixPrivateForumInconsistencies`, `Funcs`, `MergePrivateTopicsWithSameParticipants`, `UserCredentials`, `VerifiedEmailsForUser`, `AssociateEmail`, `UserExists`, `CreateUserWithEmail`, `CreatePasswordReset`, `CreatePasswordResetForUser`, `VerifyPasswordReset`, `SearchWords`, `AdminListUsers`, `AdminUserPendingPasswordResetCounts`, `AdminDashboardStats`, `AdminCommentsByUser`, `AdminListPasswordResets`, `AdminApprovePasswordReset`, `AdminDenyPasswordReset`, `AllAnsweredFAQ`, `UpdateFAQCategory`, `CreateFAQQuestion`, `SignShareURL`, `SignShareURLQuery`, `SignImageURL`, `SignCacheURL`, `SignLinkURL`, `SignFeedURL`, `MapImageURL`, `MapFullImageURL`, `ThumbnailReferenceForImage`, `ThumbnailReferenceForCache`, `MapLinkURL`
- **`IndexGroup`**:
- **`MailProvider`** (Interface): Defines a core contract for this module.
- **`PrivateTopicParticipant`**:
- **`MockQuerier`**:
  - Methods: `GetExternalLink`
- **`CategoryFAQs`**:
- **`QuerierFake`**:
  - Methods: `SystemCheckGrant`, `SystemCheckRoleGrant`, `AdminListTopicsWithUserGrantsNoRoles`
- **`OpenGraph`**:
  - Methods: `URLMeta`, `ImageMeta`, `SecureImageMeta`, `ImageWidthMeta`, `ImageHeightMeta`, `TwitterImageMeta`, `TypeMeta`, `ExpirationTimeMeta`, `PublishedTimeMeta`, `ModifiedTimeMeta`, `SiteNameMeta`, `UpdatedTimeMeta`, `JSONLDScript`
- **`NotFoundLink`**:
- **`Breadcrumb`**:
- **`Pagination`** (Interface): Defines a core contract for this module.
- **`JSONLDer`** (Interface): Defines a core contract for this module.
- **`Person`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`ImageBBSBoard`**:
- **`PageLink`**:
- **`PrivateTopic`**:
- **`BlogPosting`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`WebAuthnUser`**:
  - Methods: `WebAuthnID`, `WebAuthnName`, `WebAuthnDisplayName`, `WebAuthnIcon`, `WebAuthnCredentials`, `EstablishLegacyCredentialFlags`
- **`AssociateEmailParams`**:
- **`ThreadUpdatedEvent`**:
- **`LatestWritingsOption`**:
- **`CreatePrivateTopicParams`**:
- **`NewsArticle`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`CoreOption`**:
- **`ImageBBSThread`**:
- **`ImageBBSThreadPosts`**:
- **`MergeGroup`**:
- **`FAQ`**:
- **`CreateFAQQuestionParams`**:
- **`IndexItem`**:
- **`NavigationProvider`** (Interface): Defines a core contract for this module.
- **`DataCache`**:
- **`PageNumberPagination`**:
  - Methods: `StartLink`, `PrevLink`, `NextLink`, `GetLinks`
- **`LanguageCache`**:
  - Methods: `Load`, `Invalidate`
- **`SessionManager`** (Interface): Defines a core contract for this module.
- **`ImageBBSPoster`**:
- **`Organization`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`Article`**:
  - Methods: `LDType`, `MarshalJSONLD`
- **`AdminSection`**:
- **`StoreImageParams`**:

### Exported Functions

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
- `CanSearch`
- `Allowed`
- `WithPrivateForumTopics`
- `HighlightSearchTerms`
- `UnmarshalJSONLD`
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
- `NewGoa4WebLinkProvider`
- `NewTestCoreData`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/core/common"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
