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

func (s *QuerierStub) GetPendingPasswordByCode(ctx context.Context, verificationCode string) (*PendingPassword, error) {
	s.mu.Lock()
	s.GetPendingPasswordByCodeCalls = append(s.GetPendingPasswordByCodeCalls, verificationCode)
	fn := s.GetPendingPasswordByCodeFn
	ret := s.GetPendingPasswordByCodeReturns
	err := s.GetPendingPasswordByCodeErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, verificationCode)
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
