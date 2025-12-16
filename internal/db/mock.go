package db

import (
	"context"
	"database/sql"
	"time"
)

var _ Querier = &QuerierProxier{}

type QuerierProxier struct {
	Querier
	OverwrittenAddContentLabelStatus                                                                          func(ctx context.Context, arg AddContentLabelStatusParams) error
	OverwrittenAddContentPrivateLabel                                                                         func(ctx context.Context, arg AddContentPrivateLabelParams) error
	OverwrittenAddContentPublicLabel                                                                          func(ctx context.Context, arg AddContentPublicLabelParams) error
	OverwrittenAdminApproveImagePost                                                                          func(ctx context.Context, idimagepost int32) error
	OverwrittenAdminArchiveBlog                                                                               func(ctx context.Context, arg AdminArchiveBlogParams) error
	OverwrittenAdminArchiveComment                                                                            func(ctx context.Context, arg AdminArchiveCommentParams) error
	OverwrittenAdminArchiveImagepost                                                                          func(ctx context.Context, arg AdminArchiveImagepostParams) error
	OverwrittenAdminArchiveLink                                                                               func(ctx context.Context, arg AdminArchiveLinkParams) error
	OverwrittenAdminArchiveUser                                                                               func(ctx context.Context, idusers int32) error
	OverwrittenAdminArchiveWriting                                                                            func(ctx context.Context, arg AdminArchiveWritingParams) error
	OverwrittenAdminCancelBannedIp                                                                            func(ctx context.Context, ipNet string) error
	OverwrittenAdminClearExternalLinkCache                                                                    func(ctx context.Context, arg AdminClearExternalLinkCacheParams) error
	OverwrittenAdminCompleteWordList                                                                          func(ctx context.Context) ([]sql.NullString, error)
	OverwrittenAdminCountForumCategories                                                                      func(ctx context.Context, arg AdminCountForumCategoriesParams) (int64, error)
	OverwrittenAdminCountForumThreads                                                                         func(ctx context.Context) (int64, error)
	OverwrittenAdminCountForumTopics                                                                          func(ctx context.Context) (int64, error)
	OverwrittenAdminCountLinksByCategory                                                                      func(ctx context.Context, categoryID sql.NullInt32) (int64, error)
	OverwrittenAdminCountThreadsByBoard                                                                       func(ctx context.Context, imageboardIdimageboard sql.NullInt32) (int64, error)
	OverwrittenAdminCountWordList                                                                             func(ctx context.Context) (int64, error)
	OverwrittenAdminCountWordListByPrefix                                                                     func(ctx context.Context, prefix interface{}) (int64, error)
	OverwrittenAdminCreateFAQCategory                                                                         func(ctx context.Context, name sql.NullString) error
	OverwrittenAdminCreateGrant                                                                               func(ctx context.Context, arg AdminCreateGrantParams) (int64, error)
	OverwrittenAdminCreateImageBoard                                                                          func(ctx context.Context, arg AdminCreateImageBoardParams) error
	OverwrittenAdminCreateLanguage                                                                            func(ctx context.Context, nameof sql.NullString) error
	OverwrittenAdminCreateLinkerCategory                                                                      func(ctx context.Context, arg AdminCreateLinkerCategoryParams) error
	OverwrittenAdminCreateLinkerItem                                                                          func(ctx context.Context, arg AdminCreateLinkerItemParams) error
	OverwrittenAdminDeleteExternalLink                                                                        func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteExternalLinkByURL                                                                   func(ctx context.Context, url string) error
	OverwrittenAdminDeleteFAQ                                                                                 func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteFAQCategory                                                                         func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteForumCategory                                                                       func(ctx context.Context, idforumcategory int32) error
	OverwrittenAdminDeleteForumThread                                                                         func(ctx context.Context, idforumthread int32) error
	OverwrittenAdminDeleteForumTopic                                                                          func(ctx context.Context, idforumtopic int32) error
	OverwrittenAdminDeleteGrant                                                                               func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteImageBoard                                                                          func(ctx context.Context, idimageboard int32) error
	OverwrittenAdminDeleteImagePost                                                                           func(ctx context.Context, idimagepost int32) error
	OverwrittenAdminDeleteLanguage                                                                            func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteLinkerCategory                                                                      func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteLinkerQueuedItem                                                                    func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteNotification                                                                        func(ctx context.Context, id int32) error
	OverwrittenAdminDeletePendingEmail                                                                        func(ctx context.Context, id int32) error
	OverwrittenAdminDeleteTemplateOverride                                                                    func(ctx context.Context, name string) error
	OverwrittenAdminDeleteUserByID                                                                            func(ctx context.Context, idusers int32) error
	OverwrittenAdminDeleteUserRole                                                                            func(ctx context.Context, iduserRoles int32) error
	OverwrittenAdminDemoteAnnouncement                                                                        func(ctx context.Context, id int32) error
	OverwrittenAdminForumCategoryThreadCounts                                                                 func(ctx context.Context) ([]*AdminForumCategoryThreadCountsRow, error)
	OverwrittenAdminForumHandlerThreadCounts                                                                  func(ctx context.Context) ([]*AdminForumHandlerThreadCountsRow, error)
	OverwrittenAdminForumTopicThreadCounts                                                                    func(ctx context.Context) ([]*AdminForumTopicThreadCountsRow, error)
	OverwrittenAdminGetAllBlogEntriesByUser                                                                   func(ctx context.Context, authorID int32) ([]*AdminGetAllBlogEntriesByUserRow, error)
	OverwrittenAdminGetAllCommentsByUser                                                                      func(ctx context.Context, userID int32) ([]*AdminGetAllCommentsByUserRow, error)
	OverwrittenAdminGetAllWritingsByAuthor                                                                    func(ctx context.Context, authorID int32) ([]*AdminGetAllWritingsByAuthorRow, error)
	OverwrittenAdminGetDashboardStats                                                                         func(ctx context.Context) (*AdminGetDashboardStatsRow, error)
	OverwrittenAdminGetFAQByID                                                                                func(ctx context.Context, id int32) (*Faq, error)
	OverwrittenAdminGetFAQCategories                                                                          func(ctx context.Context) ([]*FaqCategory, error)
	OverwrittenAdminGetFAQCategoriesWithQuestionCount                                                         func(ctx context.Context) ([]*AdminGetFAQCategoriesWithQuestionCountRow, error)
	OverwrittenAdminGetFAQCategoryWithQuestionCountByID                                                       func(ctx context.Context, id int32) (*AdminGetFAQCategoryWithQuestionCountByIDRow, error)
	OverwrittenAdminGetFAQDismissedQuestions                                                                  func(ctx context.Context) ([]*Faq, error)
	OverwrittenAdminGetFAQQuestionsByCategory                                                                 func(ctx context.Context, categoryID sql.NullInt32) ([]*Faq, error)
	OverwrittenAdminGetFAQUnansweredQuestions                                                                 func(ctx context.Context) ([]*Faq, error)
	OverwrittenAdminGetForumStats                                                                             func(ctx context.Context) (*AdminGetForumStatsRow, error)
	OverwrittenAdminGetImagePost                                                                              func(ctx context.Context, idimagepost int32) (*AdminGetImagePostRow, error)
	OverwrittenAdminGetNotification                                                                           func(ctx context.Context, id int32) (*Notification, error)
	OverwrittenAdminGetPendingEmailByID                                                                       func(ctx context.Context, id int32) (*AdminGetPendingEmailByIDRow, error)
	OverwrittenAdminGetRecentAuditLogs                                                                        func(ctx context.Context, limit int32) ([]*AdminGetRecentAuditLogsRow, error)
	OverwrittenAdminGetRequestByID                                                                            func(ctx context.Context, id int32) (*AdminRequestQueue, error)
	OverwrittenAdminGetRoleByID                                                                               func(ctx context.Context, id int32) (*Role, error)
	OverwrittenAdminGetRoleByNameForUser                                                                      func(ctx context.Context, arg AdminGetRoleByNameForUserParams) (int32, error)
	OverwrittenAdminGetSearchStats                                                                            func(ctx context.Context) (*AdminGetSearchStatsRow, error)
	OverwrittenAdminGetThreadsStartedByUser                                                                   func(ctx context.Context, usersIdusers int32) ([]*Forumthread, error)
	OverwrittenAdminGetThreadsStartedByUserWithTopic                                                          func(ctx context.Context, usersIdusers int32) ([]*AdminGetThreadsStartedByUserWithTopicRow, error)
	OverwrittenAdminGetWritingsByCategoryId                                                                   func(ctx context.Context, writingCategoryID int32) ([]*AdminGetWritingsByCategoryIdRow, error)
	OverwrittenAdminImageboardPostCounts                                                                      func(ctx context.Context) ([]*AdminImageboardPostCountsRow, error)
	OverwrittenAdminInsertBannedIp                                                                            func(ctx context.Context, arg AdminInsertBannedIpParams) error
	OverwrittenAdminInsertLanguage                                                                            func(ctx context.Context, nameof sql.NullString) (sql.Result, error)
	OverwrittenAdminInsertQueuedLinkFromQueue                                                                 func(ctx context.Context, id int32) (int64, error)
	OverwrittenAdminInsertRequestComment                                                                      func(ctx context.Context, arg AdminInsertRequestCommentParams) error
	OverwrittenAdminInsertRequestQueue                                                                        func(ctx context.Context, arg AdminInsertRequestQueueParams) (sql.Result, error)
	OverwrittenAdminInsertWritingCategory                                                                     func(ctx context.Context, arg AdminInsertWritingCategoryParams) error
	OverwrittenAdminIsBlogDeactivated                                                                         func(ctx context.Context, idblogs int32) (bool, error)
	OverwrittenAdminIsCommentDeactivated                                                                      func(ctx context.Context, idcomments int32) (bool, error)
	OverwrittenAdminIsImagepostDeactivated                                                                    func(ctx context.Context, idimagepost int32) (bool, error)
	OverwrittenAdminIsLinkDeactivated                                                                         func(ctx context.Context, id int32) (bool, error)
	OverwrittenAdminIsUserDeactivated                                                                         func(ctx context.Context, idusers int32) (bool, error)
	OverwrittenAdminIsWritingDeactivated                                                                      func(ctx context.Context, idwriting int32) (bool, error)
	OverwrittenAdminLanguageUsageCounts                                                                       func(ctx context.Context, arg AdminLanguageUsageCountsParams) (*AdminLanguageUsageCountsRow, error)
	OverwrittenAdminListAdministratorEmails                                                                   func(ctx context.Context) ([]string, error)
	OverwrittenAdminListAllCommentsWithThreadInfo                                                             func(ctx context.Context, arg AdminListAllCommentsWithThreadInfoParams) ([]*AdminListAllCommentsWithThreadInfoRow, error)
	OverwrittenAdminListAllPrivateForumThreads                                                                func(ctx context.Context) ([]*AdminListAllPrivateForumThreadsRow, error)
	OverwrittenAdminListAllPrivateTopics                                                                      func(ctx context.Context) ([]*AdminListAllPrivateTopicsRow, error)
	OverwrittenAdminListAllUserIDs                                                                            func(ctx context.Context) ([]int32, error)
	OverwrittenAdminListAllUsers                                                                              func(ctx context.Context) ([]*AdminListAllUsersRow, error)
	OverwrittenAdminListAnnouncementsWithNews                                                                 func(ctx context.Context) ([]*AdminListAnnouncementsWithNewsRow, error)
	OverwrittenAdminListArchivedRequests                                                                      func(ctx context.Context) ([]*AdminRequestQueue, error)
	OverwrittenAdminListAuditLogs                                                                             func(ctx context.Context, arg AdminListAuditLogsParams) ([]*AdminListAuditLogsRow, error)
	OverwrittenAdminListBoards                                                                                func(ctx context.Context, arg AdminListBoardsParams) ([]*Imageboard, error)
	OverwrittenAdminListDeactivatedBlogs                                                                      func(ctx context.Context, arg AdminListDeactivatedBlogsParams) ([]*AdminListDeactivatedBlogsRow, error)
	OverwrittenAdminListDeactivatedComments                                                                   func(ctx context.Context, arg AdminListDeactivatedCommentsParams) ([]*AdminListDeactivatedCommentsRow, error)
	OverwrittenAdminListDeactivatedImageposts                                                                 func(ctx context.Context, arg AdminListDeactivatedImagepostsParams) ([]*AdminListDeactivatedImagepostsRow, error)
	OverwrittenAdminListDeactivatedLinks                                                                      func(ctx context.Context, arg AdminListDeactivatedLinksParams) ([]*AdminListDeactivatedLinksRow, error)
	OverwrittenAdminListDeactivatedUsers                                                                      func(ctx context.Context, arg AdminListDeactivatedUsersParams) ([]*AdminListDeactivatedUsersRow, error)
	OverwrittenAdminListDeactivatedWritings                                                                   func(ctx context.Context, arg AdminListDeactivatedWritingsParams) ([]*AdminListDeactivatedWritingsRow, error)
	OverwrittenAdminListExternalLinks                                                                         func(ctx context.Context, arg AdminListExternalLinksParams) ([]*ExternalLink, error)
	OverwrittenAdminListFailedEmails                                                                          func(ctx context.Context, arg AdminListFailedEmailsParams) ([]*AdminListFailedEmailsRow, error)
	OverwrittenAdminListForumCategoriesWithCounts                                                             func(ctx context.Context, arg AdminListForumCategoriesWithCountsParams) ([]*AdminListForumCategoriesWithCountsRow, error)
	OverwrittenAdminListForumTopics                                                                           func(ctx context.Context, arg AdminListForumTopicsParams) ([]*Forumtopic, error)
	OverwrittenAdminListGrantsByRoleID                                                                        func(ctx context.Context, roleID sql.NullInt32) ([]*Grant, error)
	OverwrittenAdminListGrantsByThreadID                                                                      func(ctx context.Context, itemID sql.NullInt32) ([]*AdminListGrantsByThreadIDRow, error)
	OverwrittenAdminListGrantsByTopicID                                                                       func(ctx context.Context, itemID sql.NullInt32) ([]*AdminListGrantsByTopicIDRow, error)
	OverwrittenAdminListLoginAttempts                                                                         func(ctx context.Context) ([]*LoginAttempt, error)
	OverwrittenAdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescending                            func(ctx context.Context, arg AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams) ([]*AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error)
	OverwrittenAdminListPendingDeactivatedBlogs                                                               func(ctx context.Context, arg AdminListPendingDeactivatedBlogsParams) ([]*AdminListPendingDeactivatedBlogsRow, error)
	OverwrittenAdminListPendingDeactivatedComments                                                            func(ctx context.Context, arg AdminListPendingDeactivatedCommentsParams) ([]*AdminListPendingDeactivatedCommentsRow, error)
	OverwrittenAdminListPendingDeactivatedImageposts                                                          func(ctx context.Context, arg AdminListPendingDeactivatedImagepostsParams) ([]*AdminListPendingDeactivatedImagepostsRow, error)
	OverwrittenAdminListPendingDeactivatedLinks                                                               func(ctx context.Context, arg AdminListPendingDeactivatedLinksParams) ([]*AdminListPendingDeactivatedLinksRow, error)
	OverwrittenAdminListPendingDeactivatedWritings                                                            func(ctx context.Context, arg AdminListPendingDeactivatedWritingsParams) ([]*AdminListPendingDeactivatedWritingsRow, error)
	OverwrittenAdminListPendingRequests                                                                       func(ctx context.Context) ([]*AdminRequestQueue, error)
	OverwrittenAdminListPendingUsers                                                                          func(ctx context.Context) ([]*AdminListPendingUsersRow, error)
	OverwrittenAdminListRecentNotifications                                                                   func(ctx context.Context, limit int32) ([]*Notification, error)
	OverwrittenAdminListRequestComments                                                                       func(ctx context.Context, requestID int32) ([]*AdminRequestComment, error)
	OverwrittenAdminListRoles                                                                                 func(ctx context.Context) ([]*Role, error)
	OverwrittenAdminListRolesWithUsers                                                                        func(ctx context.Context) ([]*AdminListRolesWithUsersRow, error)
	OverwrittenAdminListSentEmails                                                                            func(ctx context.Context, arg AdminListSentEmailsParams) ([]*AdminListSentEmailsRow, error)
	OverwrittenAdminListSessions                                                                              func(ctx context.Context) ([]*AdminListSessionsRow, error)
	OverwrittenAdminListTopicsWithUserGrantsNoRoles                                                           func(ctx context.Context, includeAdmin interface{}) ([]*AdminListTopicsWithUserGrantsNoRolesRow, error)
	OverwrittenAdminListUnsentPendingEmails                                                                   func(ctx context.Context, arg AdminListUnsentPendingEmailsParams) ([]*AdminListUnsentPendingEmailsRow, error)
	OverwrittenAdminListUploadedImages                                                                        func(ctx context.Context, arg AdminListUploadedImagesParams) ([]*UploadedImage, error)
	OverwrittenAdminListUserEmails                                                                            func(ctx context.Context, userID int32) ([]*UserEmail, error)
	OverwrittenAdminListUserIDsByRole                                                                         func(ctx context.Context, name string) ([]int32, error)
	OverwrittenAdminListUsersByID                                                                             func(ctx context.Context, ids []int32) ([]*AdminListUsersByIDRow, error)
	OverwrittenAdminListUsersByRoleID                                                                         func(ctx context.Context, roleID int32) ([]*AdminListUsersByRoleIDRow, error)
	OverwrittenAdminMarkBlogRestored                                                                          func(ctx context.Context, idblogs int32) error
	OverwrittenAdminMarkCommentRestored                                                                       func(ctx context.Context, idcomments int32) error
	OverwrittenAdminMarkImagepostRestored                                                                     func(ctx context.Context, idimagepost int32) error
	OverwrittenAdminMarkLinkRestored                                                                          func(ctx context.Context, id int32) error
	OverwrittenAdminMarkNotificationRead                                                                      func(ctx context.Context, id int32) error
	OverwrittenAdminMarkNotificationUnread                                                                    func(ctx context.Context, id int32) error
	OverwrittenAdminMarkWritingRestored                                                                       func(ctx context.Context, idwriting int32) error
	OverwrittenAdminPromoteAnnouncement                                                                       func(ctx context.Context, siteNewsID int32) error
	OverwrittenAdminPurgeReadNotifications                                                                    func(ctx context.Context) error
	OverwrittenAdminRebuildAllForumTopicMetaColumns                                                           func(ctx context.Context) error
	OverwrittenAdminRecalculateAllForumThreadMetaData                                                         func(ctx context.Context) error
	OverwrittenAdminRecalculateForumThreadByIdMetaData                                                        func(ctx context.Context, idforumthread int32) error
	OverwrittenAdminRenameFAQCategory                                                                         func(ctx context.Context, arg AdminRenameFAQCategoryParams) error
	OverwrittenAdminRenameLanguage                                                                            func(ctx context.Context, arg AdminRenameLanguageParams) error
	OverwrittenAdminRenameLinkerCategory                                                                      func(ctx context.Context, arg AdminRenameLinkerCategoryParams) error
	OverwrittenAdminReplaceSiteNewsURL                                                                        func(ctx context.Context, arg AdminReplaceSiteNewsURLParams) error
	OverwrittenAdminRestoreBlog                                                                               func(ctx context.Context, arg AdminRestoreBlogParams) error
	OverwrittenAdminRestoreComment                                                                            func(ctx context.Context, arg AdminRestoreCommentParams) error
	OverwrittenAdminRestoreImagepost                                                                          func(ctx context.Context, arg AdminRestoreImagepostParams) error
	OverwrittenAdminRestoreLink                                                                               func(ctx context.Context, arg AdminRestoreLinkParams) error
	OverwrittenAdminRestoreUser                                                                               func(ctx context.Context, idusers int32) error
	OverwrittenAdminRestoreWriting                                                                            func(ctx context.Context, arg AdminRestoreWritingParams) error
	OverwrittenAdminScrubBlog                                                                                 func(ctx context.Context, arg AdminScrubBlogParams) error
	OverwrittenAdminScrubComment                                                                              func(ctx context.Context, arg AdminScrubCommentParams) error
	OverwrittenAdminScrubImagepost                                                                            func(ctx context.Context, idimagepost int32) error
	OverwrittenAdminScrubLink                                                                                 func(ctx context.Context, arg AdminScrubLinkParams) error
	OverwrittenAdminScrubUser                                                                                 func(ctx context.Context, arg AdminScrubUserParams) error
	OverwrittenAdminScrubWriting                                                                              func(ctx context.Context, arg AdminScrubWritingParams) error
	OverwrittenAdminSetAnnouncementActive                                                                     func(ctx context.Context, arg AdminSetAnnouncementActiveParams) error
	OverwrittenAdminSetTemplateOverride                                                                       func(ctx context.Context, arg AdminSetTemplateOverrideParams) error
	OverwrittenAdminUpdateBannedIp                                                                            func(ctx context.Context, arg AdminUpdateBannedIpParams) error
	OverwrittenAdminUpdateFAQQuestionAnswer                                                                   func(ctx context.Context, arg AdminUpdateFAQQuestionAnswerParams) error
	OverwrittenAdminUpdateGrantActive                                                                         func(ctx context.Context, arg AdminUpdateGrantActiveParams) error
	OverwrittenAdminUpdateImageBoard                                                                          func(ctx context.Context, arg AdminUpdateImageBoardParams) error
	OverwrittenAdminUpdateImagePost                                                                           func(ctx context.Context, arg AdminUpdateImagePostParams) error
	OverwrittenAdminUpdateLinkerCategorySortOrder                                                             func(ctx context.Context, arg AdminUpdateLinkerCategorySortOrderParams) error
	OverwrittenAdminUpdateLinkerItem                                                                          func(ctx context.Context, arg AdminUpdateLinkerItemParams) error
	OverwrittenAdminUpdateLinkerQueuedItem                                                                    func(ctx context.Context, arg AdminUpdateLinkerQueuedItemParams) error
	OverwrittenAdminUpdateRequestStatus                                                                       func(ctx context.Context, arg AdminUpdateRequestStatusParams) error
	OverwrittenAdminUpdateRole                                                                                func(ctx context.Context, arg AdminUpdateRoleParams) error
	OverwrittenAdminUpdateRolePublicProfileAllowed                                                            func(ctx context.Context, arg AdminUpdateRolePublicProfileAllowedParams) error
	OverwrittenAdminUpdateUserEmail                                                                           func(ctx context.Context, arg AdminUpdateUserEmailParams) error
	OverwrittenAdminUpdateUserRole                                                                            func(ctx context.Context, arg AdminUpdateUserRoleParams) error
	OverwrittenAdminUpdateUsernameByID                                                                        func(ctx context.Context, arg AdminUpdateUsernameByIDParams) error
	OverwrittenAdminUpdateWritingCategory                                                                     func(ctx context.Context, arg AdminUpdateWritingCategoryParams) error
	OverwrittenAdminUserPostCounts                                                                            func(ctx context.Context) ([]*AdminUserPostCountsRow, error)
	OverwrittenAdminUserPostCountsByID                                                                        func(ctx context.Context, idusers int32) (*AdminUserPostCountsByIDRow, error)
	OverwrittenAdminWordListWithCounts                                                                        func(ctx context.Context, arg AdminWordListWithCountsParams) ([]*AdminWordListWithCountsRow, error)
	OverwrittenAdminWordListWithCountsByPrefix                                                                func(ctx context.Context, arg AdminWordListWithCountsByPrefixParams) ([]*AdminWordListWithCountsByPrefixRow, error)
	OverwrittenAdminWritingCategoryCounts                                                                     func(ctx context.Context) ([]*AdminWritingCategoryCountsRow, error)
	OverwrittenCheckUserHasGrant                                                                              func(ctx context.Context, arg CheckUserHasGrantParams) (bool, error)
	OverwrittenClearUnreadContentPrivateLabelExceptUser                                                       func(ctx context.Context, arg ClearUnreadContentPrivateLabelExceptUserParams) error
	OverwrittenCreateBlogEntryForWriter                                                                       func(ctx context.Context, arg CreateBlogEntryForWriterParams) (int64, error)
	OverwrittenCreateBookmarksForLister                                                                       func(ctx context.Context, arg CreateBookmarksForListerParams) error
	OverwrittenCreateCommentInSectionForCommenter                                                             func(ctx context.Context, arg CreateCommentInSectionForCommenterParams) (int64, error)
	OverwrittenCreateFAQQuestionForWriter                                                                     func(ctx context.Context, arg CreateFAQQuestionForWriterParams) error
	OverwrittenCreateForumThreadForPoster                                                                     func(ctx context.Context, arg CreateForumThreadForPosterParams) (int64, error)
	OverwrittenCreateForumTopicForPoster                                                                      func(ctx context.Context, arg CreateForumTopicForPosterParams) (int64, error)
	OverwrittenCreateGrant                                                                                    func(ctx context.Context, arg CreateGrantParams) error
	OverwrittenCreateImagePostForPoster                                                                       func(ctx context.Context, arg CreateImagePostForPosterParams) (int64, error)
	OverwrittenCreateLinkerQueuedItemForWriter                                                                func(ctx context.Context, arg CreateLinkerQueuedItemForWriterParams) error
	OverwrittenCreateNewsPostForWriter                                                                        func(ctx context.Context, arg CreateNewsPostForWriterParams) (int64, error)
	OverwrittenCreatePasswordResetForUser                                                                     func(ctx context.Context, arg CreatePasswordResetForUserParams) error
	OverwrittenCreateUploadedImageForUploader                                                                 func(ctx context.Context, arg CreateUploadedImageForUploaderParams) (int64, error)
	OverwrittenCreateWritingForWriter                                                                         func(ctx context.Context, arg CreateWritingForWriterParams) (int64, error)
	OverwrittenDeactivateNewsPost                                                                             func(ctx context.Context, idsitenews int32) error
	OverwrittenDeleteGrantByProperties                                                                        func(ctx context.Context, arg DeleteGrantByPropertiesParams) error
	OverwrittenDeleteGrantsByRoleID                                                                           func(ctx context.Context, roleID sql.NullInt32) error
	OverwrittenDeleteNotificationForLister                                                                    func(ctx context.Context, arg DeleteNotificationForListerParams) error
	OverwrittenDeleteSubscriptionByIDForSubscriber                                                            func(ctx context.Context, arg DeleteSubscriptionByIDForSubscriberParams) error
	OverwrittenDeleteSubscriptionForSubscriber                                                                func(ctx context.Context, arg DeleteSubscriptionForSubscriberParams) error
	OverwrittenDeleteUserEmailForOwner                                                                        func(ctx context.Context, arg DeleteUserEmailForOwnerParams) error
	OverwrittenDeleteUserLanguagesForUser                                                                     func(ctx context.Context, userID int32) error
	OverwrittenGetActiveAnnouncementWithNewsForLister                                                         func(ctx context.Context, arg GetActiveAnnouncementWithNewsForListerParams) (*GetActiveAnnouncementWithNewsForListerRow, error)
	OverwrittenGetAdministratorUserRole                                                                       func(ctx context.Context, usersIdusers int32) (*UserRole, error)
	OverwrittenGetAllAnsweredFAQWithFAQCategoriesForUser                                                      func(ctx context.Context, arg GetAllAnsweredFAQWithFAQCategoriesForUserParams) ([]*GetAllAnsweredFAQWithFAQCategoriesForUserRow, error)
	OverwrittenGetAllCommentsForIndex                                                                         func(ctx context.Context) ([]*GetAllCommentsForIndexRow, error)
	OverwrittenGetAllForumCategories                                                                          func(ctx context.Context, arg GetAllForumCategoriesParams) ([]*Forumcategory, error)
	OverwrittenGetAllForumCategoriesWithSubcategoryCount                                                      func(ctx context.Context, arg GetAllForumCategoriesWithSubcategoryCountParams) ([]*GetAllForumCategoriesWithSubcategoryCountRow, error)
	OverwrittenGetAllForumThreadsWithTopic                                                                    func(ctx context.Context) ([]*GetAllForumThreadsWithTopicRow, error)
	OverwrittenGetAllForumTopics                                                                              func(ctx context.Context, arg GetAllForumTopicsParams) ([]*Forumtopic, error)
	OverwrittenGetAllForumTopicsByCategoryIdForUserWithLastPosterName                                         func(ctx context.Context, arg GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams) ([]*GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow, error)
	OverwrittenGetAllImagePostsForIndex                                                                       func(ctx context.Context) ([]*GetAllImagePostsForIndexRow, error)
	OverwrittenGetAllLinkerCategories                                                                         func(ctx context.Context) ([]*LinkerCategory, error)
	OverwrittenGetAllLinkerCategoriesForUser                                                                  func(ctx context.Context, arg GetAllLinkerCategoriesForUserParams) ([]*LinkerCategory, error)
	OverwrittenGetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending                    func(ctx context.Context, arg GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams) ([]*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow, error)
	OverwrittenGetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUser             func(ctx context.Context, arg GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserParams) ([]*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserRow, error)
	OverwrittenGetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated    func(ctx context.Context, arg GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedParams) ([]*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow, error)
	OverwrittenGetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow func(ctx context.Context, arg GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRowParams) ([]*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRowRow, error)
	OverwrittenGetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginated           func(ctx context.Context, arg GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginatedParams) ([]*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginatedRow, error)
	OverwrittenGetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails                                        func(ctx context.Context) ([]*GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow, error)
	OverwrittenGetAllLinkersForIndex                                                                          func(ctx context.Context) ([]*GetAllLinkersForIndexRow, error)
	OverwrittenGetAllSiteNewsForIndex                                                                         func(ctx context.Context) ([]*GetAllSiteNewsForIndexRow, error)
	OverwrittenGetAllWritingsByAuthorForLister                                                                func(ctx context.Context, arg GetAllWritingsByAuthorForListerParams) ([]*GetAllWritingsByAuthorForListerRow, error)
	OverwrittenGetAllWritingsForIndex                                                                         func(ctx context.Context) ([]*GetAllWritingsForIndexRow, error)
	OverwrittenGetBlogEntryForListerByID                                                                      func(ctx context.Context, arg GetBlogEntryForListerByIDParams) (*GetBlogEntryForListerByIDRow, error)
	OverwrittenGetBookmarksForUser                                                                            func(ctx context.Context, usersIdusers int32) (*GetBookmarksForUserRow, error)
	OverwrittenGetCommentById                                                                                 func(ctx context.Context, idcomments int32) (*Comment, error)
	OverwrittenGetCommentByIdForUser                                                                          func(ctx context.Context, arg GetCommentByIdForUserParams) (*GetCommentByIdForUserRow, error)
	OverwrittenGetCommentsByIdsForUserWithThreadInfo                                                          func(ctx context.Context, arg GetCommentsByIdsForUserWithThreadInfoParams) ([]*GetCommentsByIdsForUserWithThreadInfoRow, error)
	OverwrittenGetCommentsBySectionThreadIdForUser                                                            func(ctx context.Context, arg GetCommentsBySectionThreadIdForUserParams) ([]*GetCommentsBySectionThreadIdForUserRow, error)
	OverwrittenGetCommentsByThreadIdForUser                                                                   func(ctx context.Context, arg GetCommentsByThreadIdForUserParams) ([]*GetCommentsByThreadIdForUserRow, error)
	OverwrittenGetContentReadMarker                                                                           func(ctx context.Context, arg GetContentReadMarkerParams) (*GetContentReadMarkerRow, error)
	OverwrittenGetExternalLink                                                                                func(ctx context.Context, url string) (*ExternalLink, error)
	OverwrittenGetExternalLinkByID                                                                            func(ctx context.Context, id int32) (*ExternalLink, error)
	OverwrittenGetFAQAnsweredQuestions                                                                        func(ctx context.Context, arg GetFAQAnsweredQuestionsParams) ([]*Faq, error)
	OverwrittenGetFAQByID                                                                                     func(ctx context.Context, arg GetFAQByIDParams) (*Faq, error)
	OverwrittenGetFAQQuestionsByCategory                                                                      func(ctx context.Context, arg GetFAQQuestionsByCategoryParams) ([]*Faq, error)
	OverwrittenGetFAQRevisionsForAdmin                                                                        func(ctx context.Context, faqID int32) ([]*FaqRevision, error)
	OverwrittenGetForumCategoryById                                                                           func(ctx context.Context, arg GetForumCategoryByIdParams) (*Forumcategory, error)
	OverwrittenGetForumThreadIdByNewsPostId                                                                   func(ctx context.Context, idsitenews int32) (*GetForumThreadIdByNewsPostIdRow, error)
	OverwrittenGetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText                     func(ctx context.Context, arg GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error)
	OverwrittenGetForumTopicById                                                                              func(ctx context.Context, idforumtopic int32) (*Forumtopic, error)
	OverwrittenGetForumTopicByIdForUser                                                                       func(ctx context.Context, arg GetForumTopicByIdForUserParams) (*GetForumTopicByIdForUserRow, error)
	OverwrittenGetForumTopicIdByThreadId                                                                      func(ctx context.Context, idforumthread int32) (int32, error)
	OverwrittenGetForumTopicsByCategoryId                                                                     func(ctx context.Context, arg GetForumTopicsByCategoryIdParams) ([]*Forumtopic, error)
	OverwrittenGetForumTopicsForUser                                                                          func(ctx context.Context, arg GetForumTopicsForUserParams) ([]*GetForumTopicsForUserRow, error)
	OverwrittenGetGrantsByRoleID                                                                              func(ctx context.Context, roleID sql.NullInt32) ([]*Grant, error)
	OverwrittenGetImageBoardById                                                                              func(ctx context.Context, idimageboard int32) (*Imageboard, error)
	OverwrittenGetImagePostByIDForLister                                                                      func(ctx context.Context, arg GetImagePostByIDForListerParams) (*GetImagePostByIDForListerRow, error)
	OverwrittenGetImagePostInfoByPath                                                                         func(ctx context.Context, arg GetImagePostInfoByPathParams) (*GetImagePostInfoByPathRow, error)
	OverwrittenGetImagePostsByUserDescending                                                                  func(ctx context.Context, arg GetImagePostsByUserDescendingParams) ([]*GetImagePostsByUserDescendingRow, error)
	OverwrittenGetImagePostsByUserDescendingAll                                                               func(ctx context.Context, arg GetImagePostsByUserDescendingAllParams) ([]*GetImagePostsByUserDescendingAllRow, error)
	OverwrittenGetLatestAnnouncementByNewsID                                                                  func(ctx context.Context, siteNewsID int32) (*SiteAnnouncement, error)
	OverwrittenGetLinkerCategoriesWithCount                                                                   func(ctx context.Context) ([]*GetLinkerCategoriesWithCountRow, error)
	OverwrittenGetLinkerCategoryById                                                                          func(ctx context.Context, id int32) (*LinkerCategory, error)
	OverwrittenGetLinkerCategoryLinkCounts                                                                    func(ctx context.Context) ([]*GetLinkerCategoryLinkCountsRow, error)
	OverwrittenGetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending                                  func(ctx context.Context, id int32) (*GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error)
	OverwrittenGetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser                           func(ctx context.Context, arg GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams) (*GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error)
	OverwrittenGetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending                                func(ctx context.Context, linkerids []int32) ([]*GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow, error)
	OverwrittenGetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser                         func(ctx context.Context, arg GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUserParams) ([]*GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUserRow, error)
	OverwrittenGetLinkerItemsByUserDescending                                                                 func(ctx context.Context, arg GetLinkerItemsByUserDescendingParams) ([]*GetLinkerItemsByUserDescendingRow, error)
	OverwrittenGetLinkerItemsByUserDescendingForUser                                                          func(ctx context.Context, arg GetLinkerItemsByUserDescendingForUserParams) ([]*GetLinkerItemsByUserDescendingForUserRow, error)
	OverwrittenGetLoginRoleForUser                                                                            func(ctx context.Context, usersIdusers int32) (int32, error)
	OverwrittenGetMaxNotificationPriority                                                                     func(ctx context.Context, userID int32) (interface{}, error)
	OverwrittenGetNewsPostByIdWithWriterIdAndThreadCommentCount                                               func(ctx context.Context, arg GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error)
	OverwrittenGetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount                                      func(ctx context.Context, arg GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams) ([]*GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow, error)
	OverwrittenGetNewsPostsWithWriterUsernameAndThreadCommentCountDescending                                  func(ctx context.Context, arg GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams) ([]*GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error)
	OverwrittenGetNotificationEmailByUserID                                                                   func(ctx context.Context, userID int32) (*UserEmail, error)
	OverwrittenGetNotificationForLister                                                                       func(ctx context.Context, arg GetNotificationForListerParams) (*Notification, error)
	OverwrittenGetPasswordResetByCode                                                                         func(ctx context.Context, arg GetPasswordResetByCodeParams) (*PendingPassword, error)
	OverwrittenGetPasswordResetByUser                                                                         func(ctx context.Context, arg GetPasswordResetByUserParams) (*PendingPassword, error)
	OverwrittenGetPendingEmailErrorCount                                                                      func(ctx context.Context, id int32) (int32, error)
	OverwrittenGetPermissionsByUserID                                                                         func(ctx context.Context, usersIdusers int32) ([]*GetPermissionsByUserIDRow, error)
	OverwrittenGetPermissionsWithUsers                                                                        func(ctx context.Context, arg GetPermissionsWithUsersParams) ([]*GetPermissionsWithUsersRow, error)
	OverwrittenGetPreferenceForLister                                                                         func(ctx context.Context, listerID int32) (*Preference, error)
	OverwrittenGetPublicProfileRoleForUser                                                                    func(ctx context.Context, usersIdusers int32) (int32, error)
	OverwrittenGetPublicWritings                                                                              func(ctx context.Context, arg GetPublicWritingsParams) ([]*Writing, error)
	OverwrittenGetRoleByName                                                                                  func(ctx context.Context, name string) (*Role, error)
	OverwrittenGetThreadBySectionThreadIDForReplier                                                           func(ctx context.Context, arg GetThreadBySectionThreadIDForReplierParams) (*Forumthread, error)
	OverwrittenGetThreadLastPosterAndPerms                                                                    func(ctx context.Context, arg GetThreadLastPosterAndPermsParams) (*GetThreadLastPosterAndPermsRow, error)
	OverwrittenGetUnreadNotificationCountForLister                                                            func(ctx context.Context, listerID int32) (int64, error)
	OverwrittenGetUserEmailByCode                                                                             func(ctx context.Context, lastVerificationCode sql.NullString) (*UserEmail, error)
	OverwrittenGetUserEmailByEmail                                                                            func(ctx context.Context, email string) (*UserEmail, error)
	OverwrittenGetUserEmailByID                                                                               func(ctx context.Context, id int32) (*UserEmail, error)
	OverwrittenGetUserLanguages                                                                               func(ctx context.Context, usersIdusers int32) ([]*UserLanguage, error)
	OverwrittenGetUserRole                                                                                    func(ctx context.Context, usersIdusers int32) (string, error)
	OverwrittenGetUserRoles                                                                                   func(ctx context.Context) ([]*GetUserRolesRow, error)
	OverwrittenGetVerifiedUserEmails                                                                          func(ctx context.Context) ([]*GetVerifiedUserEmailsRow, error)
	OverwrittenGetWritingCategoryById                                                                         func(ctx context.Context, idwritingcategory int32) (*WritingCategory, error)
	OverwrittenGetWritingForListerByID                                                                        func(ctx context.Context, arg GetWritingForListerByIDParams) (*GetWritingForListerByIDRow, error)
	OverwrittenInsertAdminUserComment                                                                         func(ctx context.Context, arg InsertAdminUserCommentParams) error
	OverwrittenInsertAuditLog                                                                                 func(ctx context.Context, arg InsertAuditLogParams) error
	OverwrittenInsertEmailPreferenceForLister                                                                 func(ctx context.Context, arg InsertEmailPreferenceForListerParams) error
	OverwrittenInsertFAQQuestionForWriter                                                                     func(ctx context.Context, arg InsertFAQQuestionForWriterParams) (sql.Result, error)
	OverwrittenInsertFAQRevisionForUser                                                                       func(ctx context.Context, arg InsertFAQRevisionForUserParams) error
	OverwrittenInsertPassword                                                                                 func(ctx context.Context, arg InsertPasswordParams) error
	OverwrittenInsertPendingEmail                                                                             func(ctx context.Context, arg InsertPendingEmailParams) error
	OverwrittenInsertPreferenceForLister                                                                      func(ctx context.Context, arg InsertPreferenceForListerParams) error
	OverwrittenInsertSubscription                                                                             func(ctx context.Context, arg InsertSubscriptionParams) error
	OverwrittenInsertUserEmail                                                                                func(ctx context.Context, arg InsertUserEmailParams) error
	OverwrittenInsertUserLang                                                                                 func(ctx context.Context, arg InsertUserLangParams) error
	OverwrittenInsertWriting                                                                                  func(ctx context.Context, arg InsertWritingParams) (int64, error)
	OverwrittenLatestAdminUserComment                                                                         func(ctx context.Context, usersIdusers int32) (*AdminUserComment, error)
	OverwrittenLinkerSearchFirst                                                                              func(ctx context.Context, arg LinkerSearchFirstParams) ([]int32, error)
	OverwrittenLinkerSearchNext                                                                               func(ctx context.Context, arg LinkerSearchNextParams) ([]int32, error)
	OverwrittenListActiveBans                                                                                 func(ctx context.Context) ([]*BannedIp, error)
	OverwrittenListAdminUserComments                                                                          func(ctx context.Context, usersIdusers int32) ([]*AdminUserComment, error)
	OverwrittenListBannedIps                                                                                  func(ctx context.Context) ([]*BannedIp, error)
	OverwrittenListBlogEntriesByAuthorForLister                                                               func(ctx context.Context, arg ListBlogEntriesByAuthorForListerParams) ([]*ListBlogEntriesByAuthorForListerRow, error)
	OverwrittenListBlogEntriesByIDsForLister                                                                  func(ctx context.Context, arg ListBlogEntriesByIDsForListerParams) ([]*ListBlogEntriesByIDsForListerRow, error)
	OverwrittenListBlogEntriesForLister                                                                       func(ctx context.Context, arg ListBlogEntriesForListerParams) ([]*ListBlogEntriesForListerRow, error)
	OverwrittenListBlogIDsBySearchWordFirstForLister                                                          func(ctx context.Context, arg ListBlogIDsBySearchWordFirstForListerParams) ([]int32, error)
	OverwrittenListBlogIDsBySearchWordNextForLister                                                           func(ctx context.Context, arg ListBlogIDsBySearchWordNextForListerParams) ([]int32, error)
	OverwrittenListBloggersForLister                                                                          func(ctx context.Context, arg ListBloggersForListerParams) ([]*ListBloggersForListerRow, error)
	OverwrittenListBloggersSearchForLister                                                                    func(ctx context.Context, arg ListBloggersSearchForListerParams) ([]*ListBloggersSearchForListerRow, error)
	OverwrittenListBoardsByParentIDForLister                                                                  func(ctx context.Context, arg ListBoardsByParentIDForListerParams) ([]*Imageboard, error)
	OverwrittenListBoardsForLister                                                                            func(ctx context.Context, arg ListBoardsForListerParams) ([]*Imageboard, error)
	OverwrittenListCommentIDsBySearchWordFirstForListerInRestrictedTopic                                      func(ctx context.Context, arg ListCommentIDsBySearchWordFirstForListerInRestrictedTopicParams) ([]int32, error)
	OverwrittenListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic                                   func(ctx context.Context, arg ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopicParams) ([]int32, error)
	OverwrittenListCommentIDsBySearchWordNextForListerInRestrictedTopic                                       func(ctx context.Context, arg ListCommentIDsBySearchWordNextForListerInRestrictedTopicParams) ([]int32, error)
	OverwrittenListCommentIDsBySearchWordNextForListerNotInRestrictedTopic                                    func(ctx context.Context, arg ListCommentIDsBySearchWordNextForListerNotInRestrictedTopicParams) ([]int32, error)
	OverwrittenListContentLabelStatus                                                                         func(ctx context.Context, arg ListContentLabelStatusParams) ([]*ListContentLabelStatusRow, error)
	OverwrittenListContentPrivateLabels                                                                       func(ctx context.Context, arg ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error)
	OverwrittenListContentPublicLabels                                                                        func(ctx context.Context, arg ListContentPublicLabelsParams) ([]*ListContentPublicLabelsRow, error)
	OverwrittenListEffectiveRoleIDsByUserID                                                                   func(ctx context.Context, usersIdusers int32) ([]int32, error)
	OverwrittenListForumcategoryPath                                                                          func(ctx context.Context, categoryID int32) ([]*ListForumcategoryPathRow, error)
	OverwrittenListGrants                                                                                     func(ctx context.Context) ([]*Grant, error)
	OverwrittenListGrantsByUserID                                                                             func(ctx context.Context, userID sql.NullInt32) ([]*Grant, error)
	OverwrittenListImagePostsByBoardForLister                                                                 func(ctx context.Context, arg ListImagePostsByBoardForListerParams) ([]*ListImagePostsByBoardForListerRow, error)
	OverwrittenListImagePostsByPosterForLister                                                                func(ctx context.Context, arg ListImagePostsByPosterForListerParams) ([]*ListImagePostsByPosterForListerRow, error)
	OverwrittenListImageboardPath                                                                             func(ctx context.Context, boardID int32) ([]*ListImageboardPathRow, error)
	OverwrittenListLinkerCategoryPath                                                                         func(ctx context.Context, categoryID int32) ([]*ListLinkerCategoryPathRow, error)
	OverwrittenListNotificationsForLister                                                                     func(ctx context.Context, arg ListNotificationsForListerParams) ([]*Notification, error)
	OverwrittenListPrivateTopicParticipantsByTopicIDForUser                                                   func(ctx context.Context, arg ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*ListPrivateTopicParticipantsByTopicIDForUserRow, error)
	OverwrittenListPrivateTopicsByUserID                                                                      func(ctx context.Context, userID sql.NullInt32) ([]*ListPrivateTopicsByUserIDRow, error)
	OverwrittenListPublicWritingsByUserForLister                                                              func(ctx context.Context, arg ListPublicWritingsByUserForListerParams) ([]*ListPublicWritingsByUserForListerRow, error)
	OverwrittenListPublicWritingsInCategoryForLister                                                          func(ctx context.Context, arg ListPublicWritingsInCategoryForListerParams) ([]*ListPublicWritingsInCategoryForListerRow, error)
	OverwrittenListSiteNewsSearchFirstForLister                                                               func(ctx context.Context, arg ListSiteNewsSearchFirstForListerParams) ([]int32, error)
	OverwrittenListSiteNewsSearchNextForLister                                                                func(ctx context.Context, arg ListSiteNewsSearchNextForListerParams) ([]int32, error)
	OverwrittenListSubscribersForPattern                                                                      func(ctx context.Context, arg ListSubscribersForPatternParams) ([]int32, error)
	OverwrittenListSubscribersForPatterns                                                                     func(ctx context.Context, arg ListSubscribersForPatternsParams) ([]int32, error)
	OverwrittenListSubscriptionsByUser                                                                        func(ctx context.Context, usersIdusers int32) ([]*ListSubscriptionsByUserRow, error)
	OverwrittenListUnreadNotificationsForLister                                                               func(ctx context.Context, arg ListUnreadNotificationsForListerParams) ([]*Notification, error)
	OverwrittenListUploadedImagesByUserForLister                                                              func(ctx context.Context, arg ListUploadedImagesByUserForListerParams) ([]*UploadedImage, error)
	OverwrittenListUserEmailsForLister                                                                        func(ctx context.Context, arg ListUserEmailsForListerParams) ([]*UserEmail, error)
	OverwrittenListUsersWithRoles                                                                             func(ctx context.Context) ([]*ListUsersWithRolesRow, error)
	OverwrittenListWritersForLister                                                                           func(ctx context.Context, arg ListWritersForListerParams) ([]*ListWritersForListerRow, error)
	OverwrittenListWritersSearchForLister                                                                     func(ctx context.Context, arg ListWritersSearchForListerParams) ([]*ListWritersSearchForListerRow, error)
	OverwrittenListWritingCategoriesForLister                                                                 func(ctx context.Context, arg ListWritingCategoriesForListerParams) ([]*WritingCategory, error)
	OverwrittenListWritingSearchFirstForLister                                                                func(ctx context.Context, arg ListWritingSearchFirstForListerParams) ([]int32, error)
	OverwrittenListWritingSearchNextForLister                                                                 func(ctx context.Context, arg ListWritingSearchNextForListerParams) ([]int32, error)
	OverwrittenListWritingcategoryPath                                                                        func(ctx context.Context, categoryID int32) ([]*ListWritingcategoryPathRow, error)
	OverwrittenListWritingsByIDsForLister                                                                     func(ctx context.Context, arg ListWritingsByIDsForListerParams) ([]*ListWritingsByIDsForListerRow, error)
	OverwrittenRemoveContentLabelStatus                                                                       func(ctx context.Context, arg RemoveContentLabelStatusParams) error
	OverwrittenRemoveContentPrivateLabel                                                                      func(ctx context.Context, arg RemoveContentPrivateLabelParams) error
	OverwrittenRemoveContentPublicLabel                                                                       func(ctx context.Context, arg RemoveContentPublicLabelParams) error
	OverwrittenSetNotificationPriorityForLister                                                               func(ctx context.Context, arg SetNotificationPriorityForListerParams) error
	OverwrittenSetNotificationReadForLister                                                                   func(ctx context.Context, arg SetNotificationReadForListerParams) error
	OverwrittenSetNotificationUnreadForLister                                                                 func(ctx context.Context, arg SetNotificationUnreadForListerParams) error
	OverwrittenSetVerificationCodeForLister                                                                   func(ctx context.Context, arg SetVerificationCodeForListerParams) error
	OverwrittenSystemAddToBlogsSearch                                                                         func(ctx context.Context, arg SystemAddToBlogsSearchParams) error
	OverwrittenSystemAddToForumCommentSearch                                                                  func(ctx context.Context, arg SystemAddToForumCommentSearchParams) error
	OverwrittenSystemAddToForumWritingSearch                                                                  func(ctx context.Context, arg SystemAddToForumWritingSearchParams) error
	OverwrittenSystemAddToImagePostSearch                                                                     func(ctx context.Context, arg SystemAddToImagePostSearchParams) error
	OverwrittenSystemAddToLinkerSearch                                                                        func(ctx context.Context, arg SystemAddToLinkerSearchParams) error
	OverwrittenSystemAddToSiteNewsSearch                                                                      func(ctx context.Context, arg SystemAddToSiteNewsSearchParams) error
	OverwrittenSystemAssignBlogEntryThreadID                                                                  func(ctx context.Context, arg SystemAssignBlogEntryThreadIDParams) error
	OverwrittenSystemAssignImagePostThreadID                                                                  func(ctx context.Context, arg SystemAssignImagePostThreadIDParams) error
	OverwrittenSystemAssignLinkerThreadID                                                                     func(ctx context.Context, arg SystemAssignLinkerThreadIDParams) error
	OverwrittenSystemAssignNewsThreadID                                                                       func(ctx context.Context, arg SystemAssignNewsThreadIDParams) error
	OverwrittenSystemAssignWritingThreadID                                                                    func(ctx context.Context, arg SystemAssignWritingThreadIDParams) error
	OverwrittenSystemCheckGrant                                                                               func(ctx context.Context, arg SystemCheckGrantParams) (int32, error)
	OverwrittenSystemCheckRoleGrant                                                                           func(ctx context.Context, arg SystemCheckRoleGrantParams) (int32, error)
	OverwrittenSystemClearContentLabelStatus                                                                  func(ctx context.Context, arg SystemClearContentLabelStatusParams) error
	OverwrittenSystemClearContentPrivateLabel                                                                 func(ctx context.Context, arg SystemClearContentPrivateLabelParams) error
	OverwrittenSystemCountDeadLetters                                                                         func(ctx context.Context) (int64, error)
	OverwrittenSystemCountLanguages                                                                           func(ctx context.Context) (int64, error)
	OverwrittenSystemCountRecentLoginAttempts                                                                 func(ctx context.Context, arg SystemCountRecentLoginAttemptsParams) (int64, error)
	OverwrittenSystemCreateGrant                                                                              func(ctx context.Context, arg SystemCreateGrantParams) (int64, error)
	OverwrittenSystemCreateNotification                                                                       func(ctx context.Context, arg SystemCreateNotificationParams) error
	OverwrittenSystemCreateSearchWord                                                                         func(ctx context.Context, word string) (int64, error)
	OverwrittenSystemCreateThread                                                                             func(ctx context.Context, forumtopicIdforumtopic int32) (int64, error)
	OverwrittenSystemCreateUserRole                                                                           func(ctx context.Context, arg SystemCreateUserRoleParams) error
	OverwrittenSystemDeleteBlogsSearch                                                                        func(ctx context.Context) error
	OverwrittenSystemDeleteCommentsSearch                                                                     func(ctx context.Context) error
	OverwrittenSystemDeleteDeadLetter                                                                         func(ctx context.Context, id int32) error
	OverwrittenSystemDeleteImagePostSearch                                                                    func(ctx context.Context) error
	OverwrittenSystemDeleteLinkerSearch                                                                       func(ctx context.Context) error
	OverwrittenSystemDeletePasswordReset                                                                      func(ctx context.Context, id int32) error
	OverwrittenSystemDeletePasswordResetsByUser                                                               func(ctx context.Context, userID int32) (sql.Result, error)
	OverwrittenSystemDeleteSessionByID                                                                        func(ctx context.Context, sessionID string) error
	OverwrittenSystemDeleteSiteNewsSearch                                                                     func(ctx context.Context) error
	OverwrittenSystemDeleteUserEmailsByEmailExceptID                                                          func(ctx context.Context, arg SystemDeleteUserEmailsByEmailExceptIDParams) error
	OverwrittenSystemDeleteWritingSearch                                                                      func(ctx context.Context) error
	OverwrittenSystemDeleteWritingSearchByWritingID                                                           func(ctx context.Context, writingID int32) error
	OverwrittenSystemGetAllBlogsForIndex                                                                      func(ctx context.Context) ([]*SystemGetAllBlogsForIndexRow, error)
	OverwrittenSystemGetBlogEntryByID                                                                         func(ctx context.Context, idblogs int32) (*SystemGetBlogEntryByIDRow, error)
	OverwrittenSystemGetFAQQuestions                                                                          func(ctx context.Context) ([]*Faq, error)
	OverwrittenSystemGetForumTopicByTitle                                                                     func(ctx context.Context, title sql.NullString) (*Forumtopic, error)
	OverwrittenSystemGetLanguageIDByName                                                                      func(ctx context.Context, nameof sql.NullString) (int32, error)
	OverwrittenSystemGetLastNotificationForRecipientByMessage                                                 func(ctx context.Context, arg SystemGetLastNotificationForRecipientByMessageParams) (*Notification, error)
	OverwrittenSystemGetLogin                                                                                 func(ctx context.Context, username sql.NullString) (*SystemGetLoginRow, error)
	OverwrittenSystemGetNewsPostByID                                                                          func(ctx context.Context, idsitenews int32) (int32, error)
	OverwrittenSystemGetSearchWordByWordLowercased                                                            func(ctx context.Context, lcase string) (*Searchwordlist, error)
	OverwrittenSystemGetTemplateOverride                                                                      func(ctx context.Context, name string) (string, error)
	OverwrittenSystemGetUserByEmail                                                                           func(ctx context.Context, email string) (*SystemGetUserByEmailRow, error)
	OverwrittenSystemGetUserByID                                                                              func(ctx context.Context, idusers int32) (*SystemGetUserByIDRow, error)
	OverwrittenSystemGetUserByUsername                                                                        func(ctx context.Context, username sql.NullString) (*SystemGetUserByUsernameRow, error)
	OverwrittenSystemGetWritingByID                                                                           func(ctx context.Context, idwriting int32) (int32, error)
	OverwrittenSystemIncrementPendingEmailError                                                               func(ctx context.Context, id int32) error
	OverwrittenSystemInsertDeadLetter                                                                         func(ctx context.Context, message string) error
	OverwrittenSystemInsertLoginAttempt                                                                       func(ctx context.Context, arg SystemInsertLoginAttemptParams) error
	OverwrittenSystemInsertSession                                                                            func(ctx context.Context, arg SystemInsertSessionParams) error
	OverwrittenSystemInsertUser                                                                               func(ctx context.Context, username sql.NullString) (int64, error)
	OverwrittenSystemLatestDeadLetter                                                                         func(ctx context.Context) (interface{}, error)
	OverwrittenSystemListAllUsers                                                                             func(ctx context.Context) ([]*SystemListAllUsersRow, error)
	OverwrittenSystemListBoardsByParentID                                                                     func(ctx context.Context, arg SystemListBoardsByParentIDParams) ([]*Imageboard, error)
	OverwrittenSystemListCommentsByThreadID                                                                   func(ctx context.Context, forumthreadID int32) ([]*SystemListCommentsByThreadIDRow, error)
	OverwrittenSystemListDeadLetters                                                                          func(ctx context.Context, limit int32) ([]*DeadLetter, error)
	OverwrittenSystemListLanguages                                                                            func(ctx context.Context) ([]*Language, error)
	OverwrittenSystemListPendingEmails                                                                        func(ctx context.Context, arg SystemListPendingEmailsParams) ([]*SystemListPendingEmailsRow, error)
	OverwrittenSystemListPublicWritingsByAuthor                                                               func(ctx context.Context, arg SystemListPublicWritingsByAuthorParams) ([]*SystemListPublicWritingsByAuthorRow, error)
	OverwrittenSystemListPublicWritingsInCategory                                                             func(ctx context.Context, arg SystemListPublicWritingsInCategoryParams) ([]*SystemListPublicWritingsInCategoryRow, error)
	OverwrittenSystemListUserInfo                                                                             func(ctx context.Context) ([]*SystemListUserInfoRow, error)
	OverwrittenSystemListVerifiedEmailsByUserID                                                               func(ctx context.Context, userID int32) ([]*UserEmail, error)
	OverwrittenSystemListWritingCategories                                                                    func(ctx context.Context, arg SystemListWritingCategoriesParams) ([]*WritingCategory, error)
	OverwrittenSystemMarkPasswordResetVerified                                                                func(ctx context.Context, id int32) error
	OverwrittenSystemMarkPendingEmailSent                                                                     func(ctx context.Context, id int32) error
	OverwrittenSystemMarkUserEmailVerified                                                                    func(ctx context.Context, arg SystemMarkUserEmailVerifiedParams) error
	OverwrittenSystemPurgeDeadLettersBefore                                                                   func(ctx context.Context, createdAt time.Time) error
	OverwrittenSystemPurgePasswordResetsBefore                                                                func(ctx context.Context, createdAt time.Time) (sql.Result, error)
	OverwrittenSystemRebuildForumTopicMetaByID                                                                func(ctx context.Context, idforumtopic int32) error
	OverwrittenSystemRegisterExternalLinkClick                                                                func(ctx context.Context, url string) error
	OverwrittenSystemSetBlogLastIndex                                                                         func(ctx context.Context, id int32) error
	OverwrittenSystemSetCommentLastIndex                                                                      func(ctx context.Context, idcomments int32) error
	OverwrittenSystemSetForumTopicHandlerByID                                                                 func(ctx context.Context, arg SystemSetForumTopicHandlerByIDParams) error
	OverwrittenSystemSetImagePostLastIndex                                                                    func(ctx context.Context, idimagepost int32) error
	OverwrittenSystemSetLinkerLastIndex                                                                       func(ctx context.Context, id int32) error
	OverwrittenSystemSetSiteNewsLastIndex                                                                     func(ctx context.Context, idsitenews int32) error
	OverwrittenSystemSetWritingLastIndex                                                                      func(ctx context.Context, idwriting int32) error
	OverwrittenUpdateAutoSubscribeRepliesForLister                                                            func(ctx context.Context, arg UpdateAutoSubscribeRepliesForListerParams) error
	OverwrittenUpdateBlogEntryForWriter                                                                       func(ctx context.Context, arg UpdateBlogEntryForWriterParams) error
	OverwrittenUpdateBookmarksForLister                                                                       func(ctx context.Context, arg UpdateBookmarksForListerParams) error
	OverwrittenUpdateCommentForEditor                                                                         func(ctx context.Context, arg UpdateCommentForEditorParams) error
	OverwrittenUpdateEmailForumUpdatesForLister                                                               func(ctx context.Context, arg UpdateEmailForumUpdatesForListerParams) error
	OverwrittenUpdateNewsPostForWriter                                                                        func(ctx context.Context, arg UpdateNewsPostForWriterParams) error
	OverwrittenUpdatePreferenceForLister                                                                      func(ctx context.Context, arg UpdatePreferenceForListerParams) error
	OverwrittenUpdatePublicProfileEnabledAtForUser                                                            func(ctx context.Context, arg UpdatePublicProfileEnabledAtForUserParams) error
	OverwrittenUpdateSubscriptionByIDForSubscriber                                                            func(ctx context.Context, arg UpdateSubscriptionByIDForSubscriberParams) error
	OverwrittenUpdateTimezoneForLister                                                                        func(ctx context.Context, arg UpdateTimezoneForListerParams) error
	OverwrittenUpdateWritingForWriter                                                                         func(ctx context.Context, arg UpdateWritingForWriterParams) error
	OverwrittenUpsertContentReadMarker                                                                        func(ctx context.Context, arg UpsertContentReadMarkerParams) error
}
func (q *QuerierProxier) GetPermissionsByUserID(ctx context.Context, usersIdusers int32) ([]*GetPermissionsByUserIDRow, error) {
	if q.OverwrittenGetPermissionsByUserID == nil {
		panic("GetPermissionsByUserID not implemented")
	}
	return q.OverwrittenGetPermissionsByUserID(ctx, usersIdusers)
}

func (q *QuerierProxier) SystemCheckRoleGrant(ctx context.Context, arg SystemCheckRoleGrantParams) (int32, error) {
	if q.OverwrittenSystemCheckRoleGrant == nil {
		panic("SystemCheckRoleGrant not implemented")
	}
	return q.OverwrittenSystemCheckRoleGrant(ctx, arg)
}

func (q *QuerierProxier) SystemGetUserByID(ctx context.Context, idusers int32) (*SystemGetUserByIDRow, error) {
	if q.OverwrittenSystemGetUserByID == nil {
		panic("SystemGetUserByID not implemented")
	}
	return q.OverwrittenSystemGetUserByID(ctx, idusers)
}

func (q *QuerierProxier) CreateForumTopicForPoster(ctx context.Context, arg CreateForumTopicForPosterParams) (int64, error) {
	if q.OverwrittenCreateForumTopicForPoster == nil {
		panic("CreateForumTopicForPoster not implemented")
	}
	return q.OverwrittenCreateForumTopicForPoster(ctx, arg)
}

func (q *QuerierProxier) CreateForumThreadForPoster(ctx context.Context, arg CreateForumThreadForPosterParams) (int64, error) {
	if q.OverwrittenCreateForumThreadForPoster == nil {
		panic("CreateForumThreadForPoster not implemented")
	}
	return q.OverwrittenCreateForumThreadForPoster(ctx, arg)
}

func (q *QuerierProxier) CreateCommentInSectionForCommenter(ctx context.Context, arg CreateCommentInSectionForCommenterParams) (int64, error) {
	if q.OverwrittenCreateCommentInSectionForCommenter == nil {
		panic("CreateCommentInSectionForCommenter not implemented")
	}
	return q.OverwrittenCreateCommentInSectionForCommenter(ctx, arg)
}

func (q *QuerierProxier) SystemCreateGrant(ctx context.Context, arg SystemCreateGrantParams) (int64, error) {
	if q.OverwrittenSystemCreateGrant == nil {
		panic("SystemCreateGrant not implemented")
	}
	return q.OverwrittenSystemCreateGrant(ctx, arg)
}

func (q *QuerierProxier) SystemCheckGrant(ctx context.Context, arg SystemCheckGrantParams) (int32, error) {
	if q.OverwrittenSystemCheckGrant == nil {
		panic("SystemCheckGrant not implemented")
	}
	return q.OverwrittenSystemCheckGrant(ctx, arg)
}

func (q *QuerierProxier) ListPrivateTopicParticipantsByTopicIDForUser(ctx context.Context, arg ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
	if q.OverwrittenListPrivateTopicParticipantsByTopicIDForUser == nil {
		panic("ListPrivateTopicParticipantsByTopicIDForUser not implemented")
	}
	return q.OverwrittenListPrivateTopicParticipantsByTopicIDForUser(ctx, arg)
}

func (q *QuerierProxier) ListContentLabelStatus(ctx context.Context, arg ListContentLabelStatusParams) ([]*ListContentLabelStatusRow, error) {
	if q.OverwrittenListContentLabelStatus == nil {
		panic("ListContentLabelStatus not implemented")
	}
	return q.OverwrittenListContentLabelStatus(ctx, arg)
}
