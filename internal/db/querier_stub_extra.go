package db

import (
	"context"
	"database/sql"
	"time"
)

func (s *QuerierStub) AdminGetRequestByID(ctx context.Context, id int32) (*AdminRequestQueue, error) {
	s.mu.Lock()
	s.AdminGetRequestByIDCalls = append(s.AdminGetRequestByIDCalls, id)
	fn := s.AdminGetRequestByIDFn
	ret := s.AdminGetRequestByIDReturns
	err := s.AdminGetRequestByIDErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return ret, err
}

func (s *QuerierStub) AdminListRequestComments(ctx context.Context, requestID int32) ([]*AdminRequestComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListRequestCommentsCalls = append(s.AdminListRequestCommentsCalls, requestID)
	if s.AdminListRequestCommentsFn != nil {
		return s.AdminListRequestCommentsFn(ctx, requestID)
	}
	return s.AdminListRequestCommentsReturns, s.AdminListRequestCommentsErr
}

func (s *QuerierStub) AdminScrubComment(ctx context.Context, arg AdminScrubCommentParams) error {
	s.mu.Lock()
	s.AdminScrubCommentCalls = append(s.AdminScrubCommentCalls, arg)
	fn := s.AdminScrubCommentFn
	err := s.AdminScrubCommentErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminArchiveComment(ctx context.Context, arg AdminArchiveCommentParams) error {
	s.mu.Lock()
	s.AdminArchiveCommentCalls = append(s.AdminArchiveCommentCalls, arg)
	fn := s.AdminArchiveCommentFn
	err := s.AdminArchiveCommentErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminIsCommentDeactivated(ctx context.Context, id int32) (bool, error) {
	s.mu.Lock()
	s.AdminIsCommentDeactivatedCalls = append(s.AdminIsCommentDeactivatedCalls, id)
	fn := s.AdminIsCommentDeactivatedFn
	ret := s.AdminIsCommentDeactivatedReturns
	err := s.AdminIsCommentDeactivatedErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return ret, err
}

func (s *QuerierStub) AdminRestoreComment(ctx context.Context, arg AdminRestoreCommentParams) error {
	s.mu.Lock()
	s.AdminRestoreCommentCalls = append(s.AdminRestoreCommentCalls, arg)
	fn := s.AdminRestoreCommentFn
	err := s.AdminRestoreCommentErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminListDeactivatedComments(ctx context.Context, arg AdminListDeactivatedCommentsParams) ([]*AdminListDeactivatedCommentsRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListDeactivatedCommentsCalls = append(s.AdminListDeactivatedCommentsCalls, arg)
	if s.AdminListDeactivatedCommentsFn != nil {
		return s.AdminListDeactivatedCommentsFn(ctx, arg)
	}
	return s.AdminListDeactivatedCommentsReturns, s.AdminListDeactivatedCommentsErr
}

func (s *QuerierStub) AdminMarkCommentRestored(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.AdminMarkCommentRestoredCalls = append(s.AdminMarkCommentRestoredCalls, id)
	fn := s.AdminMarkCommentRestoredFn
	err := s.AdminMarkCommentRestoredErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) AdminCountAllImagePosts(ctx context.Context) (int64, error) {
	s.mu.Lock()
	s.AdminCountAllImagePostsCalls++
	fn := s.AdminCountAllImagePostsFn
	ret := s.AdminCountAllImagePostsReturns
	err := s.AdminCountAllImagePostsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminListAllImagePosts(ctx context.Context, arg AdminListAllImagePostsParams) ([]*AdminListAllImagePostsRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListAllImagePostsCalls = append(s.AdminListAllImagePostsCalls, arg)
	if s.AdminListAllImagePostsFn != nil {
		return s.AdminListAllImagePostsFn(ctx, arg)
	}
	return s.AdminListAllImagePostsReturns, s.AdminListAllImagePostsErr
}

func (s *QuerierStub) GetImagePostInfoByPath(ctx context.Context, arg GetImagePostInfoByPathParams) (*GetImagePostInfoByPathRow, error) {
	s.mu.Lock()
	s.GetImagePostInfoByPathCalls = append(s.GetImagePostInfoByPathCalls, arg)
	fn := s.GetImagePostInfoByPathFn
	row := s.GetImagePostInfoByPathRow
	err := s.GetImagePostInfoByPathErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return row, err
}

func (s *QuerierStub) GetCommentsByIdsForUserWithThreadInfo(ctx context.Context, arg GetCommentsByIdsForUserWithThreadInfoParams) ([]*GetCommentsByIdsForUserWithThreadInfoRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetCommentsByIdsForUserWithThreadInfoCalls = append(s.GetCommentsByIdsForUserWithThreadInfoCalls, arg)
	if s.GetCommentsByIdsForUserWithThreadInfoFn != nil {
		return s.GetCommentsByIdsForUserWithThreadInfoFn(ctx, arg)
	}
	return s.GetCommentsByIdsForUserWithThreadInfoReturns, s.GetCommentsByIdsForUserWithThreadInfoErr
}

func (s *QuerierStub) AdminInsertBannedIp(ctx context.Context, arg AdminInsertBannedIpParams) error {
	s.mu.Lock()
	s.AdminInsertBannedIpCalls = append(s.AdminInsertBannedIpCalls, arg)
	fn := s.AdminInsertBannedIpFn
	err := s.AdminInsertBannedIpErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminGetImagePost(ctx context.Context, idimagepost int32) (*AdminGetImagePostRow, error) {
	s.mu.Lock()
	s.AdminGetImagePostCalls = append(s.AdminGetImagePostCalls, idimagepost)
	fn := s.AdminGetImagePostFn
	row := s.AdminGetImagePostRow
	err := s.AdminGetImagePostErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, idimagepost)
	}
	return row, err
}

func (s *QuerierStub) AdminApproveImagePost(ctx context.Context, idimagepost int32) error {
	s.mu.Lock()
	s.AdminApproveImagePostCalls = append(s.AdminApproveImagePostCalls, idimagepost)
	fn := s.AdminApproveImagePostFn
	err := s.AdminApproveImagePostErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, idimagepost)
	}
	return err
}

func (s *QuerierStub) InsertPassword(ctx context.Context, arg InsertPasswordParams) error {
	s.mu.Lock()
	s.InsertPasswordCalls = append(s.InsertPasswordCalls, arg)
	fn := s.InsertPasswordFn
	err := s.InsertPasswordErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) SystemGetUserByUsername(ctx context.Context, arg sql.NullString) (*SystemGetUserByUsernameRow, error) {
	s.mu.Lock()
	s.SystemGetUserByUsernameCalls = append(s.SystemGetUserByUsernameCalls, arg)
	fn := s.SystemGetUserByUsernameFn
	row := s.SystemGetUserByUsernameRow
	err := s.SystemGetUserByUsernameErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return row, err
}

func (s *QuerierStub) SystemDeletePasswordResetsByUser(ctx context.Context, userID int32) (sql.Result, error) {
	s.mu.Lock()
	s.SystemDeletePasswordResetsByUserCalls = append(s.SystemDeletePasswordResetsByUserCalls, userID)
	fn := s.SystemDeletePasswordResetsByUserFn
	ret := s.SystemDeletePasswordResetsByUserResult
	err := s.SystemDeletePasswordResetsByUserErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, userID)
	}
	return ret, err
}

func (s *QuerierStub) AdminUpdateFAQ(ctx context.Context, arg AdminUpdateFAQParams) error {
	s.mu.Lock()
	s.AdminUpdateFAQCalls = append(s.AdminUpdateFAQCalls, arg)
	fn := s.AdminUpdateFAQFn
	err := s.AdminUpdateFAQErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminCountPasswordResets(ctx context.Context, arg AdminCountPasswordResetsParams) (int64, error) {
	s.mu.Lock()
	s.AdminCountPasswordResetsCalls = append(s.AdminCountPasswordResetsCalls, arg)
	fn := s.AdminCountPasswordResetsFn
	ret := s.AdminCountPasswordResetsReturns
	err := s.AdminCountPasswordResetsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) AdminInsertRequestComment(ctx context.Context, arg AdminInsertRequestCommentParams) error {
	s.mu.Lock()
	s.AdminInsertRequestCommentCalls = append(s.AdminInsertRequestCommentCalls, arg)
	fn := s.AdminInsertRequestCommentFn
	err := s.AdminInsertRequestCommentErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) InsertAdminUserComment(ctx context.Context, arg InsertAdminUserCommentParams) error {
	s.mu.Lock()
	s.InsertAdminUserCommentCalls = append(s.InsertAdminUserCommentCalls, arg)
	fn := s.InsertAdminUserCommentFn
	err := s.InsertAdminUserCommentErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminCountPendingPasswordResetsByUser(ctx context.Context) ([]*AdminCountPendingPasswordResetsByUserRow, error) {
	s.mu.Lock()
	s.AdminCountPendingPasswordResetsByUserCalls++
	fn := s.AdminCountPendingPasswordResetsByUserFn
	ret := s.AdminCountPendingPasswordResetsByUserReturns
	err := s.AdminCountPendingPasswordResetsByUserErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetPasswordResetByID(ctx context.Context, id int32) (*PendingPassword, error) {
	s.mu.Lock()
	s.AdminGetPasswordResetByIDCalls = append(s.AdminGetPasswordResetByIDCalls, id)
	fn := s.AdminGetPasswordResetByIDFn
	ret := s.AdminGetPasswordResetByIDReturns
	err := s.AdminGetPasswordResetByIDErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return ret, err
}

func (s *QuerierStub) AdminListPasswordResets(ctx context.Context, arg AdminListPasswordResetsParams) ([]*AdminListPasswordResetsRow, error) {
	s.mu.Lock()
	s.AdminListPasswordResetsCalls = append(s.AdminListPasswordResetsCalls, arg)
	fn := s.AdminListPasswordResetsFn
	ret := s.AdminListPasswordResetsReturns
	err := s.AdminListPasswordResetsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) DeletePendingPassword(ctx context.Context, userID int32) error {
	s.mu.Lock()
	s.DeletePendingPasswordCalls = append(s.DeletePendingPasswordCalls, userID)
	fn := s.DeletePendingPasswordFn
	err := s.DeletePendingPasswordErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, userID)
	}
	return err
}

func (s *QuerierStub) GetPasswordResetByCode(ctx context.Context, arg GetPasswordResetByCodeParams) (*PendingPassword, error) {
	s.mu.Lock()
	s.GetPasswordResetByCodeCalls = append(s.GetPasswordResetByCodeCalls, arg)
	fn := s.GetPasswordResetByCodeFn
	ret := s.GetPasswordResetByCodeReturns
	err := s.GetPasswordResetByCodeErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) GetPendingPassword(ctx context.Context, userID int32) (*PendingPassword, error) {
	s.mu.Lock()
	s.GetPendingPasswordCalls = append(s.GetPendingPasswordCalls, userID)
	fn := s.GetPendingPasswordFn
	ret := s.GetPendingPasswordReturns
	err := s.GetPendingPasswordErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, userID)
	}
	return ret, err
}

func (s *QuerierStub) GetPendingPasswordByCode(ctx context.Context, verificationCode string) (*PendingPassword, error) {
	s.mu.Lock()
	s.GetPendingPasswordByCodeCalls = append(s.GetPendingPasswordByCodeCalls, verificationCode)
	fn := s.GetPendingPasswordByCodeFn
	row := s.GetPendingPasswordByCodeRow
	err := s.GetPendingPasswordByCodeErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, verificationCode)
	}
	return row, err
}

func (s *QuerierStub) AdminPromoteAnnouncement(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.AdminPromoteAnnouncementCalls = append(s.AdminPromoteAnnouncementCalls, id)
	fn := s.AdminPromoteAnnouncementFn
	err := s.AdminPromoteAnnouncementErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) SystemDeletePasswordReset(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.SystemDeletePasswordResetCalls = append(s.SystemDeletePasswordResetCalls, id)
	fn := s.SystemDeletePasswordResetFn
	err := s.SystemDeletePasswordResetErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) SystemMarkPasswordResetVerified(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.SystemMarkPasswordResetVerifiedCalls = append(s.SystemMarkPasswordResetVerifiedCalls, id)
	fn := s.SystemMarkPasswordResetVerifiedFn
	err := s.SystemMarkPasswordResetVerifiedErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) SystemPurgePasswordResetsBefore(ctx context.Context, createdAt time.Time) (sql.Result, error) {
	s.mu.Lock()
	s.SystemPurgePasswordResetsBeforeCalls = append(s.SystemPurgePasswordResetsBeforeCalls, createdAt)
	fn := s.SystemPurgePasswordResetsBeforeFn
	ret := s.SystemPurgePasswordResetsBeforeResult
	err := s.SystemPurgePasswordResetsBeforeErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, createdAt)
	}
	return ret, err
}

func (s *QuerierStub) AdminDemoteAnnouncement(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.AdminDemoteAnnouncementCalls = append(s.AdminDemoteAnnouncementCalls, id)
	fn := s.AdminDemoteAnnouncementFn
	err := s.AdminDemoteAnnouncementErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) AdminDeleteForumThread(ctx context.Context, idforumthread int32) error {
	s.mu.Lock()
	s.AdminDeleteForumThreadCalls = append(s.AdminDeleteForumThreadCalls, idforumthread)
	fn := s.AdminDeleteForumThreadFn
	err := s.AdminDeleteForumThreadErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, idforumthread)
	}
	return err
}

func (s *QuerierStub) AdminCancelBannedIp(ctx context.Context, ip string) error {
	s.mu.Lock()
	s.AdminCancelBannedIpCalls = append(s.AdminCancelBannedIpCalls, ip)
	fn := s.AdminCancelBannedIpFn
	err := s.AdminCancelBannedIpErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, ip)
	}
	return err
}

func (s *QuerierStub) SystemListPendingEmails(ctx context.Context, arg SystemListPendingEmailsParams) ([]*SystemListPendingEmailsRow, error) {
	s.mu.Lock()
	s.SystemListPendingEmailsCalls = append(s.SystemListPendingEmailsCalls, arg)
	fn := s.SystemListPendingEmailsFn
	ret := s.SystemListPendingEmailsReturn
	err := s.SystemListPendingEmailsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) SystemMarkPendingEmailSent(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.SystemMarkPendingEmailSentCalls = append(s.SystemMarkPendingEmailSentCalls, id)
	fn := s.SystemMarkPendingEmailSentFn
	err := s.SystemMarkPendingEmailSentErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) AdminDeletePendingEmail(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.AdminDeletePendingEmailCalls = append(s.AdminDeletePendingEmailCalls, id)
	fn := s.AdminDeletePendingEmailFn
	err := s.AdminDeletePendingEmailErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) SystemGetLogin(ctx context.Context, arg sql.NullString) (*SystemGetLoginRow, error) {
	s.mu.Lock()
	s.SystemGetLoginCalls = append(s.SystemGetLoginCalls, arg)
	fn := s.SystemGetLoginFn
	row := s.SystemGetLoginRow
	err := s.SystemGetLoginErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return row, err
}

func (s *QuerierStub) SystemListVerifiedEmailsByUserID(ctx context.Context, userID int32) ([]*UserEmail, error) {
	s.mu.Lock()
	s.SystemListVerifiedEmailsByUserIDCalls = append(s.SystemListVerifiedEmailsByUserIDCalls, userID)
	fn := s.SystemListVerifiedEmailsByUserIDFn
	ret := s.SystemListVerifiedEmailsByUserIDReturn
	err := s.SystemListVerifiedEmailsByUserIDErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, userID)
	}
	return ret, err
}

func (s *QuerierStub) SystemRebuildForumTopicMetaByID(ctx context.Context, idforumtopic int32) error {
	s.mu.Lock()
	s.SystemRebuildForumTopicMetaByIDCalls = append(s.SystemRebuildForumTopicMetaByIDCalls, idforumtopic)
	fn := s.SystemRebuildForumTopicMetaByIDFn
	err := s.SystemRebuildForumTopicMetaByIDErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, idforumtopic)
	}
	return err
}

func (s *QuerierStub) GetLoginRoleForUser(ctx context.Context, userID int32) (int32, error) {
	s.mu.Lock()
	s.GetLoginRoleForUserCalls = append(s.GetLoginRoleForUserCalls, userID)
	fn := s.GetLoginRoleForUserFn
	ret := s.GetLoginRoleForUserReturns
	err := s.GetLoginRoleForUserErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, userID)
	}
	return ret, err
}

func (s *QuerierStub) ListBlogEntriesByAuthorForLister(ctx context.Context, arg ListBlogEntriesByAuthorForListerParams) ([]*ListBlogEntriesByAuthorForListerRow, error) {
	s.mu.Lock()
	s.ListBlogEntriesByAuthorForListerCalls = append(s.ListBlogEntriesByAuthorForListerCalls, arg)
	fn := s.ListBlogEntriesByAuthorForListerFn
	ret := s.ListBlogEntriesByAuthorForListerReturns
	err := s.ListBlogEntriesByAuthorForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) ListBlogEntriesByIDsForLister(ctx context.Context, arg ListBlogEntriesByIDsForListerParams) ([]*ListBlogEntriesByIDsForListerRow, error) {
	s.mu.Lock()
	s.ListBlogEntriesByIDsForListerCalls = append(s.ListBlogEntriesByIDsForListerCalls, arg)
	fn := s.ListBlogEntriesByIDsForListerFn
	ret := s.ListBlogEntriesByIDsForListerReturns
	err := s.ListBlogEntriesByIDsForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) SystemIncrementPendingEmailError(ctx context.Context, id int32) error {
	s.mu.Lock()
	s.SystemIncrementPendingEmailErrorCalls = append(s.SystemIncrementPendingEmailErrorCalls, id)
	fn := s.SystemIncrementPendingEmailErrorFn
	err := s.SystemIncrementPendingEmailErrorErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return err
}

func (s *QuerierStub) GetPendingEmailErrorCount(ctx context.Context, id int32) (int32, error) {
	s.mu.Lock()
	s.GetPendingEmailErrorCountCalls = append(s.GetPendingEmailErrorCountCalls, id)
	fn := s.GetPendingEmailErrorCountFn
	ret := s.GetPendingEmailErrorCountReturns
	err := s.GetPendingEmailErrorCountErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, id)
	}
	return ret, err
}

func (s *QuerierStub) GetAllAnsweredFAQWithFAQCategoriesForUser(ctx context.Context, arg GetAllAnsweredFAQWithFAQCategoriesForUserParams) ([]*GetAllAnsweredFAQWithFAQCategoriesForUserRow, error) {
	s.mu.Lock()
	s.GetAllAnsweredFAQWithFAQCategoriesForUserCalls = append(s.GetAllAnsweredFAQWithFAQCategoriesForUserCalls, arg)
	fn := s.GetAllAnsweredFAQWithFAQCategoriesForUserFn
	ret := s.GetAllAnsweredFAQWithFAQCategoriesForUserReturns
	err := s.GetAllAnsweredFAQWithFAQCategoriesForUserErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetFAQActiveQuestions(ctx context.Context) ([]*Faq, error) {
	s.mu.Lock()
	s.AdminGetFAQActiveQuestionsCalls++
	fn := s.AdminGetFAQActiveQuestionsFn
	ret := s.AdminGetFAQActiveQuestionsReturns
	err := s.AdminGetFAQActiveQuestionsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetFAQCategories(ctx context.Context) ([]*FaqCategory, error) {
	s.mu.Lock()
	s.AdminGetFAQCategoriesCalls++
	fn := s.AdminGetFAQCategoriesFn
	ret := s.AdminGetFAQCategoriesReturns
	err := s.AdminGetFAQCategoriesErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetFAQUnansweredQuestions(ctx context.Context) ([]*Faq, error) {
	s.mu.Lock()
	s.AdminGetFAQUnansweredQuestionsCalls++
	fn := s.AdminGetFAQUnansweredQuestionsFn
	ret := s.AdminGetFAQUnansweredQuestionsReturns
	err := s.AdminGetFAQUnansweredQuestionsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetFAQDismissedQuestions(ctx context.Context) ([]*AdminGetFAQDismissedQuestionsRow, error) {
	s.mu.Lock()
	s.AdminGetFAQDismissedQuestionsCalls++
	fn := s.AdminGetFAQDismissedQuestionsFn
	ret := s.AdminGetFAQDismissedQuestionsReturns
	err := s.AdminGetFAQDismissedQuestionsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetAllWritingsByAuthor(ctx context.Context, authorID int32) ([]*AdminGetAllWritingsByAuthorRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminGetAllWritingsByAuthorCalls = append(s.AdminGetAllWritingsByAuthorCalls, authorID)
	if s.AdminGetAllWritingsByAuthorFn != nil {
		return s.AdminGetAllWritingsByAuthorFn(ctx, authorID)
	}
	return s.AdminGetAllWritingsByAuthorReturns, s.AdminGetAllWritingsByAuthorErr
}

func (s *QuerierStub) AdminGetUserEmailByID(ctx context.Context, id int32) (*UserEmail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminGetUserEmailByIDCalls = append(s.AdminGetUserEmailByIDCalls, id)
	if s.AdminGetUserEmailByIDFn != nil {
		return s.AdminGetUserEmailByIDFn(ctx, id)
	}
	return s.AdminGetUserEmailByIDReturns, s.AdminGetUserEmailByIDErr
}

func (s *QuerierStub) InsertUserEmail(ctx context.Context, arg InsertUserEmailParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InsertUserEmailCalls = append(s.InsertUserEmailCalls, arg)
	if s.InsertUserEmailFn != nil {
		return s.InsertUserEmailFn(ctx, arg)
	}
	return s.InsertUserEmailErr
}

func (s *QuerierStub) AdminDeleteUserEmail(ctx context.Context, id int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminDeleteUserEmailCalls = append(s.AdminDeleteUserEmailCalls, id)
	if s.AdminDeleteUserEmailFn != nil {
		return s.AdminDeleteUserEmailFn(ctx, id)
	}
	return s.AdminDeleteUserEmailErr
}

func (s *QuerierStub) AdminUpdateUserEmailDetails(ctx context.Context, arg AdminUpdateUserEmailDetailsParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminUpdateUserEmailDetailsCalls = append(s.AdminUpdateUserEmailDetailsCalls, arg)
	if s.AdminUpdateUserEmailDetailsFn != nil {
		return s.AdminUpdateUserEmailDetailsFn(ctx, arg)
	}
	return s.AdminUpdateUserEmailDetailsErr
}

func (s *QuerierStub) SystemUpdateVerificationCode(ctx context.Context, arg SystemUpdateVerificationCodeParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SystemUpdateVerificationCodeCalls = append(s.SystemUpdateVerificationCodeCalls, arg)
	if s.SystemUpdateVerificationCodeFn != nil {
		return s.SystemUpdateVerificationCodeFn(ctx, arg)
	}
	return s.SystemUpdateVerificationCodeErr
}

func (s *QuerierStub) AdminListRecentNotifications(ctx context.Context, limit int32) ([]*Notification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListRecentNotificationsCalls = append(s.AdminListRecentNotificationsCalls, limit)
	if s.AdminListRecentNotificationsFn != nil {
		return s.AdminListRecentNotificationsFn(ctx, limit)
	}
	return s.AdminListRecentNotificationsReturns, s.AdminListRecentNotificationsErr
}

func (s *QuerierStub) AdminListAnnouncementsWithNews(ctx context.Context) ([]*AdminListAnnouncementsWithNewsRow, error) {
	s.mu.Lock()
	s.AdminListAnnouncementsWithNewsCalls++
	fn := s.AdminListAnnouncementsWithNewsFn
	ret := s.AdminListAnnouncementsWithNewsReturns
	err := s.AdminListAnnouncementsWithNewsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) ListBannedIps(ctx context.Context) ([]*BannedIp, error) {
	s.mu.Lock()
	s.ListBannedIpsCalls++
	fn := s.ListBannedIpsFn
	ret := s.ListBannedIpsReturns
	err := s.ListBannedIpsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminListAllCommentsWithThreadInfo(ctx context.Context, arg AdminListAllCommentsWithThreadInfoParams) ([]*AdminListAllCommentsWithThreadInfoRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListAllCommentsWithThreadInfoCalls = append(s.AdminListAllCommentsWithThreadInfoCalls, arg)
	return s.AdminListAllCommentsWithThreadInfoReturns, s.AdminListAllCommentsWithThreadInfoErr
}

func (s *QuerierStub) AdminCreateGrant(ctx context.Context, arg AdminCreateGrantParams) (int64, error) {
	s.mu.Lock()
	s.AdminCreateGrantCalls = append(s.AdminCreateGrantCalls, arg)
	fn := s.AdminCreateGrantFn
	ret := s.AdminCreateGrantReturns
	err := s.AdminCreateGrantErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) UpdateSubscriptionByIDForSubscriber(ctx context.Context, arg UpdateSubscriptionByIDForSubscriberParams) error {
	s.mu.Lock()
	s.UpdateSubscriptionByIDForSubscriberCalls = append(s.UpdateSubscriptionByIDForSubscriberCalls, arg)
	fn := s.UpdateSubscriptionByIDForSubscriberFn
	err := s.UpdateSubscriptionByIDForSubscriberErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) DeleteSubscriptionByIDForSubscriber(ctx context.Context, arg DeleteSubscriptionByIDForSubscriberParams) error {
	s.mu.Lock()
	s.DeleteSubscriptionByIDForSubscriberCalls = append(s.DeleteSubscriptionByIDForSubscriberCalls, arg)
	fn := s.DeleteSubscriptionByIDForSubscriberFn
	err := s.DeleteSubscriptionByIDForSubscriberErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}
