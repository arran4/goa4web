package db

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

// QuerierStub records calls for selective db.Querier methods in tests.
type QuerierStub struct {
	Querier
	mu sync.Mutex

	SystemGetUserByIDRow   *SystemGetUserByIDRow
	SystemGetUserByIDErr   error
	SystemGetUserByIDCalls []int32

	SystemGetUserByEmailRow   *SystemGetUserByEmailRow
	SystemGetUserByEmailErr   error
	SystemGetUserByEmailCalls []string

	SystemGetLastNotificationForRecipientByMessageRow   *Notification
	SystemGetLastNotificationForRecipientByMessageErr   error
	SystemGetLastNotificationForRecipientByMessageCalls []SystemGetLastNotificationForRecipientByMessageParams

	SystemCreateNotificationErr   error
	SystemCreateNotificationCalls []SystemCreateNotificationParams

	InsertPendingEmailErr   error
	InsertPendingEmailCalls []InsertPendingEmailParams

	AdminListAdministratorEmailsErr     error
	AdminListAdministratorEmailsReturns []string
	AdminListAdministratorEmailsCalls   int

	SystemGetTemplateOverrideReturns string
	SystemGetTemplateOverrideErr     error
	SystemGetTemplateOverrideCalls   []string

	ListSubscribersForPatternsParams []ListSubscribersForPatternsParams
	ListSubscribersForPatternsReturn map[string][]int32

	GetPreferenceForListerParams []int32
	GetPreferenceForListerReturn map[int32]*Preference

	InsertSubscriptionParams []InsertSubscriptionParams

	ListSubscribersForPatternParams []ListSubscribersForPatternParams
	ListSubscribersForPatternReturn map[string][]int32

	DeleteThreadsByTopicIDCalls []int32
	DeleteThreadsByTopicIDErr   error

	SystemCheckGrantReturns int32
	SystemCheckGrantErr     error
	SystemCheckGrantCalls   []SystemCheckGrantParams
	SystemCheckGrantFn      func(SystemCheckGrantParams) (int32, error)

	SystemCheckRoleGrantReturns int32
	SystemCheckRoleGrantErr     error
	SystemCheckRoleGrantCalls   []SystemCheckRoleGrantParams
	SystemCheckRoleGrantFn      func(SystemCheckRoleGrantParams) (int32, error)

	SystemGetUserByUsernameRow   *SystemGetUserByUsernameRow
	SystemGetUserByUsernameErr   error
	SystemGetUserByUsernameCalls []sql.NullString

	SystemCreateUserRoleErr   error
	SystemCreateUserRoleCalls []SystemCreateUserRoleParams

	AdminUpdateUsernameByIDErr   error
	AdminUpdateUsernameByIDCalls []AdminUpdateUsernameByIDParams

	AdminListForumTopicGrantsByTopicIDCalls   []sql.NullInt32
	AdminListForumTopicGrantsByTopicIDReturns []*AdminListForumTopicGrantsByTopicIDRow
	AdminListForumTopicGrantsByTopicIDErr     error

	AdminListPrivateTopicParticipantsByTopicIDCalls   []sql.NullInt32
	AdminListPrivateTopicParticipantsByTopicIDReturns []*AdminListPrivateTopicParticipantsByTopicIDRow
	AdminListPrivateTopicParticipantsByTopicIDErr     error
}

func (s *QuerierStub) AdminListPrivateTopicParticipantsByTopicID(ctx context.Context, itemID sql.NullInt32) ([]*AdminListPrivateTopicParticipantsByTopicIDRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListPrivateTopicParticipantsByTopicIDCalls = append(s.AdminListPrivateTopicParticipantsByTopicIDCalls, itemID)
	return s.AdminListPrivateTopicParticipantsByTopicIDReturns, s.AdminListPrivateTopicParticipantsByTopicIDErr
}

func (s *QuerierStub) AdminListForumTopicGrantsByTopicID(ctx context.Context, itemID sql.NullInt32) ([]*AdminListForumTopicGrantsByTopicIDRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListForumTopicGrantsByTopicIDCalls = append(s.AdminListForumTopicGrantsByTopicIDCalls, itemID)
	return s.AdminListForumTopicGrantsByTopicIDReturns, s.AdminListForumTopicGrantsByTopicIDErr
}

func (s *QuerierStub) DeleteThreadsByTopicID(ctx context.Context, forumtopicIdforumtopic int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DeleteThreadsByTopicIDCalls = append(s.DeleteThreadsByTopicIDCalls, forumtopicIdforumtopic)
	return s.DeleteThreadsByTopicIDErr
}

// SystemCheckGrant records the call and returns the configured response.
func (s *QuerierStub) SystemCheckGrant(ctx context.Context, arg SystemCheckGrantParams) (int32, error) {
	s.mu.Lock()
	s.SystemCheckGrantCalls = append(s.SystemCheckGrantCalls, arg)
	fn := s.SystemCheckGrantFn
	ret := s.SystemCheckGrantReturns
	err := s.SystemCheckGrantErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	if err != nil {
		return 0, err
	}
	if ret == 0 {
		ret = 1
	}
	return ret, nil
}

// SystemCheckRoleGrant records the call and returns the configured response.
func (s *QuerierStub) SystemCheckRoleGrant(ctx context.Context, arg SystemCheckRoleGrantParams) (int32, error) {
	s.mu.Lock()
	s.SystemCheckRoleGrantCalls = append(s.SystemCheckRoleGrantCalls, arg)
	fn := s.SystemCheckRoleGrantFn
	ret := s.SystemCheckRoleGrantReturns
	err := s.SystemCheckRoleGrantErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	if err != nil {
		return 0, err
	}
	if ret == 0 {
		ret = 1
	}
	return ret, nil
}

// SystemGetUserByUsername records the call and returns the configured response.
func (s *QuerierStub) SystemGetUserByUsername(ctx context.Context, username sql.NullString) (*SystemGetUserByUsernameRow, error) {
	s.mu.Lock()
	s.SystemGetUserByUsernameCalls = append(s.SystemGetUserByUsernameCalls, username)
	s.mu.Unlock()
	if s.SystemGetUserByUsernameErr != nil {
		return nil, s.SystemGetUserByUsernameErr
	}
	if s.SystemGetUserByUsernameRow == nil {
		return nil, errors.New("SystemGetUserByUsername not stubbed")
	}
	return s.SystemGetUserByUsernameRow, nil
}

// SystemCreateUserRole records the call and returns the configured response.
func (s *QuerierStub) SystemCreateUserRole(ctx context.Context, arg SystemCreateUserRoleParams) error {
	s.mu.Lock()
	s.SystemCreateUserRoleCalls = append(s.SystemCreateUserRoleCalls, arg)
	s.mu.Unlock()
	return s.SystemCreateUserRoleErr
}

// AdminUpdateUsernameByID records the call and returns the configured response.
func (s *QuerierStub) AdminUpdateUsernameByID(ctx context.Context, arg AdminUpdateUsernameByIDParams) error {
	s.mu.Lock()
	s.AdminUpdateUsernameByIDCalls = append(s.AdminUpdateUsernameByIDCalls, arg)
	s.mu.Unlock()
	return s.AdminUpdateUsernameByIDErr
}

// SystemGetUserByID records the call and returns the configured response.
func (s *QuerierStub) SystemGetUserByID(ctx context.Context, idusers int32) (*SystemGetUserByIDRow, error) {
	s.mu.Lock()
	s.SystemGetUserByIDCalls = append(s.SystemGetUserByIDCalls, idusers)
	s.mu.Unlock()
	if s.SystemGetUserByIDErr != nil {
		return nil, s.SystemGetUserByIDErr
	}
	if s.SystemGetUserByIDRow == nil {
		return nil, errors.New("SystemGetUserByID not stubbed")
	}
	return s.SystemGetUserByIDRow, nil
}

// SystemGetUserByEmail records the call and returns the configured response.
func (s *QuerierStub) SystemGetUserByEmail(ctx context.Context, email string) (*SystemGetUserByEmailRow, error) {
	s.mu.Lock()
	s.SystemGetUserByEmailCalls = append(s.SystemGetUserByEmailCalls, email)
	s.mu.Unlock()
	if s.SystemGetUserByEmailErr != nil {
		return nil, s.SystemGetUserByEmailErr
	}
	if s.SystemGetUserByEmailRow == nil {
		return nil, errors.New("SystemGetUserByEmail not stubbed")
	}
	return s.SystemGetUserByEmailRow, nil
}

// SystemGetLastNotificationForRecipientByMessage records the call and returns the configured response.
func (s *QuerierStub) SystemGetLastNotificationForRecipientByMessage(ctx context.Context, arg SystemGetLastNotificationForRecipientByMessageParams) (*Notification, error) {
	s.mu.Lock()
	s.SystemGetLastNotificationForRecipientByMessageCalls = append(s.SystemGetLastNotificationForRecipientByMessageCalls, arg)
	s.mu.Unlock()
	if s.SystemGetLastNotificationForRecipientByMessageErr != nil {
		return nil, s.SystemGetLastNotificationForRecipientByMessageErr
	}
	if s.SystemGetLastNotificationForRecipientByMessageRow == nil {
		return nil, errors.New("SystemGetLastNotificationForRecipientByMessage not stubbed")
	}
	return s.SystemGetLastNotificationForRecipientByMessageRow, nil
}

// SystemCreateNotification records the call and returns the configured response.
func (s *QuerierStub) SystemCreateNotification(ctx context.Context, arg SystemCreateNotificationParams) error {
	s.mu.Lock()
	s.SystemCreateNotificationCalls = append(s.SystemCreateNotificationCalls, arg)
	s.mu.Unlock()
	return s.SystemCreateNotificationErr
}

// InsertPendingEmail records the call and returns the configured response.
func (s *QuerierStub) InsertPendingEmail(ctx context.Context, arg InsertPendingEmailParams) error {
	s.mu.Lock()
	s.InsertPendingEmailCalls = append(s.InsertPendingEmailCalls, arg)
	s.mu.Unlock()
	return s.InsertPendingEmailErr
}

// AdminListAdministratorEmails records the call and returns the configured response.
func (s *QuerierStub) AdminListAdministratorEmails(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	s.AdminListAdministratorEmailsCalls++
	s.mu.Unlock()
	return s.AdminListAdministratorEmailsReturns, s.AdminListAdministratorEmailsErr
}

// SystemGetTemplateOverride records the call and returns the configured response.
func (s *QuerierStub) SystemGetTemplateOverride(ctx context.Context, name string) (string, error) {
	s.mu.Lock()
	s.SystemGetTemplateOverrideCalls = append(s.SystemGetTemplateOverrideCalls, name)
	s.mu.Unlock()
	return s.SystemGetTemplateOverrideReturns, s.SystemGetTemplateOverrideErr
}

func (s *QuerierStub) ListSubscribersForPatterns(ctx context.Context, arg ListSubscribersForPatternsParams) ([]int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListSubscribersForPatternsParams = append(s.ListSubscribersForPatternsParams, arg)
	// Flatten returns for all patterns or just return a default set
	var ret []int32
	if s.ListSubscribersForPatternsReturn != nil {
		for _, p := range arg.Patterns {
			ret = append(ret, s.ListSubscribersForPatternsReturn[p]...)
		}
	}
	return ret, nil
}

func (s *QuerierStub) GetPreferenceForLister(ctx context.Context, listerID int32) (*Preference, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetPreferenceForListerParams = append(s.GetPreferenceForListerParams, listerID)
	if s.GetPreferenceForListerReturn != nil {
		if v, ok := s.GetPreferenceForListerReturn[listerID]; ok {
			return v, nil
		}
	}
	return nil, nil
}

func (s *QuerierStub) InsertSubscription(ctx context.Context, arg InsertSubscriptionParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InsertSubscriptionParams = append(s.InsertSubscriptionParams, arg)
	return nil
}

func (s *QuerierStub) ListSubscribersForPattern(ctx context.Context, arg ListSubscribersForPatternParams) ([]int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListSubscribersForPatternParams = append(s.ListSubscribersForPatternParams, arg)
	if s.ListSubscribersForPatternReturn != nil {
		if v, ok := s.ListSubscribersForPatternReturn[arg.Pattern]; ok {
			return v, nil
		}
	}
	return nil, nil
}
