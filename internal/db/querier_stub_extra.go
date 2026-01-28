package db

import (
	"context"
	"database/sql"
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
