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

	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow   *GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingErr   error
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls []int32

	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow   *GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserErr   error
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls []GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams

	GetCommentsBySectionThreadIdForUserRows  []*GetCommentsBySectionThreadIdForUserRow
	GetCommentsBySectionThreadIdForUserErr   error
	GetCommentsBySectionThreadIdForUserCalls []GetCommentsBySectionThreadIdForUserParams
	GetThreadLastPosterAndPermsRow           *GetThreadLastPosterAndPermsRow
	GetThreadLastPosterAndPermsErr           error
	GetThreadLastPosterAndPermsCalls         []GetThreadLastPosterAndPermsParams
	AdminInsertQueuedLinkFromQueueReturn     int64
	AdminInsertQueuedLinkFromQueueErr        error
	AdminInsertQueuedLinkFromQueueCalls      []int32
	SystemCreateSearchWordReturnID           int64
	SystemCreateSearchWordErr                error
	SystemCreateSearchWordCalls              []string
	SystemCreateSearchWordFn                 func(string) (int64, error)
	SystemAddToLinkerSearchCalls             []SystemAddToLinkerSearchParams
	SystemAddToLinkerSearchErr               error
	SystemAddToLinkerSearchFn                func(SystemAddToLinkerSearchParams) error
	SystemSetLinkerLastIndexCalls            []int32
	SystemSetLinkerLastIndexErr              error

	SystemCheckGrantReturns int32
	SystemCheckGrantErr     error
	SystemCheckGrantCalls   []SystemCheckGrantParams
	SystemCheckGrantFn      func(SystemCheckGrantParams) (int32, error)

	SystemCheckRoleGrantReturns int32
	SystemCheckRoleGrantErr     error
	SystemCheckRoleGrantCalls   []SystemCheckRoleGrantParams
	SystemCheckRoleGrantFn      func(SystemCheckRoleGrantParams) (int32, error)

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

func (s *QuerierStub) GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(ctx context.Context, id int32) (*GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error) {
	s.mu.Lock()
	s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls = append(s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls, id)
	row := s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
	err := s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending not stubbed")
	}
	return row, nil
}

func (s *QuerierStub) GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(ctx context.Context, arg GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams) (*GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	s.mu.Lock()
	s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls = append(s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls, arg)
	row := s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
	err := s.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser not stubbed")
	}
	return row, nil
}

func (s *QuerierStub) GetCommentsBySectionThreadIdForUser(ctx context.Context, arg GetCommentsBySectionThreadIdForUserParams) ([]*GetCommentsBySectionThreadIdForUserRow, error) {
	s.mu.Lock()
	s.GetCommentsBySectionThreadIdForUserCalls = append(s.GetCommentsBySectionThreadIdForUserCalls, arg)
	rows := s.GetCommentsBySectionThreadIdForUserRows
	err := s.GetCommentsBySectionThreadIdForUserErr
	s.mu.Unlock()
	if rows == nil {
		rows = []*GetCommentsBySectionThreadIdForUserRow{}
	}
	return rows, err
}

func (s *QuerierStub) GetThreadLastPosterAndPerms(ctx context.Context, arg GetThreadLastPosterAndPermsParams) (*GetThreadLastPosterAndPermsRow, error) {
	s.mu.Lock()
	s.GetThreadLastPosterAndPermsCalls = append(s.GetThreadLastPosterAndPermsCalls, arg)
	row := s.GetThreadLastPosterAndPermsRow
	err := s.GetThreadLastPosterAndPermsErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("GetThreadLastPosterAndPerms not stubbed")
	}
	return row, nil
}

func (s *QuerierStub) AdminInsertQueuedLinkFromQueue(ctx context.Context, id int32) (int64, error) {
	s.mu.Lock()
	s.AdminInsertQueuedLinkFromQueueCalls = append(s.AdminInsertQueuedLinkFromQueueCalls, id)
	ret := s.AdminInsertQueuedLinkFromQueueReturn
	err := s.AdminInsertQueuedLinkFromQueueErr
	s.mu.Unlock()
	if err != nil {
		return 0, err
	}
	if ret == 0 {
		return 0, errors.New("AdminInsertQueuedLinkFromQueue not stubbed")
	}
	return ret, nil
}

func (s *QuerierStub) SystemCreateSearchWord(ctx context.Context, word string) (int64, error) {
	s.mu.Lock()
	s.SystemCreateSearchWordCalls = append(s.SystemCreateSearchWordCalls, word)
	fn := s.SystemCreateSearchWordFn
	ret := s.SystemCreateSearchWordReturnID
	err := s.SystemCreateSearchWordErr
	count := len(s.SystemCreateSearchWordCalls)
	s.mu.Unlock()
	if fn != nil {
		return fn(word)
	}
	if err != nil {
		return 0, err
	}
	if ret == 0 {
		ret = int64(count)
	}
	return ret, nil
}

func (s *QuerierStub) SystemAddToLinkerSearch(ctx context.Context, arg SystemAddToLinkerSearchParams) error {
	s.mu.Lock()
	s.SystemAddToLinkerSearchCalls = append(s.SystemAddToLinkerSearchCalls, arg)
	fn := s.SystemAddToLinkerSearchFn
	err := s.SystemAddToLinkerSearchErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return err
}

func (s *QuerierStub) SystemSetLinkerLastIndex(ctx context.Context, linkerID int32) error {
	s.mu.Lock()
	s.SystemSetLinkerLastIndexCalls = append(s.SystemSetLinkerLastIndexCalls, linkerID)
	err := s.SystemSetLinkerLastIndexErr
	s.mu.Unlock()
	return err
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
