package db

import (
	"context"
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
