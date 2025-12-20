package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// UnimplementedQuerier provides panic implementations for db.Querier.
type UnimplementedQuerier struct {
	t testing.TB
}

// NewUnimplementedQuerier returns a querier that fails tests for unimplemented calls.
func NewUnimplementedQuerier(t testing.TB) *UnimplementedQuerier {
	t.Helper()
	return &UnimplementedQuerier{t: t}
}

func (q *UnimplementedQuerier) fail(method string) {
	if q != nil && q.t != nil {
		q.t.Helper()
		q.t.Fatalf("UnimplementedQuerier: %s not implemented", method)
		return
	}
	panic(fmt.Sprintf("UnimplementedQuerier: %s not implemented", method))
}

func (q *UnimplementedQuerier) AddContentLabelStatus(ctx context.Context, arg db.AddContentLabelStatusParams) error {
	q.fail("AddContentLabelStatus")
	return nil
}

func (q *UnimplementedQuerier) AddContentPrivateLabel(ctx context.Context, arg db.AddContentPrivateLabelParams) error {
	q.fail("AddContentPrivateLabel")
	return nil
}

func (q *UnimplementedQuerier) AddContentPublicLabel(ctx context.Context, arg db.AddContentPublicLabelParams) error {
	q.fail("AddContentPublicLabel")
	return nil
}

func (q *UnimplementedQuerier) AdminApproveImagePost(ctx context.Context, idimagepost int32) error {
	q.fail("AdminApproveImagePost")
	return nil
}

func (q *UnimplementedQuerier) AdminArchiveBlog(ctx context.Context, arg db.AdminArchiveBlogParams) error {
	q.fail("AdminArchiveBlog")
	return nil
}

func (q *UnimplementedQuerier) AdminArchiveComment(ctx context.Context, arg db.AdminArchiveCommentParams) error {
	q.fail("AdminArchiveComment")
	return nil
}

func (q *UnimplementedQuerier) AdminArchiveImagepost(ctx context.Context, arg db.AdminArchiveImagepostParams) error {
	q.fail("AdminArchiveImagepost")
	return nil
}

func (q *UnimplementedQuerier) AdminArchiveLink(ctx context.Context, arg db.AdminArchiveLinkParams) error {
	q.fail("AdminArchiveLink")
	return nil
}

func (q *UnimplementedQuerier) AdminArchiveUser(ctx context.Context, idusers int32) error {
	q.fail("AdminArchiveUser")
	return nil
}

func (q *UnimplementedQuerier) AdminArchiveWriting(ctx context.Context, arg db.AdminArchiveWritingParams) error {
	q.fail("AdminArchiveWriting")
	return nil
}

func (q *UnimplementedQuerier) AdminCancelBannedIp(ctx context.Context, ipNet string) error {
	q.fail("AdminCancelBannedIp")
	return nil
}

func (q *UnimplementedQuerier) AdminClearExternalLinkCache(ctx context.Context, arg db.AdminClearExternalLinkCacheParams) error {
	q.fail("AdminClearExternalLinkCache")
	return nil
}

func (q *UnimplementedQuerier) AdminCompleteWordList(ctx context.Context) ([]sql.NullString, error) {
	q.fail("AdminCompleteWordList")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminCountForumCategories(ctx context.Context, arg db.AdminCountForumCategoriesParams) (int64, error) {
	q.fail("AdminCountForumCategories")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCountForumThreads(ctx context.Context) (int64, error) {
	q.fail("AdminCountForumThreads")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCountForumTopics(ctx context.Context) (int64, error) {
	q.fail("AdminCountForumTopics")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCountLinksByCategory(ctx context.Context, categoryID sql.NullInt32) (int64, error) {
	q.fail("AdminCountLinksByCategory")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCountThreadsByBoard(ctx context.Context, imageboardIdimageboard sql.NullInt32) (int64, error) {
	q.fail("AdminCountThreadsByBoard")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCountWordList(ctx context.Context) (int64, error) {
	q.fail("AdminCountWordList")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCountWordListByPrefix(ctx context.Context, prefix interface{}) (int64, error) {
	q.fail("AdminCountWordListByPrefix")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCreateFAQCategory(ctx context.Context, name sql.NullString) error {
	q.fail("AdminCreateFAQCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminCreateForumCategory(ctx context.Context, arg db.AdminCreateForumCategoryParams) (int64, error) {
	q.fail("AdminCreateForumCategory")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCreateForumTopic(ctx context.Context, arg db.AdminCreateForumTopicParams) (int64, error) {
	q.fail("AdminCreateForumTopic")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCreateGrant(ctx context.Context, arg db.AdminCreateGrantParams) (int64, error) {
	q.fail("AdminCreateGrant")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminCreateImageBoard(ctx context.Context, arg db.AdminCreateImageBoardParams) error {
	q.fail("AdminCreateImageBoard")
	return nil
}

func (q *UnimplementedQuerier) AdminCreateLanguage(ctx context.Context, nameof sql.NullString) error {
	q.fail("AdminCreateLanguage")
	return nil
}

func (q *UnimplementedQuerier) AdminCreateLinkerCategory(ctx context.Context, arg db.AdminCreateLinkerCategoryParams) error {
	q.fail("AdminCreateLinkerCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminCreateLinkerItem(ctx context.Context, arg db.AdminCreateLinkerItemParams) error {
	q.fail("AdminCreateLinkerItem")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteExternalLink(ctx context.Context, id int32) error {
	q.fail("AdminDeleteExternalLink")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteExternalLinkByURL(ctx context.Context, url string) error {
	q.fail("AdminDeleteExternalLinkByURL")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteFAQ(ctx context.Context, id int32) error {
	q.fail("AdminDeleteFAQ")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteFAQCategory(ctx context.Context, id int32) error {
	q.fail("AdminDeleteFAQCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteForumCategory(ctx context.Context, idforumcategory int32) error {
	q.fail("AdminDeleteForumCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteForumThread(ctx context.Context, idforumthread int32) error {
	q.fail("AdminDeleteForumThread")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteForumTopic(ctx context.Context, idforumtopic int32) error {
	q.fail("AdminDeleteForumTopic")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteGrant(ctx context.Context, id int32) error {
	q.fail("AdminDeleteGrant")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteImageBoard(ctx context.Context, idimageboard int32) error {
	q.fail("AdminDeleteImageBoard")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteImagePost(ctx context.Context, idimagepost int32) error {
	q.fail("AdminDeleteImagePost")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteLanguage(ctx context.Context, id int32) error {
	q.fail("AdminDeleteLanguage")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteLinkerCategory(ctx context.Context, id int32) error {
	q.fail("AdminDeleteLinkerCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteLinkerQueuedItem(ctx context.Context, id int32) error {
	q.fail("AdminDeleteLinkerQueuedItem")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteNotification(ctx context.Context, id int32) error {
	q.fail("AdminDeleteNotification")
	return nil
}

func (q *UnimplementedQuerier) AdminDeletePendingEmail(ctx context.Context, id int32) error {
	q.fail("AdminDeletePendingEmail")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteTemplateOverride(ctx context.Context, name string) error {
	q.fail("AdminDeleteTemplateOverride")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteUserByID(ctx context.Context, idusers int32) error {
	q.fail("AdminDeleteUserByID")
	return nil
}

func (q *UnimplementedQuerier) AdminDeleteUserRole(ctx context.Context, iduserRoles int32) error {
	q.fail("AdminDeleteUserRole")
	return nil
}

func (q *UnimplementedQuerier) AdminDemoteAnnouncement(ctx context.Context, id int32) error {
	q.fail("AdminDemoteAnnouncement")
	return nil
}

func (q *UnimplementedQuerier) AdminForumCategoryThreadCounts(ctx context.Context) ([]*db.AdminForumCategoryThreadCountsRow, error) {
	q.fail("AdminForumCategoryThreadCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminForumHandlerThreadCounts(ctx context.Context) ([]*db.AdminForumHandlerThreadCountsRow, error) {
	q.fail("AdminForumHandlerThreadCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminForumTopicThreadCounts(ctx context.Context) ([]*db.AdminForumTopicThreadCountsRow, error) {
	q.fail("AdminForumTopicThreadCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetAllBlogEntriesByUser(ctx context.Context, authorID int32) ([]*db.AdminGetAllBlogEntriesByUserRow, error) {
	q.fail("AdminGetAllBlogEntriesByUser")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetAllCommentsByUser(ctx context.Context, userID int32) ([]*db.AdminGetAllCommentsByUserRow, error) {
	q.fail("AdminGetAllCommentsByUser")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetAllWritingsByAuthor(ctx context.Context, authorID int32) ([]*db.AdminGetAllWritingsByAuthorRow, error) {
	q.fail("AdminGetAllWritingsByAuthor")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetDashboardStats(ctx context.Context) (*db.AdminGetDashboardStatsRow, error) {
	q.fail("AdminGetDashboardStats")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQByID(ctx context.Context, id int32) (*db.Faq, error) {
	q.fail("AdminGetFAQByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQCategories(ctx context.Context) ([]*db.FaqCategory, error) {
	q.fail("AdminGetFAQCategories")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQCategoriesWithQuestionCount(ctx context.Context) ([]*db.AdminGetFAQCategoriesWithQuestionCountRow, error) {
	q.fail("AdminGetFAQCategoriesWithQuestionCount")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQCategoryWithQuestionCountByID(ctx context.Context, id int32) (*db.AdminGetFAQCategoryWithQuestionCountByIDRow, error) {
	q.fail("AdminGetFAQCategoryWithQuestionCountByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQDismissedQuestions(ctx context.Context) ([]*db.Faq, error) {
	q.fail("AdminGetFAQDismissedQuestions")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQQuestionsByCategory(ctx context.Context, categoryID sql.NullInt32) ([]*db.Faq, error) {
	q.fail("AdminGetFAQQuestionsByCategory")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetFAQUnansweredQuestions(ctx context.Context) ([]*db.Faq, error) {
	q.fail("AdminGetFAQUnansweredQuestions")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetForumStats(ctx context.Context) (*db.AdminGetForumStatsRow, error) {
	q.fail("AdminGetForumStats")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetImagePost(ctx context.Context, idimagepost int32) (*db.AdminGetImagePostRow, error) {
	q.fail("AdminGetImagePost")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetNotification(ctx context.Context, id int32) (*db.Notification, error) {
	q.fail("AdminGetNotification")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetPendingEmailByID(ctx context.Context, id int32) (*db.AdminGetPendingEmailByIDRow, error) {
	q.fail("AdminGetPendingEmailByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetRecentAuditLogs(ctx context.Context, limit int32) ([]*db.AdminGetRecentAuditLogsRow, error) {
	q.fail("AdminGetRecentAuditLogs")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetRequestByID(ctx context.Context, id int32) (*db.AdminRequestQueue, error) {
	q.fail("AdminGetRequestByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetRoleByID(ctx context.Context, id int32) (*db.Role, error) {
	q.fail("AdminGetRoleByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetRoleByNameForUser(ctx context.Context, arg db.AdminGetRoleByNameForUserParams) (int32, error) {
	q.fail("AdminGetRoleByNameForUser")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminGetSearchStats(ctx context.Context) (*db.AdminGetSearchStatsRow, error) {
	q.fail("AdminGetSearchStats")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetThreadsStartedByUser(ctx context.Context, usersIdusers int32) ([]*db.Forumthread, error) {
	q.fail("AdminGetThreadsStartedByUser")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetThreadsStartedByUserWithTopic(ctx context.Context, usersIdusers int32) ([]*db.AdminGetThreadsStartedByUserWithTopicRow, error) {
	q.fail("AdminGetThreadsStartedByUserWithTopic")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminGetWritingsByCategoryId(ctx context.Context, writingCategoryID int32) ([]*db.AdminGetWritingsByCategoryIdRow, error) {
	q.fail("AdminGetWritingsByCategoryId")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminImageboardPostCounts(ctx context.Context) ([]*db.AdminImageboardPostCountsRow, error) {
	q.fail("AdminImageboardPostCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminInsertBannedIp(ctx context.Context, arg db.AdminInsertBannedIpParams) error {
	q.fail("AdminInsertBannedIp")
	return nil
}

func (q *UnimplementedQuerier) AdminInsertLanguage(ctx context.Context, nameof sql.NullString) (sql.Result, error) {
	q.fail("AdminInsertLanguage")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminInsertQueuedLinkFromQueue(ctx context.Context, id int32) (int64, error) {
	q.fail("AdminInsertQueuedLinkFromQueue")
	return 0, nil
}

func (q *UnimplementedQuerier) AdminInsertRequestComment(ctx context.Context, arg db.AdminInsertRequestCommentParams) error {
	q.fail("AdminInsertRequestComment")
	return nil
}

func (q *UnimplementedQuerier) AdminInsertRequestQueue(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
	q.fail("AdminInsertRequestQueue")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminInsertWritingCategory(ctx context.Context, arg db.AdminInsertWritingCategoryParams) error {
	q.fail("AdminInsertWritingCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminIsBlogDeactivated(ctx context.Context, idblogs int32) (bool, error) {
	q.fail("AdminIsBlogDeactivated")
	return false, nil
}

func (q *UnimplementedQuerier) AdminIsCommentDeactivated(ctx context.Context, idcomments int32) (bool, error) {
	q.fail("AdminIsCommentDeactivated")
	return false, nil
}

func (q *UnimplementedQuerier) AdminIsImagepostDeactivated(ctx context.Context, idimagepost int32) (bool, error) {
	q.fail("AdminIsImagepostDeactivated")
	return false, nil
}

func (q *UnimplementedQuerier) AdminIsLinkDeactivated(ctx context.Context, id int32) (bool, error) {
	q.fail("AdminIsLinkDeactivated")
	return false, nil
}

func (q *UnimplementedQuerier) AdminIsUserDeactivated(ctx context.Context, idusers int32) (bool, error) {
	q.fail("AdminIsUserDeactivated")
	return false, nil
}

func (q *UnimplementedQuerier) AdminIsWritingDeactivated(ctx context.Context, idwriting int32) (bool, error) {
	q.fail("AdminIsWritingDeactivated")
	return false, nil
}

func (q *UnimplementedQuerier) AdminLanguageUsageCounts(ctx context.Context, arg db.AdminLanguageUsageCountsParams) (*db.AdminLanguageUsageCountsRow, error) {
	q.fail("AdminLanguageUsageCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAdministratorEmails(ctx context.Context) ([]string, error) {
	q.fail("AdminListAdministratorEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAllCommentsWithThreadInfo(ctx context.Context, arg db.AdminListAllCommentsWithThreadInfoParams) ([]*db.AdminListAllCommentsWithThreadInfoRow, error) {
	q.fail("AdminListAllCommentsWithThreadInfo")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAllPrivateForumThreads(ctx context.Context) ([]*db.AdminListAllPrivateForumThreadsRow, error) {
	q.fail("AdminListAllPrivateForumThreads")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAllPrivateTopics(ctx context.Context) ([]*db.AdminListAllPrivateTopicsRow, error) {
	q.fail("AdminListAllPrivateTopics")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAllUserIDs(ctx context.Context) ([]int32, error) {
	q.fail("AdminListAllUserIDs")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAllUsers(ctx context.Context) ([]*db.AdminListAllUsersRow, error) {
	q.fail("AdminListAllUsers")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAnnouncementsWithNews(ctx context.Context) ([]*db.AdminListAnnouncementsWithNewsRow, error) {
	q.fail("AdminListAnnouncementsWithNews")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListArchivedRequests(ctx context.Context) ([]*db.AdminRequestQueue, error) {
	q.fail("AdminListArchivedRequests")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListAuditLogs(ctx context.Context, arg db.AdminListAuditLogsParams) ([]*db.AdminListAuditLogsRow, error) {
	q.fail("AdminListAuditLogs")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListBoards(ctx context.Context, arg db.AdminListBoardsParams) ([]*db.Imageboard, error) {
	q.fail("AdminListBoards")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListDeactivatedBlogs(ctx context.Context, arg db.AdminListDeactivatedBlogsParams) ([]*db.AdminListDeactivatedBlogsRow, error) {
	q.fail("AdminListDeactivatedBlogs")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListDeactivatedComments(ctx context.Context, arg db.AdminListDeactivatedCommentsParams) ([]*db.AdminListDeactivatedCommentsRow, error) {
	q.fail("AdminListDeactivatedComments")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListDeactivatedImageposts(ctx context.Context, arg db.AdminListDeactivatedImagepostsParams) ([]*db.AdminListDeactivatedImagepostsRow, error) {
	q.fail("AdminListDeactivatedImageposts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListDeactivatedLinks(ctx context.Context, arg db.AdminListDeactivatedLinksParams) ([]*db.AdminListDeactivatedLinksRow, error) {
	q.fail("AdminListDeactivatedLinks")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListDeactivatedUsers(ctx context.Context, arg db.AdminListDeactivatedUsersParams) ([]*db.AdminListDeactivatedUsersRow, error) {
	q.fail("AdminListDeactivatedUsers")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListDeactivatedWritings(ctx context.Context, arg db.AdminListDeactivatedWritingsParams) ([]*db.AdminListDeactivatedWritingsRow, error) {
	q.fail("AdminListDeactivatedWritings")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListExternalLinks(ctx context.Context, arg db.AdminListExternalLinksParams) ([]*db.ExternalLink, error) {
	q.fail("AdminListExternalLinks")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListFailedEmails(ctx context.Context, arg db.AdminListFailedEmailsParams) ([]*db.AdminListFailedEmailsRow, error) {
	q.fail("AdminListFailedEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListForumCategoriesWithCounts(ctx context.Context, arg db.AdminListForumCategoriesWithCountsParams) ([]*db.AdminListForumCategoriesWithCountsRow, error) {
	q.fail("AdminListForumCategoriesWithCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListForumThreadGrantsByThreadID(ctx context.Context, itemID sql.NullInt32) ([]*db.AdminListForumThreadGrantsByThreadIDRow, error) {
	q.fail("AdminListForumThreadGrantsByThreadID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListForumThreads(ctx context.Context, arg db.AdminListForumThreadsParams) ([]*db.AdminListForumThreadsRow, error) {
	q.fail("AdminListForumThreads")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListForumTopicGrantsByTopicID(ctx context.Context, itemID sql.NullInt32) ([]*db.AdminListForumTopicGrantsByTopicIDRow, error) {
	q.fail("AdminListForumTopicGrantsByTopicID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListForumTopics(ctx context.Context, arg db.AdminListForumTopicsParams) ([]*db.Forumtopic, error) {
	q.fail("AdminListForumTopics")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListGrantsByRoleID(ctx context.Context, roleID sql.NullInt32) ([]*db.Grant, error) {
	q.fail("AdminListGrantsByRoleID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListGrantsByThreadID(ctx context.Context, itemID sql.NullInt32) ([]*db.AdminListGrantsByThreadIDRow, error) {
	q.fail("AdminListGrantsByThreadID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListGrantsByTopicID(ctx context.Context, itemID sql.NullInt32) ([]*db.AdminListGrantsByTopicIDRow, error) {
	q.fail("AdminListGrantsByTopicID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListLoginAttempts(ctx context.Context) ([]*db.LoginAttempt, error) {
	q.fail("AdminListLoginAttempts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescending(ctx context.Context, arg db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams) ([]*db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	q.fail("AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingDeactivatedBlogs(ctx context.Context, arg db.AdminListPendingDeactivatedBlogsParams) ([]*db.AdminListPendingDeactivatedBlogsRow, error) {
	q.fail("AdminListPendingDeactivatedBlogs")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingDeactivatedComments(ctx context.Context, arg db.AdminListPendingDeactivatedCommentsParams) ([]*db.AdminListPendingDeactivatedCommentsRow, error) {
	q.fail("AdminListPendingDeactivatedComments")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingDeactivatedImageposts(ctx context.Context, arg db.AdminListPendingDeactivatedImagepostsParams) ([]*db.AdminListPendingDeactivatedImagepostsRow, error) {
	q.fail("AdminListPendingDeactivatedImageposts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingDeactivatedLinks(ctx context.Context, arg db.AdminListPendingDeactivatedLinksParams) ([]*db.AdminListPendingDeactivatedLinksRow, error) {
	q.fail("AdminListPendingDeactivatedLinks")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingDeactivatedWritings(ctx context.Context, arg db.AdminListPendingDeactivatedWritingsParams) ([]*db.AdminListPendingDeactivatedWritingsRow, error) {
	q.fail("AdminListPendingDeactivatedWritings")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingRequests(ctx context.Context) ([]*db.AdminRequestQueue, error) {
	q.fail("AdminListPendingRequests")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListPendingUsers(ctx context.Context) ([]*db.AdminListPendingUsersRow, error) {
	q.fail("AdminListPendingUsers")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListRecentNotifications(ctx context.Context, limit int32) ([]*db.Notification, error) {
	q.fail("AdminListRecentNotifications")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListRequestComments(ctx context.Context, requestID int32) ([]*db.AdminRequestComment, error) {
	q.fail("AdminListRequestComments")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListRoles(ctx context.Context) ([]*db.Role, error) {
	q.fail("AdminListRoles")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListRolesWithUsers(ctx context.Context) ([]*db.AdminListRolesWithUsersRow, error) {
	q.fail("AdminListRolesWithUsers")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListSentEmails(ctx context.Context, arg db.AdminListSentEmailsParams) ([]*db.AdminListSentEmailsRow, error) {
	q.fail("AdminListSentEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListSessions(ctx context.Context) ([]*db.AdminListSessionsRow, error) {
	q.fail("AdminListSessions")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListTopicsWithUserGrantsNoRoles(ctx context.Context, includeAdmin interface{}) ([]*db.AdminListTopicsWithUserGrantsNoRolesRow, error) {
	q.fail("AdminListTopicsWithUserGrantsNoRoles")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListUnsentPendingEmails(ctx context.Context, arg db.AdminListUnsentPendingEmailsParams) ([]*db.AdminListUnsentPendingEmailsRow, error) {
	q.fail("AdminListUnsentPendingEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListUploadedImages(ctx context.Context, arg db.AdminListUploadedImagesParams) ([]*db.UploadedImage, error) {
	q.fail("AdminListUploadedImages")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListUserEmails(ctx context.Context, userID int32) ([]*db.UserEmail, error) {
	q.fail("AdminListUserEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListUserIDsByRole(ctx context.Context, name string) ([]int32, error) {
	q.fail("AdminListUserIDsByRole")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListUsersByID(ctx context.Context, ids []int32) ([]*db.AdminListUsersByIDRow, error) {
	q.fail("AdminListUsersByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminListUsersByRoleID(ctx context.Context, roleID int32) ([]*db.AdminListUsersByRoleIDRow, error) {
	q.fail("AdminListUsersByRoleID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminMarkBlogRestored(ctx context.Context, idblogs int32) error {
	q.fail("AdminMarkBlogRestored")
	return nil
}

func (q *UnimplementedQuerier) AdminMarkCommentRestored(ctx context.Context, idcomments int32) error {
	q.fail("AdminMarkCommentRestored")
	return nil
}

func (q *UnimplementedQuerier) AdminMarkImagepostRestored(ctx context.Context, idimagepost int32) error {
	q.fail("AdminMarkImagepostRestored")
	return nil
}

func (q *UnimplementedQuerier) AdminMarkLinkRestored(ctx context.Context, id int32) error {
	q.fail("AdminMarkLinkRestored")
	return nil
}

func (q *UnimplementedQuerier) AdminMarkNotificationRead(ctx context.Context, id int32) error {
	q.fail("AdminMarkNotificationRead")
	return nil
}

func (q *UnimplementedQuerier) AdminMarkNotificationUnread(ctx context.Context, id int32) error {
	q.fail("AdminMarkNotificationUnread")
	return nil
}

func (q *UnimplementedQuerier) AdminMarkWritingRestored(ctx context.Context, idwriting int32) error {
	q.fail("AdminMarkWritingRestored")
	return nil
}

func (q *UnimplementedQuerier) AdminPromoteAnnouncement(ctx context.Context, siteNewsID int32) error {
	q.fail("AdminPromoteAnnouncement")
	return nil
}

func (q *UnimplementedQuerier) AdminPurgeReadNotifications(ctx context.Context) error {
	q.fail("AdminPurgeReadNotifications")
	return nil
}

func (q *UnimplementedQuerier) AdminRebuildAllForumTopicMetaColumns(ctx context.Context) error {
	q.fail("AdminRebuildAllForumTopicMetaColumns")
	return nil
}

func (q *UnimplementedQuerier) AdminRecalculateAllForumThreadMetaData(ctx context.Context) error {
	q.fail("AdminRecalculateAllForumThreadMetaData")
	return nil
}

func (q *UnimplementedQuerier) AdminRecalculateForumThreadByIdMetaData(ctx context.Context, idforumthread int32) error {
	q.fail("AdminRecalculateForumThreadByIdMetaData")
	return nil
}

func (q *UnimplementedQuerier) AdminRenameFAQCategory(ctx context.Context, arg db.AdminRenameFAQCategoryParams) error {
	q.fail("AdminRenameFAQCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminRenameLanguage(ctx context.Context, arg db.AdminRenameLanguageParams) error {
	q.fail("AdminRenameLanguage")
	return nil
}

func (q *UnimplementedQuerier) AdminRenameLinkerCategory(ctx context.Context, arg db.AdminRenameLinkerCategoryParams) error {
	q.fail("AdminRenameLinkerCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminReplaceSiteNewsURL(ctx context.Context, arg db.AdminReplaceSiteNewsURLParams) error {
	q.fail("AdminReplaceSiteNewsURL")
	return nil
}

func (q *UnimplementedQuerier) AdminRestoreBlog(ctx context.Context, arg db.AdminRestoreBlogParams) error {
	q.fail("AdminRestoreBlog")
	return nil
}

func (q *UnimplementedQuerier) AdminRestoreComment(ctx context.Context, arg db.AdminRestoreCommentParams) error {
	q.fail("AdminRestoreComment")
	return nil
}

func (q *UnimplementedQuerier) AdminRestoreImagepost(ctx context.Context, arg db.AdminRestoreImagepostParams) error {
	q.fail("AdminRestoreImagepost")
	return nil
}

func (q *UnimplementedQuerier) AdminRestoreLink(ctx context.Context, arg db.AdminRestoreLinkParams) error {
	q.fail("AdminRestoreLink")
	return nil
}

func (q *UnimplementedQuerier) AdminRestoreUser(ctx context.Context, idusers int32) error {
	q.fail("AdminRestoreUser")
	return nil
}

func (q *UnimplementedQuerier) AdminRestoreWriting(ctx context.Context, arg db.AdminRestoreWritingParams) error {
	q.fail("AdminRestoreWriting")
	return nil
}

func (q *UnimplementedQuerier) AdminScrubBlog(ctx context.Context, arg db.AdminScrubBlogParams) error {
	q.fail("AdminScrubBlog")
	return nil
}

func (q *UnimplementedQuerier) AdminScrubComment(ctx context.Context, arg db.AdminScrubCommentParams) error {
	q.fail("AdminScrubComment")
	return nil
}

func (q *UnimplementedQuerier) AdminScrubImagepost(ctx context.Context, idimagepost int32) error {
	q.fail("AdminScrubImagepost")
	return nil
}

func (q *UnimplementedQuerier) AdminScrubLink(ctx context.Context, arg db.AdminScrubLinkParams) error {
	q.fail("AdminScrubLink")
	return nil
}

func (q *UnimplementedQuerier) AdminScrubUser(ctx context.Context, arg db.AdminScrubUserParams) error {
	q.fail("AdminScrubUser")
	return nil
}

func (q *UnimplementedQuerier) AdminScrubWriting(ctx context.Context, arg db.AdminScrubWritingParams) error {
	q.fail("AdminScrubWriting")
	return nil
}

func (q *UnimplementedQuerier) AdminSetAnnouncementActive(ctx context.Context, arg db.AdminSetAnnouncementActiveParams) error {
	q.fail("AdminSetAnnouncementActive")
	return nil
}

func (q *UnimplementedQuerier) AdminSetTemplateOverride(ctx context.Context, arg db.AdminSetTemplateOverrideParams) error {
	q.fail("AdminSetTemplateOverride")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateBannedIp(ctx context.Context, arg db.AdminUpdateBannedIpParams) error {
	q.fail("AdminUpdateBannedIp")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateFAQQuestionAnswer(ctx context.Context, arg db.AdminUpdateFAQQuestionAnswerParams) error {
	q.fail("AdminUpdateFAQQuestionAnswer")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateForumCategory(ctx context.Context, arg db.AdminUpdateForumCategoryParams) error {
	q.fail("AdminUpdateForumCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateForumTopic(ctx context.Context, arg db.AdminUpdateForumTopicParams) error {
	q.fail("AdminUpdateForumTopic")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateGrantActive(ctx context.Context, arg db.AdminUpdateGrantActiveParams) error {
	q.fail("AdminUpdateGrantActive")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateImageBoard(ctx context.Context, arg db.AdminUpdateImageBoardParams) error {
	q.fail("AdminUpdateImageBoard")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateImagePost(ctx context.Context, arg db.AdminUpdateImagePostParams) error {
	q.fail("AdminUpdateImagePost")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateLinkerCategorySortOrder(ctx context.Context, arg db.AdminUpdateLinkerCategorySortOrderParams) error {
	q.fail("AdminUpdateLinkerCategorySortOrder")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateLinkerItem(ctx context.Context, arg db.AdminUpdateLinkerItemParams) error {
	q.fail("AdminUpdateLinkerItem")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateLinkerQueuedItem(ctx context.Context, arg db.AdminUpdateLinkerQueuedItemParams) error {
	q.fail("AdminUpdateLinkerQueuedItem")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateRequestStatus(ctx context.Context, arg db.AdminUpdateRequestStatusParams) error {
	q.fail("AdminUpdateRequestStatus")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateRole(ctx context.Context, arg db.AdminUpdateRoleParams) error {
	q.fail("AdminUpdateRole")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateRolePublicProfileAllowed(ctx context.Context, arg db.AdminUpdateRolePublicProfileAllowedParams) error {
	q.fail("AdminUpdateRolePublicProfileAllowed")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateUserEmail(ctx context.Context, arg db.AdminUpdateUserEmailParams) error {
	q.fail("AdminUpdateUserEmail")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateUserRole(ctx context.Context, arg db.AdminUpdateUserRoleParams) error {
	q.fail("AdminUpdateUserRole")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateUsernameByID(ctx context.Context, arg db.AdminUpdateUsernameByIDParams) error {
	q.fail("AdminUpdateUsernameByID")
	return nil
}

func (q *UnimplementedQuerier) AdminUpdateWritingCategory(ctx context.Context, arg db.AdminUpdateWritingCategoryParams) error {
	q.fail("AdminUpdateWritingCategory")
	return nil
}

func (q *UnimplementedQuerier) AdminUserPostCounts(ctx context.Context) ([]*db.AdminUserPostCountsRow, error) {
	q.fail("AdminUserPostCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminUserPostCountsByID(ctx context.Context, idusers int32) (*db.AdminUserPostCountsByIDRow, error) {
	q.fail("AdminUserPostCountsByID")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminWordListWithCounts(ctx context.Context, arg db.AdminWordListWithCountsParams) ([]*db.AdminWordListWithCountsRow, error) {
	q.fail("AdminWordListWithCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminWordListWithCountsByPrefix(ctx context.Context, arg db.AdminWordListWithCountsByPrefixParams) ([]*db.AdminWordListWithCountsByPrefixRow, error) {
	q.fail("AdminWordListWithCountsByPrefix")
	return nil, nil
}

func (q *UnimplementedQuerier) AdminWritingCategoryCounts(ctx context.Context) ([]*db.AdminWritingCategoryCountsRow, error) {
	q.fail("AdminWritingCategoryCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) CheckUserHasGrant(ctx context.Context, arg db.CheckUserHasGrantParams) (bool, error) {
	q.fail("CheckUserHasGrant")
	return false, nil
}

func (q *UnimplementedQuerier) ClearUnreadContentPrivateLabelExceptUser(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error {
	q.fail("ClearUnreadContentPrivateLabelExceptUser")
	return nil
}

func (q *UnimplementedQuerier) CreateBlogEntryForWriter(ctx context.Context, arg db.CreateBlogEntryForWriterParams) (int64, error) {
	q.fail("CreateBlogEntryForWriter")
	return 0, nil
}

func (q *UnimplementedQuerier) CreateBookmarksForLister(ctx context.Context, arg db.CreateBookmarksForListerParams) error {
	q.fail("CreateBookmarksForLister")
	return nil
}

func (q *UnimplementedQuerier) CreateCommentInSectionForCommenter(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
	q.fail("CreateCommentInSectionForCommenter")
	return 0, nil
}

func (q *UnimplementedQuerier) CreateFAQQuestionForWriter(ctx context.Context, arg db.CreateFAQQuestionForWriterParams) error {
	q.fail("CreateFAQQuestionForWriter")
	return nil
}

func (q *UnimplementedQuerier) CreateForumTopicForPoster(ctx context.Context, arg db.CreateForumTopicForPosterParams) (int64, error) {
	q.fail("CreateForumTopicForPoster")
	return 0, nil
}

func (q *UnimplementedQuerier) CreateGrant(ctx context.Context, arg db.CreateGrantParams) error {
	q.fail("CreateGrant")
	return nil
}

func (q *UnimplementedQuerier) CreateImagePostForPoster(ctx context.Context, arg db.CreateImagePostForPosterParams) (int64, error) {
	q.fail("CreateImagePostForPoster")
	return 0, nil
}

func (q *UnimplementedQuerier) CreateLinkerQueuedItemForWriter(ctx context.Context, arg db.CreateLinkerQueuedItemForWriterParams) error {
	q.fail("CreateLinkerQueuedItemForWriter")
	return nil
}

func (q *UnimplementedQuerier) CreateNewsPostForWriter(ctx context.Context, arg db.CreateNewsPostForWriterParams) (int64, error) {
	q.fail("CreateNewsPostForWriter")
	return 0, nil
}

func (q *UnimplementedQuerier) CreatePasswordResetForUser(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
	q.fail("CreatePasswordResetForUser")
	return nil
}

func (q *UnimplementedQuerier) CreateUploadedImageForUploader(ctx context.Context, arg db.CreateUploadedImageForUploaderParams) (int64, error) {
	q.fail("CreateUploadedImageForUploader")
	return 0, nil
}

func (q *UnimplementedQuerier) CreateWritingForWriter(ctx context.Context, arg db.CreateWritingForWriterParams) (int64, error) {
	q.fail("CreateWritingForWriter")
	return 0, nil
}

func (q *UnimplementedQuerier) DeactivateNewsPost(ctx context.Context, idsitenews int32) error {
	q.fail("DeactivateNewsPost")
	return nil
}

func (q *UnimplementedQuerier) DeleteGrantByProperties(ctx context.Context, arg db.DeleteGrantByPropertiesParams) error {
	q.fail("DeleteGrantByProperties")
	return nil
}

func (q *UnimplementedQuerier) DeleteGrantsByRoleID(ctx context.Context, roleID sql.NullInt32) error {
	q.fail("DeleteGrantsByRoleID")
	return nil
}

func (q *UnimplementedQuerier) DeleteNotificationForLister(ctx context.Context, arg db.DeleteNotificationForListerParams) error {
	q.fail("DeleteNotificationForLister")
	return nil
}

func (q *UnimplementedQuerier) DeleteSubscriptionByIDForSubscriber(ctx context.Context, arg db.DeleteSubscriptionByIDForSubscriberParams) error {
	q.fail("DeleteSubscriptionByIDForSubscriber")
	return nil
}

func (q *UnimplementedQuerier) DeleteSubscriptionForSubscriber(ctx context.Context, arg db.DeleteSubscriptionForSubscriberParams) error {
	q.fail("DeleteSubscriptionForSubscriber")
	return nil
}

func (q *UnimplementedQuerier) DeleteUserEmailForOwner(ctx context.Context, arg db.DeleteUserEmailForOwnerParams) error {
	q.fail("DeleteUserEmailForOwner")
	return nil
}

func (q *UnimplementedQuerier) DeleteUserLanguagesForUser(ctx context.Context, userID int32) error {
	q.fail("DeleteUserLanguagesForUser")
	return nil
}

func (q *UnimplementedQuerier) GetActiveAnnouncementWithNewsForLister(ctx context.Context, arg db.GetActiveAnnouncementWithNewsForListerParams) (*db.GetActiveAnnouncementWithNewsForListerRow, error) {
	q.fail("GetActiveAnnouncementWithNewsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAdministratorUserRole(ctx context.Context, usersIdusers int32) (*db.UserRole, error) {
	q.fail("GetAdministratorUserRole")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllAnsweredFAQWithFAQCategoriesForUser(ctx context.Context, arg db.GetAllAnsweredFAQWithFAQCategoriesForUserParams) ([]*db.GetAllAnsweredFAQWithFAQCategoriesForUserRow, error) {
	q.fail("GetAllAnsweredFAQWithFAQCategoriesForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllCommentsForIndex(ctx context.Context) ([]*db.GetAllCommentsForIndexRow, error) {
	q.fail("GetAllCommentsForIndex")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllForumCategories(ctx context.Context, arg db.GetAllForumCategoriesParams) ([]*db.Forumcategory, error) {
	q.fail("GetAllForumCategories")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllForumCategoriesWithSubcategoryCount(ctx context.Context, arg db.GetAllForumCategoriesWithSubcategoryCountParams) ([]*db.GetAllForumCategoriesWithSubcategoryCountRow, error) {
	q.fail("GetAllForumCategoriesWithSubcategoryCount")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllForumThreadsWithTopic(ctx context.Context) ([]*db.GetAllForumThreadsWithTopicRow, error) {
	q.fail("GetAllForumThreadsWithTopic")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllForumTopics(ctx context.Context, arg db.GetAllForumTopicsParams) ([]*db.Forumtopic, error) {
	q.fail("GetAllForumTopics")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllForumTopicsByCategoryIdForUserWithLastPosterName(ctx context.Context, arg db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams) ([]*db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow, error) {
	q.fail("GetAllForumTopicsByCategoryIdForUserWithLastPosterName")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllImagePostsForIndex(ctx context.Context) ([]*db.GetAllImagePostsForIndexRow, error) {
	q.fail("GetAllImagePostsForIndex")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerCategories(ctx context.Context) ([]*db.LinkerCategory, error) {
	q.fail("GetAllLinkerCategories")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerCategoriesForUser(ctx context.Context, arg db.GetAllLinkerCategoriesForUserParams) ([]*db.LinkerCategory, error) {
	q.fail("GetAllLinkerCategoriesForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(ctx context.Context, arg db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow, error) {
	q.fail("GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUser(ctx context.Context, arg db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserParams) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	q.fail("GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated(ctx context.Context, arg db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedParams) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow, error) {
	q.fail("GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow(ctx context.Context, arg db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRowParams) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRowRow, error) {
	q.fail("GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginated(ctx context.Context, arg db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginatedParams) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginatedRow, error) {
	q.fail("GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginated")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails(ctx context.Context) ([]*db.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow, error) {
	q.fail("GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllLinkersForIndex(ctx context.Context) ([]*db.GetAllLinkersForIndexRow, error) {
	q.fail("GetAllLinkersForIndex")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllSiteNewsForIndex(ctx context.Context) ([]*db.GetAllSiteNewsForIndexRow, error) {
	q.fail("GetAllSiteNewsForIndex")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllWritingsByAuthorForLister(ctx context.Context, arg db.GetAllWritingsByAuthorForListerParams) ([]*db.GetAllWritingsByAuthorForListerRow, error) {
	q.fail("GetAllWritingsByAuthorForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) GetAllWritingsForIndex(ctx context.Context) ([]*db.GetAllWritingsForIndexRow, error) {
	q.fail("GetAllWritingsForIndex")
	return nil, nil
}

func (q *UnimplementedQuerier) GetBlogEntryForListerByID(ctx context.Context, arg db.GetBlogEntryForListerByIDParams) (*db.GetBlogEntryForListerByIDRow, error) {
	q.fail("GetBlogEntryForListerByID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetBookmarksForUser(ctx context.Context, usersIdusers int32) (*db.GetBookmarksForUserRow, error) {
	q.fail("GetBookmarksForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetCommentById(ctx context.Context, idcomments int32) (*db.Comment, error) {
	q.fail("GetCommentById")
	return nil, nil
}

func (q *UnimplementedQuerier) GetCommentByIdForUser(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
	q.fail("GetCommentByIdForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetCommentsByIdsForUserWithThreadInfo(ctx context.Context, arg db.GetCommentsByIdsForUserWithThreadInfoParams) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, error) {
	q.fail("GetCommentsByIdsForUserWithThreadInfo")
	return nil, nil
}

func (q *UnimplementedQuerier) GetCommentsBySectionThreadIdForUser(ctx context.Context, arg db.GetCommentsBySectionThreadIdForUserParams) ([]*db.GetCommentsBySectionThreadIdForUserRow, error) {
	q.fail("GetCommentsBySectionThreadIdForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetCommentsByThreadIdForUser(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	q.fail("GetCommentsByThreadIdForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetContentReadMarker(ctx context.Context, arg db.GetContentReadMarkerParams) (*db.GetContentReadMarkerRow, error) {
	q.fail("GetContentReadMarker")
	return nil, nil
}

func (q *UnimplementedQuerier) GetExternalLink(ctx context.Context, url string) (*db.ExternalLink, error) {
	q.fail("GetExternalLink")
	return nil, nil
}

func (q *UnimplementedQuerier) GetExternalLinkByID(ctx context.Context, id int32) (*db.ExternalLink, error) {
	q.fail("GetExternalLinkByID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetFAQAnsweredQuestions(ctx context.Context, arg db.GetFAQAnsweredQuestionsParams) ([]*db.Faq, error) {
	q.fail("GetFAQAnsweredQuestions")
	return nil, nil
}

func (q *UnimplementedQuerier) GetFAQByID(ctx context.Context, arg db.GetFAQByIDParams) (*db.Faq, error) {
	q.fail("GetFAQByID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetFAQQuestionsByCategory(ctx context.Context, arg db.GetFAQQuestionsByCategoryParams) ([]*db.Faq, error) {
	q.fail("GetFAQQuestionsByCategory")
	return nil, nil
}

func (q *UnimplementedQuerier) GetFAQRevisionsForAdmin(ctx context.Context, faqID int32) ([]*db.FaqRevision, error) {
	q.fail("GetFAQRevisionsForAdmin")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumCategoryById(ctx context.Context, arg db.GetForumCategoryByIdParams) (*db.Forumcategory, error) {
	q.fail("GetForumCategoryById")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumThreadIdByNewsPostId(ctx context.Context, idsitenews int32) (*db.GetForumThreadIdByNewsPostIdRow, error) {
	q.fail("GetForumThreadIdByNewsPostId")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(ctx context.Context, arg db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	q.fail("GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumTopicById(ctx context.Context, idforumtopic int32) (*db.Forumtopic, error) {
	q.fail("GetForumTopicById")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumTopicByIdForUser(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
	q.fail("GetForumTopicByIdForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumTopicIdByThreadId(ctx context.Context, idforumthread int32) (int32, error) {
	q.fail("GetForumTopicIdByThreadId")
	return 0, nil
}

func (q *UnimplementedQuerier) GetForumTopicsByCategoryId(ctx context.Context, arg db.GetForumTopicsByCategoryIdParams) ([]*db.Forumtopic, error) {
	q.fail("GetForumTopicsByCategoryId")
	return nil, nil
}

func (q *UnimplementedQuerier) GetForumTopicsForUser(ctx context.Context, arg db.GetForumTopicsForUserParams) ([]*db.GetForumTopicsForUserRow, error) {
	q.fail("GetForumTopicsForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetGrantsByRoleID(ctx context.Context, roleID sql.NullInt32) ([]*db.Grant, error) {
	q.fail("GetGrantsByRoleID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetImageBoardById(ctx context.Context, idimageboard int32) (*db.Imageboard, error) {
	q.fail("GetImageBoardById")
	return nil, nil
}

func (q *UnimplementedQuerier) GetImagePostByIDForLister(ctx context.Context, arg db.GetImagePostByIDForListerParams) (*db.GetImagePostByIDForListerRow, error) {
	q.fail("GetImagePostByIDForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) GetImagePostInfoByPath(ctx context.Context, arg db.GetImagePostInfoByPathParams) (*db.GetImagePostInfoByPathRow, error) {
	q.fail("GetImagePostInfoByPath")
	return nil, nil
}

func (q *UnimplementedQuerier) GetImagePostsByUserDescending(ctx context.Context, arg db.GetImagePostsByUserDescendingParams) ([]*db.GetImagePostsByUserDescendingRow, error) {
	q.fail("GetImagePostsByUserDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) GetImagePostsByUserDescendingAll(ctx context.Context, arg db.GetImagePostsByUserDescendingAllParams) ([]*db.GetImagePostsByUserDescendingAllRow, error) {
	q.fail("GetImagePostsByUserDescendingAll")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLatestAnnouncementByNewsID(ctx context.Context, siteNewsID int32) (*db.SiteAnnouncement, error) {
	q.fail("GetLatestAnnouncementByNewsID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerCategoriesWithCount(ctx context.Context) ([]*db.GetLinkerCategoriesWithCountRow, error) {
	q.fail("GetLinkerCategoriesWithCount")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerCategoryById(ctx context.Context, id int32) (*db.LinkerCategory, error) {
	q.fail("GetLinkerCategoryById")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerCategoryLinkCounts(ctx context.Context) ([]*db.GetLinkerCategoryLinkCountsRow, error) {
	q.fail("GetLinkerCategoryLinkCounts")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(ctx context.Context, id int32) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error) {
	q.fail("GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(ctx context.Context, arg db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	q.fail("GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending(ctx context.Context, linkerids []int32) ([]*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow, error) {
	q.fail("GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser(ctx context.Context, arg db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUserParams) ([]*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	q.fail("GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerItemsByUserDescending(ctx context.Context, arg db.GetLinkerItemsByUserDescendingParams) ([]*db.GetLinkerItemsByUserDescendingRow, error) {
	q.fail("GetLinkerItemsByUserDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLinkerItemsByUserDescendingForUser(ctx context.Context, arg db.GetLinkerItemsByUserDescendingForUserParams) ([]*db.GetLinkerItemsByUserDescendingForUserRow, error) {
	q.fail("GetLinkerItemsByUserDescendingForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetLoginRoleForUser(ctx context.Context, usersIdusers int32) (int32, error) {
	q.fail("GetLoginRoleForUser")
	return 0, nil
}

func (q *UnimplementedQuerier) GetMaxNotificationPriority(ctx context.Context, userID int32) (interface{}, error) {
	q.fail("GetMaxNotificationPriority")
	return nil, nil
}

func (q *UnimplementedQuerier) GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx context.Context, arg db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
	q.fail("GetNewsPostByIdWithWriterIdAndThreadCommentCount")
	return nil, nil
}

func (q *UnimplementedQuerier) GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount(ctx context.Context, arg db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams) ([]*db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow, error) {
	q.fail("GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount")
	return nil, nil
}

func (q *UnimplementedQuerier) GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(ctx context.Context, arg db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams) ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	q.fail("GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending")
	return nil, nil
}

func (q *UnimplementedQuerier) GetNotificationEmailByUserID(ctx context.Context, userID int32) (*db.UserEmail, error) {
	q.fail("GetNotificationEmailByUserID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetNotificationForLister(ctx context.Context, arg db.GetNotificationForListerParams) (*db.Notification, error) {
	q.fail("GetNotificationForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) GetPasswordResetByCode(ctx context.Context, arg db.GetPasswordResetByCodeParams) (*db.PendingPassword, error) {
	q.fail("GetPasswordResetByCode")
	return nil, nil
}

func (q *UnimplementedQuerier) GetPasswordResetByUser(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
	q.fail("GetPasswordResetByUser")
	return nil, nil
}

func (q *UnimplementedQuerier) GetPendingEmailErrorCount(ctx context.Context, id int32) (int32, error) {
	q.fail("GetPendingEmailErrorCount")
	return 0, nil
}

func (q *UnimplementedQuerier) GetPermissionsByUserID(ctx context.Context, usersIdusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
	q.fail("GetPermissionsByUserID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetPermissionsWithUsers(ctx context.Context, arg db.GetPermissionsWithUsersParams) ([]*db.GetPermissionsWithUsersRow, error) {
	q.fail("GetPermissionsWithUsers")
	return nil, nil
}

func (q *UnimplementedQuerier) GetPreferenceForLister(ctx context.Context, listerID int32) (*db.Preference, error) {
	q.fail("GetPreferenceForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) GetPublicProfileRoleForUser(ctx context.Context, usersIdusers int32) (int32, error) {
	q.fail("GetPublicProfileRoleForUser")
	return 0, nil
}

func (q *UnimplementedQuerier) GetPublicWritings(ctx context.Context, arg db.GetPublicWritingsParams) ([]*db.Writing, error) {
	q.fail("GetPublicWritings")
	return nil, nil
}

func (q *UnimplementedQuerier) GetRoleByName(ctx context.Context, name string) (*db.Role, error) {
	q.fail("GetRoleByName")
	return nil, nil
}

func (q *UnimplementedQuerier) GetThreadBySectionThreadIDForReplier(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
	q.fail("GetThreadBySectionThreadIDForReplier")
	return nil, nil
}

func (q *UnimplementedQuerier) GetThreadLastPosterAndPerms(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
	q.fail("GetThreadLastPosterAndPerms")
	return nil, nil
}

func (q *UnimplementedQuerier) GetUnreadNotificationCountForLister(ctx context.Context, listerID int32) (int64, error) {
	q.fail("GetUnreadNotificationCountForLister")
	return 0, nil
}

func (q *UnimplementedQuerier) GetUserEmailByCode(ctx context.Context, lastVerificationCode sql.NullString) (*db.UserEmail, error) {
	q.fail("GetUserEmailByCode")
	return nil, nil
}

func (q *UnimplementedQuerier) GetUserEmailByEmail(ctx context.Context, email string) (*db.UserEmail, error) {
	q.fail("GetUserEmailByEmail")
	return nil, nil
}

func (q *UnimplementedQuerier) GetUserEmailByID(ctx context.Context, id int32) (*db.UserEmail, error) {
	q.fail("GetUserEmailByID")
	return nil, nil
}

func (q *UnimplementedQuerier) GetUserLanguages(ctx context.Context, usersIdusers int32) ([]*db.UserLanguage, error) {
	q.fail("GetUserLanguages")
	return nil, nil
}

func (q *UnimplementedQuerier) GetUserRole(ctx context.Context, usersIdusers int32) (string, error) {
	q.fail("GetUserRole")
	return "", nil
}

func (q *UnimplementedQuerier) GetUserRoles(ctx context.Context) ([]*db.GetUserRolesRow, error) {
	q.fail("GetUserRoles")
	return nil, nil
}

func (q *UnimplementedQuerier) GetVerifiedUserEmails(ctx context.Context) ([]*db.GetVerifiedUserEmailsRow, error) {
	q.fail("GetVerifiedUserEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) GetWritingCategoryById(ctx context.Context, idwritingcategory int32) (*db.WritingCategory, error) {
	q.fail("GetWritingCategoryById")
	return nil, nil
}

func (q *UnimplementedQuerier) GetWritingForListerByID(ctx context.Context, arg db.GetWritingForListerByIDParams) (*db.GetWritingForListerByIDRow, error) {
	q.fail("GetWritingForListerByID")
	return nil, nil
}

func (q *UnimplementedQuerier) InsertAdminUserComment(ctx context.Context, arg db.InsertAdminUserCommentParams) error {
	q.fail("InsertAdminUserComment")
	return nil
}

func (q *UnimplementedQuerier) InsertAuditLog(ctx context.Context, arg db.InsertAuditLogParams) error {
	q.fail("InsertAuditLog")
	return nil
}

func (q *UnimplementedQuerier) InsertEmailPreferenceForLister(ctx context.Context, arg db.InsertEmailPreferenceForListerParams) error {
	q.fail("InsertEmailPreferenceForLister")
	return nil
}

func (q *UnimplementedQuerier) InsertFAQQuestionForWriter(ctx context.Context, arg db.InsertFAQQuestionForWriterParams) (sql.Result, error) {
	q.fail("InsertFAQQuestionForWriter")
	return nil, nil
}

func (q *UnimplementedQuerier) InsertFAQRevisionForUser(ctx context.Context, arg db.InsertFAQRevisionForUserParams) error {
	q.fail("InsertFAQRevisionForUser")
	return nil
}

func (q *UnimplementedQuerier) InsertPassword(ctx context.Context, arg db.InsertPasswordParams) error {
	q.fail("InsertPassword")
	return nil
}

func (q *UnimplementedQuerier) InsertPendingEmail(ctx context.Context, arg db.InsertPendingEmailParams) error {
	q.fail("InsertPendingEmail")
	return nil
}

func (q *UnimplementedQuerier) InsertPreferenceForLister(ctx context.Context, arg db.InsertPreferenceForListerParams) error {
	q.fail("InsertPreferenceForLister")
	return nil
}

func (q *UnimplementedQuerier) InsertSubscription(ctx context.Context, arg db.InsertSubscriptionParams) error {
	q.fail("InsertSubscription")
	return nil
}

func (q *UnimplementedQuerier) InsertUserEmail(ctx context.Context, arg db.InsertUserEmailParams) error {
	q.fail("InsertUserEmail")
	return nil
}

func (q *UnimplementedQuerier) InsertUserLang(ctx context.Context, arg db.InsertUserLangParams) error {
	q.fail("InsertUserLang")
	return nil
}

func (q *UnimplementedQuerier) InsertWriting(ctx context.Context, arg db.InsertWritingParams) (int64, error) {
	q.fail("InsertWriting")
	return 0, nil
}

func (q *UnimplementedQuerier) LatestAdminUserComment(ctx context.Context, usersIdusers int32) (*db.AdminUserComment, error) {
	q.fail("LatestAdminUserComment")
	return nil, nil
}

func (q *UnimplementedQuerier) LinkerSearchFirst(ctx context.Context, arg db.LinkerSearchFirstParams) ([]int32, error) {
	q.fail("LinkerSearchFirst")
	return nil, nil
}

func (q *UnimplementedQuerier) LinkerSearchNext(ctx context.Context, arg db.LinkerSearchNextParams) ([]int32, error) {
	q.fail("LinkerSearchNext")
	return nil, nil
}

func (q *UnimplementedQuerier) ListActiveBans(ctx context.Context) ([]*db.BannedIp, error) {
	q.fail("ListActiveBans")
	return nil, nil
}

func (q *UnimplementedQuerier) ListAdminUserComments(ctx context.Context, usersIdusers int32) ([]*db.AdminUserComment, error) {
	q.fail("ListAdminUserComments")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBannedIps(ctx context.Context) ([]*db.BannedIp, error) {
	q.fail("ListBannedIps")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBlogEntriesByAuthorForLister(ctx context.Context, arg db.ListBlogEntriesByAuthorForListerParams) ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
	q.fail("ListBlogEntriesByAuthorForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBlogEntriesByIDsForLister(ctx context.Context, arg db.ListBlogEntriesByIDsForListerParams) ([]*db.ListBlogEntriesByIDsForListerRow, error) {
	q.fail("ListBlogEntriesByIDsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBlogEntriesForLister(ctx context.Context, arg db.ListBlogEntriesForListerParams) ([]*db.ListBlogEntriesForListerRow, error) {
	q.fail("ListBlogEntriesForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBlogIDsBySearchWordFirstForLister(ctx context.Context, arg db.ListBlogIDsBySearchWordFirstForListerParams) ([]int32, error) {
	q.fail("ListBlogIDsBySearchWordFirstForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBlogIDsBySearchWordNextForLister(ctx context.Context, arg db.ListBlogIDsBySearchWordNextForListerParams) ([]int32, error) {
	q.fail("ListBlogIDsBySearchWordNextForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBloggersForLister(ctx context.Context, arg db.ListBloggersForListerParams) ([]*db.ListBloggersForListerRow, error) {
	q.fail("ListBloggersForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBloggersSearchForLister(ctx context.Context, arg db.ListBloggersSearchForListerParams) ([]*db.ListBloggersSearchForListerRow, error) {
	q.fail("ListBloggersSearchForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBoardsByParentIDForLister(ctx context.Context, arg db.ListBoardsByParentIDForListerParams) ([]*db.Imageboard, error) {
	q.fail("ListBoardsByParentIDForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListBoardsForLister(ctx context.Context, arg db.ListBoardsForListerParams) ([]*db.Imageboard, error) {
	q.fail("ListBoardsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListCommentIDsBySearchWordFirstForListerInRestrictedTopic(ctx context.Context, arg db.ListCommentIDsBySearchWordFirstForListerInRestrictedTopicParams) ([]int32, error) {
	q.fail("ListCommentIDsBySearchWordFirstForListerInRestrictedTopic")
	return nil, nil
}

func (q *UnimplementedQuerier) ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic(ctx context.Context, arg db.ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopicParams) ([]int32, error) {
	q.fail("ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic")
	return nil, nil
}

func (q *UnimplementedQuerier) ListCommentIDsBySearchWordNextForListerInRestrictedTopic(ctx context.Context, arg db.ListCommentIDsBySearchWordNextForListerInRestrictedTopicParams) ([]int32, error) {
	q.fail("ListCommentIDsBySearchWordNextForListerInRestrictedTopic")
	return nil, nil
}

func (q *UnimplementedQuerier) ListCommentIDsBySearchWordNextForListerNotInRestrictedTopic(ctx context.Context, arg db.ListCommentIDsBySearchWordNextForListerNotInRestrictedTopicParams) ([]int32, error) {
	q.fail("ListCommentIDsBySearchWordNextForListerNotInRestrictedTopic")
	return nil, nil
}

func (q *UnimplementedQuerier) ListContentLabelStatus(ctx context.Context, arg db.ListContentLabelStatusParams) ([]*db.ListContentLabelStatusRow, error) {
	q.fail("ListContentLabelStatus")
	return nil, nil
}

func (q *UnimplementedQuerier) ListContentPrivateLabels(ctx context.Context, arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
	q.fail("ListContentPrivateLabels")
	return nil, nil
}

func (q *UnimplementedQuerier) ListContentPublicLabels(ctx context.Context, arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
	q.fail("ListContentPublicLabels")
	return nil, nil
}

func (q *UnimplementedQuerier) ListEffectiveRoleIDsByUserID(ctx context.Context, usersIdusers int32) ([]int32, error) {
	q.fail("ListEffectiveRoleIDsByUserID")
	return nil, nil
}

func (q *UnimplementedQuerier) ListForumcategoryPath(ctx context.Context, categoryID int32) ([]*db.ListForumcategoryPathRow, error) {
	q.fail("ListForumcategoryPath")
	return nil, nil
}

func (q *UnimplementedQuerier) ListGrants(ctx context.Context) ([]*db.Grant, error) {
	q.fail("ListGrants")
	return nil, nil
}

func (q *UnimplementedQuerier) ListGrantsByUserID(ctx context.Context, userID sql.NullInt32) ([]*db.Grant, error) {
	q.fail("ListGrantsByUserID")
	return nil, nil
}

func (q *UnimplementedQuerier) ListImagePostsByBoardForLister(ctx context.Context, arg db.ListImagePostsByBoardForListerParams) ([]*db.ListImagePostsByBoardForListerRow, error) {
	q.fail("ListImagePostsByBoardForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListImagePostsByPosterForLister(ctx context.Context, arg db.ListImagePostsByPosterForListerParams) ([]*db.ListImagePostsByPosterForListerRow, error) {
	q.fail("ListImagePostsByPosterForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListImageboardPath(ctx context.Context, boardID int32) ([]*db.ListImageboardPathRow, error) {
	q.fail("ListImageboardPath")
	return nil, nil
}

func (q *UnimplementedQuerier) ListLinkerCategoryPath(ctx context.Context, categoryID int32) ([]*db.ListLinkerCategoryPathRow, error) {
	q.fail("ListLinkerCategoryPath")
	return nil, nil
}

func (q *UnimplementedQuerier) ListNotificationsForLister(ctx context.Context, arg db.ListNotificationsForListerParams) ([]*db.Notification, error) {
	q.fail("ListNotificationsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListPrivateTopicParticipantsByTopicIDForUser(ctx context.Context, arg db.ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*db.ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
	q.fail("ListPrivateTopicParticipantsByTopicIDForUser")
	return nil, nil
}

func (q *UnimplementedQuerier) ListPrivateTopicsByUserID(ctx context.Context, userID sql.NullInt32) ([]*db.ListPrivateTopicsByUserIDRow, error) {
	q.fail("ListPrivateTopicsByUserID")
	return nil, nil
}

func (q *UnimplementedQuerier) ListPublicWritingsByUserForLister(ctx context.Context, arg db.ListPublicWritingsByUserForListerParams) ([]*db.ListPublicWritingsByUserForListerRow, error) {
	q.fail("ListPublicWritingsByUserForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListPublicWritingsInCategoryForLister(ctx context.Context, arg db.ListPublicWritingsInCategoryForListerParams) ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
	q.fail("ListPublicWritingsInCategoryForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListSiteNewsSearchFirstForLister(ctx context.Context, arg db.ListSiteNewsSearchFirstForListerParams) ([]int32, error) {
	q.fail("ListSiteNewsSearchFirstForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListSiteNewsSearchNextForLister(ctx context.Context, arg db.ListSiteNewsSearchNextForListerParams) ([]int32, error) {
	q.fail("ListSiteNewsSearchNextForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListSubscribersForPattern(ctx context.Context, arg db.ListSubscribersForPatternParams) ([]int32, error) {
	q.fail("ListSubscribersForPattern")
	return nil, nil
}

func (q *UnimplementedQuerier) ListSubscribersForPatterns(ctx context.Context, arg db.ListSubscribersForPatternsParams) ([]int32, error) {
	q.fail("ListSubscribersForPatterns")
	return nil, nil
}

func (q *UnimplementedQuerier) ListSubscriptionsByUser(ctx context.Context, usersIdusers int32) ([]*db.ListSubscriptionsByUserRow, error) {
	q.fail("ListSubscriptionsByUser")
	return nil, nil
}

func (q *UnimplementedQuerier) ListUnreadNotificationsForLister(ctx context.Context, arg db.ListUnreadNotificationsForListerParams) ([]*db.Notification, error) {
	q.fail("ListUnreadNotificationsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListUploadedImagesByUserForLister(ctx context.Context, arg db.ListUploadedImagesByUserForListerParams) ([]*db.UploadedImage, error) {
	q.fail("ListUploadedImagesByUserForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListUserEmailsForLister(ctx context.Context, arg db.ListUserEmailsForListerParams) ([]*db.UserEmail, error) {
	q.fail("ListUserEmailsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListUsersWithRoles(ctx context.Context) ([]*db.ListUsersWithRolesRow, error) {
	q.fail("ListUsersWithRoles")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritersForLister(ctx context.Context, arg db.ListWritersForListerParams) ([]*db.ListWritersForListerRow, error) {
	q.fail("ListWritersForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritersSearchForLister(ctx context.Context, arg db.ListWritersSearchForListerParams) ([]*db.ListWritersSearchForListerRow, error) {
	q.fail("ListWritersSearchForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritingCategoriesForLister(ctx context.Context, arg db.ListWritingCategoriesForListerParams) ([]*db.WritingCategory, error) {
	q.fail("ListWritingCategoriesForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritingSearchFirstForLister(ctx context.Context, arg db.ListWritingSearchFirstForListerParams) ([]int32, error) {
	q.fail("ListWritingSearchFirstForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritingSearchNextForLister(ctx context.Context, arg db.ListWritingSearchNextForListerParams) ([]int32, error) {
	q.fail("ListWritingSearchNextForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritingcategoryPath(ctx context.Context, categoryID int32) ([]*db.ListWritingcategoryPathRow, error) {
	q.fail("ListWritingcategoryPath")
	return nil, nil
}

func (q *UnimplementedQuerier) ListWritingsByIDsForLister(ctx context.Context, arg db.ListWritingsByIDsForListerParams) ([]*db.ListWritingsByIDsForListerRow, error) {
	q.fail("ListWritingsByIDsForLister")
	return nil, nil
}

func (q *UnimplementedQuerier) RemoveContentLabelStatus(ctx context.Context, arg db.RemoveContentLabelStatusParams) error {
	q.fail("RemoveContentLabelStatus")
	return nil
}

func (q *UnimplementedQuerier) RemoveContentPrivateLabel(ctx context.Context, arg db.RemoveContentPrivateLabelParams) error {
	q.fail("RemoveContentPrivateLabel")
	return nil
}

func (q *UnimplementedQuerier) RemoveContentPublicLabel(ctx context.Context, arg db.RemoveContentPublicLabelParams) error {
	q.fail("RemoveContentPublicLabel")
	return nil
}

func (q *UnimplementedQuerier) SetNotificationPriorityForLister(ctx context.Context, arg db.SetNotificationPriorityForListerParams) error {
	q.fail("SetNotificationPriorityForLister")
	return nil
}

func (q *UnimplementedQuerier) SetNotificationReadForLister(ctx context.Context, arg db.SetNotificationReadForListerParams) error {
	q.fail("SetNotificationReadForLister")
	return nil
}

func (q *UnimplementedQuerier) SetNotificationUnreadForLister(ctx context.Context, arg db.SetNotificationUnreadForListerParams) error {
	q.fail("SetNotificationUnreadForLister")
	return nil
}

func (q *UnimplementedQuerier) SetVerificationCodeForLister(ctx context.Context, arg db.SetVerificationCodeForListerParams) error {
	q.fail("SetVerificationCodeForLister")
	return nil
}

func (q *UnimplementedQuerier) SystemAddToBlogsSearch(ctx context.Context, arg db.SystemAddToBlogsSearchParams) error {
	q.fail("SystemAddToBlogsSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemAddToForumCommentSearch(ctx context.Context, arg db.SystemAddToForumCommentSearchParams) error {
	q.fail("SystemAddToForumCommentSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemAddToForumWritingSearch(ctx context.Context, arg db.SystemAddToForumWritingSearchParams) error {
	q.fail("SystemAddToForumWritingSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemAddToImagePostSearch(ctx context.Context, arg db.SystemAddToImagePostSearchParams) error {
	q.fail("SystemAddToImagePostSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemAddToLinkerSearch(ctx context.Context, arg db.SystemAddToLinkerSearchParams) error {
	q.fail("SystemAddToLinkerSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemAddToSiteNewsSearch(ctx context.Context, arg db.SystemAddToSiteNewsSearchParams) error {
	q.fail("SystemAddToSiteNewsSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemAssignBlogEntryThreadID(ctx context.Context, arg db.SystemAssignBlogEntryThreadIDParams) error {
	q.fail("SystemAssignBlogEntryThreadID")
	return nil
}

func (q *UnimplementedQuerier) SystemAssignImagePostThreadID(ctx context.Context, arg db.SystemAssignImagePostThreadIDParams) error {
	q.fail("SystemAssignImagePostThreadID")
	return nil
}

func (q *UnimplementedQuerier) SystemAssignLinkerThreadID(ctx context.Context, arg db.SystemAssignLinkerThreadIDParams) error {
	q.fail("SystemAssignLinkerThreadID")
	return nil
}

func (q *UnimplementedQuerier) SystemAssignNewsThreadID(ctx context.Context, arg db.SystemAssignNewsThreadIDParams) error {
	q.fail("SystemAssignNewsThreadID")
	return nil
}

func (q *UnimplementedQuerier) SystemAssignWritingThreadID(ctx context.Context, arg db.SystemAssignWritingThreadIDParams) error {
	q.fail("SystemAssignWritingThreadID")
	return nil
}

func (q *UnimplementedQuerier) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	q.fail("SystemCheckGrant")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	q.fail("SystemCheckRoleGrant")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemClearContentLabelStatus(ctx context.Context, arg db.SystemClearContentLabelStatusParams) error {
	q.fail("SystemClearContentLabelStatus")
	return nil
}

func (q *UnimplementedQuerier) SystemClearContentPrivateLabel(ctx context.Context, arg db.SystemClearContentPrivateLabelParams) error {
	q.fail("SystemClearContentPrivateLabel")
	return nil
}

func (q *UnimplementedQuerier) SystemCountDeadLetters(ctx context.Context) (int64, error) {
	q.fail("SystemCountDeadLetters")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCountLanguages(ctx context.Context) (int64, error) {
	q.fail("SystemCountLanguages")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCountRecentLoginAttempts(ctx context.Context, arg db.SystemCountRecentLoginAttemptsParams) (int64, error) {
	q.fail("SystemCountRecentLoginAttempts")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCreateGrant(ctx context.Context, arg db.SystemCreateGrantParams) (int64, error) {
	q.fail("SystemCreateGrant")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCreateNotification(ctx context.Context, arg db.SystemCreateNotificationParams) error {
	q.fail("SystemCreateNotification")
	return nil
}

func (q *UnimplementedQuerier) SystemCreateSearchWord(ctx context.Context, word string) (int64, error) {
	q.fail("SystemCreateSearchWord")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCreateThread(ctx context.Context, forumtopicIdforumtopic int32) (int64, error) {
	q.fail("SystemCreateThread")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemCreateUserRole(ctx context.Context, arg db.SystemCreateUserRoleParams) error {
	q.fail("SystemCreateUserRole")
	return nil
}

func (q *UnimplementedQuerier) SystemCreateUserRoleByID(ctx context.Context, arg db.SystemCreateUserRoleByIDParams) error {
	q.fail("SystemCreateUserRoleByID")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteBlogsSearch(ctx context.Context) error {
	q.fail("SystemDeleteBlogsSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteCommentsSearch(ctx context.Context) error {
	q.fail("SystemDeleteCommentsSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteDeadLetter(ctx context.Context, id int32) error {
	q.fail("SystemDeleteDeadLetter")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteImagePostSearch(ctx context.Context) error {
	q.fail("SystemDeleteImagePostSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteLinkerSearch(ctx context.Context) error {
	q.fail("SystemDeleteLinkerSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemDeletePasswordReset(ctx context.Context, id int32) error {
	q.fail("SystemDeletePasswordReset")
	return nil
}

func (q *UnimplementedQuerier) SystemDeletePasswordResetsByUser(ctx context.Context, userID int32) (sql.Result, error) {
	q.fail("SystemDeletePasswordResetsByUser")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemDeleteSessionByID(ctx context.Context, sessionID string) error {
	q.fail("SystemDeleteSessionByID")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteSiteNewsSearch(ctx context.Context) error {
	q.fail("SystemDeleteSiteNewsSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteUserEmailsByEmailExceptID(ctx context.Context, arg db.SystemDeleteUserEmailsByEmailExceptIDParams) error {
	q.fail("SystemDeleteUserEmailsByEmailExceptID")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteWritingSearch(ctx context.Context) error {
	q.fail("SystemDeleteWritingSearch")
	return nil
}

func (q *UnimplementedQuerier) SystemDeleteWritingSearchByWritingID(ctx context.Context, writingID int32) error {
	q.fail("SystemDeleteWritingSearchByWritingID")
	return nil
}

func (q *UnimplementedQuerier) SystemGetAllBlogsForIndex(ctx context.Context) ([]*db.SystemGetAllBlogsForIndexRow, error) {
	q.fail("SystemGetAllBlogsForIndex")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetBlogEntryByID(ctx context.Context, idblogs int32) (*db.SystemGetBlogEntryByIDRow, error) {
	q.fail("SystemGetBlogEntryByID")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetFAQQuestions(ctx context.Context) ([]*db.Faq, error) {
	q.fail("SystemGetFAQQuestions")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetForumTopicByTitle(ctx context.Context, title sql.NullString) (*db.Forumtopic, error) {
	q.fail("SystemGetForumTopicByTitle")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetLanguageIDByName(ctx context.Context, nameof sql.NullString) (int32, error) {
	q.fail("SystemGetLanguageIDByName")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemGetLastNotificationForRecipientByMessage(ctx context.Context, arg db.SystemGetLastNotificationForRecipientByMessageParams) (*db.Notification, error) {
	q.fail("SystemGetLastNotificationForRecipientByMessage")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetLogin(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
	q.fail("SystemGetLogin")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetNewsPostByID(ctx context.Context, idsitenews int32) (int32, error) {
	q.fail("SystemGetNewsPostByID")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemGetSearchWordByWordLowercased(ctx context.Context, lcase string) (*db.Searchwordlist, error) {
	q.fail("SystemGetSearchWordByWordLowercased")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetTemplateOverride(ctx context.Context, name string) (string, error) {
	q.fail("SystemGetTemplateOverride")
	return "", nil
}

func (q *UnimplementedQuerier) SystemGetUserByEmail(ctx context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
	q.fail("SystemGetUserByEmail")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetUserByID(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
	q.fail("SystemGetUserByID")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetUserByUsername(ctx context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
	q.fail("SystemGetUserByUsername")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemGetWritingByID(ctx context.Context, idwriting int32) (int32, error) {
	q.fail("SystemGetWritingByID")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemIncrementPendingEmailError(ctx context.Context, id int32) error {
	q.fail("SystemIncrementPendingEmailError")
	return nil
}

func (q *UnimplementedQuerier) SystemInsertDeadLetter(ctx context.Context, message string) error {
	q.fail("SystemInsertDeadLetter")
	return nil
}

func (q *UnimplementedQuerier) SystemInsertLoginAttempt(ctx context.Context, arg db.SystemInsertLoginAttemptParams) error {
	q.fail("SystemInsertLoginAttempt")
	return nil
}

func (q *UnimplementedQuerier) SystemInsertSession(ctx context.Context, arg db.SystemInsertSessionParams) error {
	q.fail("SystemInsertSession")
	return nil
}

func (q *UnimplementedQuerier) SystemInsertUser(ctx context.Context, username sql.NullString) (int64, error) {
	q.fail("SystemInsertUser")
	return 0, nil
}

func (q *UnimplementedQuerier) SystemLatestDeadLetter(ctx context.Context) (interface{}, error) {
	q.fail("SystemLatestDeadLetter")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListAllUsers(ctx context.Context) ([]*db.SystemListAllUsersRow, error) {
	q.fail("SystemListAllUsers")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListBoardsByParentID(ctx context.Context, arg db.SystemListBoardsByParentIDParams) ([]*db.Imageboard, error) {
	q.fail("SystemListBoardsByParentID")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListCommentsByThreadID(ctx context.Context, forumthreadID int32) ([]*db.SystemListCommentsByThreadIDRow, error) {
	q.fail("SystemListCommentsByThreadID")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListDeadLetters(ctx context.Context, limit int32) ([]*db.DeadLetter, error) {
	q.fail("SystemListDeadLetters")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListLanguages(ctx context.Context) ([]*db.Language, error) {
	q.fail("SystemListLanguages")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListPendingEmails(ctx context.Context, arg db.SystemListPendingEmailsParams) ([]*db.SystemListPendingEmailsRow, error) {
	q.fail("SystemListPendingEmails")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListPublicWritingsByAuthor(ctx context.Context, arg db.SystemListPublicWritingsByAuthorParams) ([]*db.SystemListPublicWritingsByAuthorRow, error) {
	q.fail("SystemListPublicWritingsByAuthor")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListPublicWritingsInCategory(ctx context.Context, arg db.SystemListPublicWritingsInCategoryParams) ([]*db.SystemListPublicWritingsInCategoryRow, error) {
	q.fail("SystemListPublicWritingsInCategory")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListUserInfo(ctx context.Context) ([]*db.SystemListUserInfoRow, error) {
	q.fail("SystemListUserInfo")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListVerifiedEmailsByUserID(ctx context.Context, userID int32) ([]*db.UserEmail, error) {
	q.fail("SystemListVerifiedEmailsByUserID")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemListWritingCategories(ctx context.Context, arg db.SystemListWritingCategoriesParams) ([]*db.WritingCategory, error) {
	q.fail("SystemListWritingCategories")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemMarkPasswordResetVerified(ctx context.Context, id int32) error {
	q.fail("SystemMarkPasswordResetVerified")
	return nil
}

func (q *UnimplementedQuerier) SystemMarkPendingEmailSent(ctx context.Context, id int32) error {
	q.fail("SystemMarkPendingEmailSent")
	return nil
}

func (q *UnimplementedQuerier) SystemMarkUserEmailVerified(ctx context.Context, arg db.SystemMarkUserEmailVerifiedParams) error {
	q.fail("SystemMarkUserEmailVerified")
	return nil
}

func (q *UnimplementedQuerier) SystemPurgeDeadLettersBefore(ctx context.Context, createdAt time.Time) error {
	q.fail("SystemPurgeDeadLettersBefore")
	return nil
}

func (q *UnimplementedQuerier) SystemPurgePasswordResetsBefore(ctx context.Context, createdAt time.Time) (sql.Result, error) {
	q.fail("SystemPurgePasswordResetsBefore")
	return nil, nil
}

func (q *UnimplementedQuerier) SystemRebuildForumTopicMetaByID(ctx context.Context, idforumtopic int32) error {
	q.fail("SystemRebuildForumTopicMetaByID")
	return nil
}

func (q *UnimplementedQuerier) SystemRegisterExternalLinkClick(ctx context.Context, url string) error {
	q.fail("SystemRegisterExternalLinkClick")
	return nil
}

func (q *UnimplementedQuerier) SystemSetBlogLastIndex(ctx context.Context, id int32) error {
	q.fail("SystemSetBlogLastIndex")
	return nil
}

func (q *UnimplementedQuerier) SystemSetCommentLastIndex(ctx context.Context, idcomments int32) error {
	q.fail("SystemSetCommentLastIndex")
	return nil
}

func (q *UnimplementedQuerier) SystemSetForumTopicHandlerByID(ctx context.Context, arg db.SystemSetForumTopicHandlerByIDParams) error {
	q.fail("SystemSetForumTopicHandlerByID")
	return nil
}

func (q *UnimplementedQuerier) SystemSetImagePostLastIndex(ctx context.Context, idimagepost int32) error {
	q.fail("SystemSetImagePostLastIndex")
	return nil
}

func (q *UnimplementedQuerier) SystemSetLinkerLastIndex(ctx context.Context, id int32) error {
	q.fail("SystemSetLinkerLastIndex")
	return nil
}

func (q *UnimplementedQuerier) SystemSetSiteNewsLastIndex(ctx context.Context, idsitenews int32) error {
	q.fail("SystemSetSiteNewsLastIndex")
	return nil
}

func (q *UnimplementedQuerier) SystemSetWritingLastIndex(ctx context.Context, idwriting int32) error {
	q.fail("SystemSetWritingLastIndex")
	return nil
}

func (q *UnimplementedQuerier) UpdateAutoSubscribeRepliesForLister(ctx context.Context, arg db.UpdateAutoSubscribeRepliesForListerParams) error {
	q.fail("UpdateAutoSubscribeRepliesForLister")
	return nil
}

func (q *UnimplementedQuerier) UpdateBlogEntryForWriter(ctx context.Context, arg db.UpdateBlogEntryForWriterParams) error {
	q.fail("UpdateBlogEntryForWriter")
	return nil
}

func (q *UnimplementedQuerier) UpdateBookmarksForLister(ctx context.Context, arg db.UpdateBookmarksForListerParams) error {
	q.fail("UpdateBookmarksForLister")
	return nil
}

func (q *UnimplementedQuerier) UpdateCommentForEditor(ctx context.Context, arg db.UpdateCommentForEditorParams) error {
	q.fail("UpdateCommentForEditor")
	return nil
}

func (q *UnimplementedQuerier) UpdateEmailForumUpdatesForLister(ctx context.Context, arg db.UpdateEmailForumUpdatesForListerParams) error {
	q.fail("UpdateEmailForumUpdatesForLister")
	return nil
}

func (q *UnimplementedQuerier) UpdateNewsPostForWriter(ctx context.Context, arg db.UpdateNewsPostForWriterParams) error {
	q.fail("UpdateNewsPostForWriter")
	return nil
}

func (q *UnimplementedQuerier) UpdatePreferenceForLister(ctx context.Context, arg db.UpdatePreferenceForListerParams) error {
	q.fail("UpdatePreferenceForLister")
	return nil
}

func (q *UnimplementedQuerier) UpdatePublicProfileEnabledAtForUser(ctx context.Context, arg db.UpdatePublicProfileEnabledAtForUserParams) error {
	q.fail("UpdatePublicProfileEnabledAtForUser")
	return nil
}

func (q *UnimplementedQuerier) UpdateSubscriptionByIDForSubscriber(ctx context.Context, arg db.UpdateSubscriptionByIDForSubscriberParams) error {
	q.fail("UpdateSubscriptionByIDForSubscriber")
	return nil
}

func (q *UnimplementedQuerier) UpdateTimezoneForLister(ctx context.Context, arg db.UpdateTimezoneForListerParams) error {
	q.fail("UpdateTimezoneForLister")
	return nil
}

func (q *UnimplementedQuerier) UpdateWritingForWriter(ctx context.Context, arg db.UpdateWritingForWriterParams) error {
	q.fail("UpdateWritingForWriter")
	return nil
}

func (q *UnimplementedQuerier) UpsertContentReadMarker(ctx context.Context, arg db.UpsertContentReadMarkerParams) error {
	q.fail("UpsertContentReadMarker")
	return nil
}
