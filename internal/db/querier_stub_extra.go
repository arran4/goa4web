package db

import (
	"context"
	"database/sql"
	"time"
)

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

func (s *QuerierStub) SystemGetUserByUsername(ctx context.Context, username sql.NullString) (*SystemGetUserByUsernameRow, error) {
	s.mu.Lock()
	s.SystemGetUserByUsernameCalls = append(s.SystemGetUserByUsernameCalls, username)
	fn := s.SystemGetUserByUsernameFn
	row := s.SystemGetUserByUsernameRow
	err := s.SystemGetUserByUsernameErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, username)
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
