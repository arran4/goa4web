package db

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"sync"
)

// FakeSQLResult implements sql.Result for tests.
type FakeSQLResult struct {
	LastInsertIDValue int64
	RowsAffectedValue int64
}

func (r FakeSQLResult) LastInsertId() (int64, error) {
	return r.LastInsertIDValue, nil
}

func (r FakeSQLResult) RowsAffected() (int64, error) {
	return r.RowsAffectedValue, nil
}

// QuerierStub records calls for selective db.Querier methods in tests.
type QuerierStub struct {
	Querier
	mu sync.Mutex

	ListActiveBansReturns []*BannedIp
	ListActiveBansErr     error
	ListActiveBansCalls   int
	ContentLabelStatus                  map[string]map[int32]map[string]struct{}
	ContentPrivateLabels                map[string]map[int32]map[int32]map[string]bool
	ContentPublicLabels                 map[string]map[int32]map[string]struct{}
	AddContentLabelStatusErr            error
	AddContentLabelStatusCalls          []AddContentLabelStatusParams
	AddContentPrivateLabelErr           error
	AddContentPrivateLabelCalls         []AddContentPrivateLabelParams
	AddContentPublicLabelErr            error
	AddContentPublicLabelCalls          []AddContentPublicLabelParams
	ListContentLabelStatusErr           error
	ListContentLabelStatusCalls         []ListContentLabelStatusParams
	ListContentPrivateLabelsErr         error
	ListContentPrivateLabelsCalls       []ListContentPrivateLabelsParams
	ListContentPublicLabelsErr          error
	ListContentPublicLabelsCalls        []ListContentPublicLabelsParams
	RemoveContentLabelStatusErr         error
	RemoveContentLabelStatusCalls       []RemoveContentLabelStatusParams
	RemoveContentPrivateLabelErr        error
	RemoveContentPrivateLabelCalls      []RemoveContentPrivateLabelParams
	RemoveContentPublicLabelErr         error
	RemoveContentPublicLabelCalls       []RemoveContentPublicLabelParams
	SystemClearContentPrivateLabelErr   error
	SystemClearContentPrivateLabelCalls []SystemClearContentPrivateLabelParams

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

	ListContentPrivateLabelsReturns []*ListContentPrivateLabelsRow
	ListContentPrivateLabelsErr     error
	ListContentPrivateLabelsCalls   []ListContentPrivateLabelsParams
	ListContentPrivateLabelsFn      func(ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error)

	ListSubscriptionsByUserReturns []*ListSubscriptionsByUserRow
	ListSubscriptionsByUserErr     error
	ListSubscriptionsByUserCalls   []int32

	SystemGetTemplateOverrideReturns string
	SystemGetTemplateOverrideErr     error
	SystemGetTemplateOverrideCalls   []string
	SystemGetTemplateOverrideSeq     []string
	systemGetTemplateOverrideIdx     int

	AdminSetTemplateOverrideCalls []AdminSetTemplateOverrideParams
	AdminSetTemplateOverrideErr   error

	ListSubscribersForPatternsParams []ListSubscribersForPatternsParams
	ListSubscribersForPatternsReturn map[string][]int32

	GetPreferenceForListerParams []int32
	GetPreferenceForListerReturn map[int32]*Preference

	InsertSubscriptionParams         []InsertSubscriptionParams
	InsertFAQQuestionForWriterCalls  []InsertFAQQuestionForWriterParams
	InsertFAQQuestionForWriterResult sql.Result
	InsertFAQQuestionForWriterErr    error

	ListSubscribersForPatternParams []ListSubscribersForPatternParams
	ListSubscribersForPatternReturn map[string][]int32

	GetCommentByIdForUserCalls []GetCommentByIdForUserParams
	GetCommentByIdForUserRow   *GetCommentByIdForUserRow
	GetCommentByIdForUserErr   error

	DeleteThreadsByTopicIDCalls []int32
	DeleteThreadsByTopicIDErr   error

	SystemCheckGrantReturns int32
	SystemCheckGrantErr     error
	SystemCheckGrantCalls   []SystemCheckGrantParams
	SystemCheckGrantFn      func(SystemCheckGrantParams) (int32, error)

	GetWritingForListerByIDRow   *GetWritingForListerByIDRow
	GetWritingForListerByIDErr   error
	GetWritingForListerByIDCalls []GetWritingForListerByIDParams

	SystemCheckRoleGrantReturns int32
	SystemCheckRoleGrantErr     error
	SystemCheckRoleGrantCalls   []SystemCheckRoleGrantParams
	SystemCheckRoleGrantFn      func(SystemCheckRoleGrantParams) (int32, error)

	GetPermissionsByUserIDReturns []*GetPermissionsByUserIDRow
	GetPermissionsByUserIDErr     error
	GetPermissionsByUserIDCalls   []int32
	GetPermissionsByUserIDFn      func(int32) ([]*GetPermissionsByUserIDRow, error)

	AdminListForumTopicGrantsByTopicIDCalls   []sql.NullInt32
	AdminListForumTopicGrantsByTopicIDReturns []*AdminListForumTopicGrantsByTopicIDRow
	AdminListForumTopicGrantsByTopicIDErr     error

	AdminListPrivateTopicParticipantsByTopicIDCalls   []sql.NullInt32
	AdminListPrivateTopicParticipantsByTopicIDReturns []*AdminListPrivateTopicParticipantsByTopicIDRow
	AdminListPrivateTopicParticipantsByTopicIDErr     error

	ListWritersForListerCalls   []ListWritersForListerParams
	ListWritersForListerReturns []*ListWritersForListerRow
	ListWritersForListerErr     error
	ListWritersForListerFn      func(ListWritersForListerParams) ([]*ListWritersForListerRow, error)

	ListWritersSearchForListerCalls   []ListWritersSearchForListerParams
	ListWritersSearchForListerReturns []*ListWritersSearchForListerRow
	ListWritersSearchForListerErr     error
	ListWritersSearchForListerFn      func(ListWritersSearchForListerParams) ([]*ListWritersSearchForListerRow, error)
}

func (s *QuerierStub) ensurePublicLabelSetLocked(item string, itemID int32) map[string]struct{} {
	if s.ContentPublicLabels == nil {
		s.ContentPublicLabels = make(map[string]map[int32]map[string]struct{})
	}
	itemMap, ok := s.ContentPublicLabels[item]
	if !ok {
		itemMap = make(map[int32]map[string]struct{})
		s.ContentPublicLabels[item] = itemMap
	}
	labels, ok := itemMap[itemID]
	if !ok {
		labels = make(map[string]struct{})
		itemMap[itemID] = labels
	}
	return labels
}

func (s *QuerierStub) publicLabelSet(item string, itemID int32) map[string]struct{} {
	itemMap := s.ContentPublicLabels[item]
	if itemMap == nil {
		return nil
	}
	return itemMap[itemID]
}

func (s *QuerierStub) ensureAuthorLabelSetLocked(item string, itemID int32) map[string]struct{} {
	if s.ContentLabelStatus == nil {
		s.ContentLabelStatus = make(map[string]map[int32]map[string]struct{})
	}
	itemMap, ok := s.ContentLabelStatus[item]
	if !ok {
		itemMap = make(map[int32]map[string]struct{})
		s.ContentLabelStatus[item] = itemMap
	}
	labels, ok := itemMap[itemID]
	if !ok {
		labels = make(map[string]struct{})
		itemMap[itemID] = labels
	}
	return labels
}

func (s *QuerierStub) authorLabelSet(item string, itemID int32) map[string]struct{} {
	itemMap := s.ContentLabelStatus[item]
	if itemMap == nil {
		return nil
	}
	return itemMap[itemID]
}

func (s *QuerierStub) ensurePrivateLabelSetLocked(item string, itemID int32, userID int32) map[string]bool {
	if s.ContentPrivateLabels == nil {
		s.ContentPrivateLabels = make(map[string]map[int32]map[int32]map[string]bool)
	}
	itemMap, ok := s.ContentPrivateLabels[item]
	if !ok {
		itemMap = make(map[int32]map[int32]map[string]bool)
		s.ContentPrivateLabels[item] = itemMap
	}
	userMap, ok := itemMap[itemID]
	if !ok {
		userMap = make(map[int32]map[string]bool)
		itemMap[itemID] = userMap
	}
	labels, ok := userMap[userID]
	if !ok {
		labels = make(map[string]bool)
		userMap[userID] = labels
	}
	return labels
}

func (s *QuerierStub) privateLabelSet(item string, itemID int32, userID int32) map[string]bool {
	itemMap := s.ContentPrivateLabels[item]
	if itemMap == nil {
		return nil
	}
	userMap := itemMap[itemID]
	if userMap == nil {
		return nil
	}
	return userMap[userID]
}

// AddContentPublicLabel records the call and stores the label for later retrieval.
func (s *QuerierStub) AddContentPublicLabel(ctx context.Context, arg AddContentPublicLabelParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AddContentPublicLabelCalls = append(s.AddContentPublicLabelCalls, arg)
	if s.AddContentPublicLabelErr != nil {
		return s.AddContentPublicLabelErr
	}
	labels := s.ensurePublicLabelSetLocked(arg.Item, arg.ItemID)
	labels[arg.Label] = struct{}{}
	return nil
}

// ListContentPublicLabels records the call and returns stored public labels.
func (s *QuerierStub) ListContentPublicLabels(ctx context.Context, arg ListContentPublicLabelsParams) ([]*ListContentPublicLabelsRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListContentPublicLabelsCalls = append(s.ListContentPublicLabelsCalls, arg)
	if s.ListContentPublicLabelsErr != nil {
		return nil, s.ListContentPublicLabelsErr
	}
	var res []*ListContentPublicLabelsRow
	labels := s.publicLabelSet(arg.Item, arg.ItemID)
	if len(labels) == 0 {
		return res, nil
	}
	names := make([]string, 0, len(labels))
	for label := range labels {
		names = append(names, label)
	}
	sort.Strings(names)
	for _, l := range names {
		res = append(res, &ListContentPublicLabelsRow{Item: arg.Item, ItemID: arg.ItemID, Label: l})
	}
	return res, nil
}

// RemoveContentPublicLabel records the call and removes the label from the store.
func (s *QuerierStub) RemoveContentPublicLabel(ctx context.Context, arg RemoveContentPublicLabelParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemoveContentPublicLabelCalls = append(s.RemoveContentPublicLabelCalls, arg)
	if s.RemoveContentPublicLabelErr != nil {
		return s.RemoveContentPublicLabelErr
	}
	if itemMap := s.ContentPublicLabels[arg.Item]; itemMap != nil {
		if labels := itemMap[arg.ItemID]; labels != nil {
			delete(labels, arg.Label)
			if len(labels) == 0 {
				delete(itemMap, arg.ItemID)
			}
		}
		if len(itemMap) == 0 {
			delete(s.ContentPublicLabels, arg.Item)
		}
	}
	return nil
}

// AddContentLabelStatus records the call and stores the author label for later retrieval.
func (s *QuerierStub) AddContentLabelStatus(ctx context.Context, arg AddContentLabelStatusParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AddContentLabelStatusCalls = append(s.AddContentLabelStatusCalls, arg)
	if s.AddContentLabelStatusErr != nil {
		return s.AddContentLabelStatusErr
	}
	labels := s.ensureAuthorLabelSetLocked(arg.Item, arg.ItemID)
	labels[arg.Label] = struct{}{}
	return nil
}

// ListContentLabelStatus records the call and returns stored author labels.
func (s *QuerierStub) ListContentLabelStatus(ctx context.Context, arg ListContentLabelStatusParams) ([]*ListContentLabelStatusRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListContentLabelStatusCalls = append(s.ListContentLabelStatusCalls, arg)
	if s.ListContentLabelStatusErr != nil {
		return nil, s.ListContentLabelStatusErr
	}
	var res []*ListContentLabelStatusRow
	labels := s.authorLabelSet(arg.Item, arg.ItemID)
	if len(labels) == 0 {
		return res, nil
	}
	names := make([]string, 0, len(labels))
	for label := range labels {
		names = append(names, label)
	}
	sort.Strings(names)
	for _, l := range names {
		res = append(res, &ListContentLabelStatusRow{Item: arg.Item, ItemID: arg.ItemID, Label: l})
	}
	return res, nil
}

// RemoveContentLabelStatus records the call and removes the author label from the store.
func (s *QuerierStub) RemoveContentLabelStatus(ctx context.Context, arg RemoveContentLabelStatusParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemoveContentLabelStatusCalls = append(s.RemoveContentLabelStatusCalls, arg)
	if s.RemoveContentLabelStatusErr != nil {
		return s.RemoveContentLabelStatusErr
	}
	if itemMap := s.ContentLabelStatus[arg.Item]; itemMap != nil {
		if labels := itemMap[arg.ItemID]; labels != nil {
			delete(labels, arg.Label)
			if len(labels) == 0 {
				delete(itemMap, arg.ItemID)
			}
		}
		if len(itemMap) == 0 {
			delete(s.ContentLabelStatus, arg.Item)
		}
	}
	return nil
}

// AddContentPrivateLabel records the call and stores the private label for later retrieval.
func (s *QuerierStub) AddContentPrivateLabel(ctx context.Context, arg AddContentPrivateLabelParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AddContentPrivateLabelCalls = append(s.AddContentPrivateLabelCalls, arg)
	if s.AddContentPrivateLabelErr != nil {
		return s.AddContentPrivateLabelErr
	}
	labels := s.ensurePrivateLabelSetLocked(arg.Item, arg.ItemID, arg.UserID)
	labels[arg.Label] = arg.Invert
	return nil
}

// ListContentPrivateLabels records the call and returns stored private labels.
func (s *QuerierStub) ListContentPrivateLabels(ctx context.Context, arg ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListContentPrivateLabelsCalls = append(s.ListContentPrivateLabelsCalls, arg)
	if s.ListContentPrivateLabelsErr != nil {
		return nil, s.ListContentPrivateLabelsErr
	}
	var res []*ListContentPrivateLabelsRow
	labels := s.privateLabelSet(arg.Item, arg.ItemID, arg.UserID)
	if len(labels) == 0 {
		return res, nil
	}
	names := make([]string, 0, len(labels))
	for label := range labels {
		names = append(names, label)
	}
	sort.Strings(names)
	for _, l := range names {
		res = append(res, &ListContentPrivateLabelsRow{Item: arg.Item, ItemID: arg.ItemID, UserID: arg.UserID, Label: l, Invert: labels[l]})
	}
	return res, nil
}

// RemoveContentPrivateLabel records the call and removes the private label from the store.
func (s *QuerierStub) RemoveContentPrivateLabel(ctx context.Context, arg RemoveContentPrivateLabelParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemoveContentPrivateLabelCalls = append(s.RemoveContentPrivateLabelCalls, arg)
	if s.RemoveContentPrivateLabelErr != nil {
		return s.RemoveContentPrivateLabelErr
	}
	if itemMap := s.ContentPrivateLabels[arg.Item]; itemMap != nil {
		if userMap := itemMap[arg.ItemID]; userMap != nil {
			if labels := userMap[arg.UserID]; labels != nil {
				delete(labels, arg.Label)
				if len(labels) == 0 {
					delete(userMap, arg.UserID)
				}
			}
			if len(userMap) == 0 {
				delete(itemMap, arg.ItemID)
			}
		}
		if len(itemMap) == 0 {
			delete(s.ContentPrivateLabels, arg.Item)
		}
	}
	return nil
}

// SystemClearContentPrivateLabel records the call and removes the label for all users.
func (s *QuerierStub) SystemClearContentPrivateLabel(ctx context.Context, arg SystemClearContentPrivateLabelParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SystemClearContentPrivateLabelCalls = append(s.SystemClearContentPrivateLabelCalls, arg)
	if s.SystemClearContentPrivateLabelErr != nil {
		return s.SystemClearContentPrivateLabelErr
	}
	if itemMap := s.ContentPrivateLabels[arg.Item]; itemMap != nil {
		if userMap := itemMap[arg.ItemID]; userMap != nil {
			for uid, labels := range userMap {
				delete(labels, arg.Label)
				if len(labels) == 0 {
					delete(userMap, uid)
				}
			}
			if len(userMap) == 0 {
				delete(itemMap, arg.ItemID)
			}
		}
		if len(itemMap) == 0 {
			delete(s.ContentPrivateLabels, arg.Item)
		}
	}
	return nil
}

func (s *QuerierStub) ListActiveBans(ctx context.Context) ([]*BannedIp, error) {
	s.mu.Lock()
	s.ListActiveBansCalls++
	rows := s.ListActiveBansReturns
	err := s.ListActiveBansErr
	s.mu.Unlock()
	if rows == nil {
		return []*BannedIp{}, err
	}
	return rows, err
}

func (s *QuerierStub) AdminListPrivateTopicParticipantsByTopicID(ctx context.Context, itemID sql.NullInt32) ([]*AdminListPrivateTopicParticipantsByTopicIDRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListPrivateTopicParticipantsByTopicIDCalls = append(s.AdminListPrivateTopicParticipantsByTopicIDCalls, itemID)
	return s.AdminListPrivateTopicParticipantsByTopicIDReturns, s.AdminListPrivateTopicParticipantsByTopicIDErr
}

func (s *QuerierStub) ListWritersForLister(ctx context.Context, arg ListWritersForListerParams) ([]*ListWritersForListerRow, error) {
	s.mu.Lock()
	s.ListWritersForListerCalls = append(s.ListWritersForListerCalls, arg)
	fn := s.ListWritersForListerFn
	ret := s.ListWritersForListerReturns
	err := s.ListWritersForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

func (s *QuerierStub) ListWritersSearchForLister(ctx context.Context, arg ListWritersSearchForListerParams) ([]*ListWritersSearchForListerRow, error) {
	s.mu.Lock()
	s.ListWritersSearchForListerCalls = append(s.ListWritersSearchForListerCalls, arg)
	fn := s.ListWritersSearchForListerFn
	ret := s.ListWritersSearchForListerReturns
	err := s.ListWritersSearchForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
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

// GetCommentByIdForUser records the call and returns the configured response.
func (s *QuerierStub) GetCommentByIdForUser(ctx context.Context, arg GetCommentByIdForUserParams) (*GetCommentByIdForUserRow, error) {
	s.mu.Lock()
	s.GetCommentByIdForUserCalls = append(s.GetCommentByIdForUserCalls, arg)
	ret := s.GetCommentByIdForUserRow
	err := s.GetCommentByIdForUserErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return ret, nil
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
	if ret != 0 {
		return ret, nil
	}
	return 0, sql.ErrNoRows
}

// GetWritingForListerByID records the call and returns the configured response.
func (s *QuerierStub) GetWritingForListerByID(ctx context.Context, arg GetWritingForListerByIDParams) (*GetWritingForListerByIDRow, error) {
	s.mu.Lock()
	s.GetWritingForListerByIDCalls = append(s.GetWritingForListerByIDCalls, arg)
	row := s.GetWritingForListerByIDRow
	err := s.GetWritingForListerByIDErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("GetWritingForListerByID not stubbed")
	}
	return row, nil
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
	if ret != 0 {
		return ret, nil
	}
	return 0, sql.ErrNoRows
}

// GetPermissionsByUserID records the call and returns the configured response.
func (s *QuerierStub) GetPermissionsByUserID(ctx context.Context, idusers int32) ([]*GetPermissionsByUserIDRow, error) {
	s.mu.Lock()
	s.GetPermissionsByUserIDCalls = append(s.GetPermissionsByUserIDCalls, idusers)
	fn := s.GetPermissionsByUserIDFn
	ret := s.GetPermissionsByUserIDReturns
	err := s.GetPermissionsByUserIDErr
	s.mu.Unlock()
	if fn != nil {
		return fn(idusers)
	}
	return ret, err
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

// ListContentPrivateLabels records the call and returns the configured response or a custom function.
func (s *QuerierStub) ListContentPrivateLabels(ctx context.Context, arg ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error) {
	s.mu.Lock()
	s.ListContentPrivateLabelsCalls = append(s.ListContentPrivateLabelsCalls, arg)
	fn := s.ListContentPrivateLabelsFn
	ret := s.ListContentPrivateLabelsReturns
	err := s.ListContentPrivateLabelsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

// ListSubscriptionsByUser records the call and returns configured rows.
func (s *QuerierStub) ListSubscriptionsByUser(ctx context.Context, userID int32) ([]*ListSubscriptionsByUserRow, error) {
	s.mu.Lock()
	s.ListSubscriptionsByUserCalls = append(s.ListSubscriptionsByUserCalls, userID)
	ret := s.ListSubscriptionsByUserReturns
	err := s.ListSubscriptionsByUserErr
	s.mu.Unlock()
	return ret, err
}

// SystemGetTemplateOverride records the call and returns the configured response.
func (s *QuerierStub) SystemGetTemplateOverride(ctx context.Context, name string) (string, error) {
	s.mu.Lock()
	s.SystemGetTemplateOverrideCalls = append(s.SystemGetTemplateOverrideCalls, name)
	idx := s.systemGetTemplateOverrideIdx
	var body string
	if len(s.SystemGetTemplateOverrideSeq) > 0 {
		if idx >= len(s.SystemGetTemplateOverrideSeq) {
			idx = len(s.SystemGetTemplateOverrideSeq) - 1
		}
		body = s.SystemGetTemplateOverrideSeq[idx]
		if s.systemGetTemplateOverrideIdx < len(s.SystemGetTemplateOverrideSeq)-1 {
			s.systemGetTemplateOverrideIdx++
		}
	} else {
		body = s.SystemGetTemplateOverrideReturns
	}
	err := s.SystemGetTemplateOverrideErr
	s.mu.Unlock()
	return body, err
}

// AdminSetTemplateOverride records the call and returns the configured response.
func (s *QuerierStub) AdminSetTemplateOverride(ctx context.Context, arg AdminSetTemplateOverrideParams) error {
	s.mu.Lock()
	s.AdminSetTemplateOverrideCalls = append(s.AdminSetTemplateOverrideCalls, arg)
	err := s.AdminSetTemplateOverrideErr
	s.mu.Unlock()
	return err
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

// InsertFAQQuestionForWriter records the call and returns the configured sql.Result.
func (s *QuerierStub) InsertFAQQuestionForWriter(ctx context.Context, arg InsertFAQQuestionForWriterParams) (sql.Result, error) {
	s.mu.Lock()
	s.InsertFAQQuestionForWriterCalls = append(s.InsertFAQQuestionForWriterCalls, arg)
	res := s.InsertFAQQuestionForWriterResult
	err := s.InsertFAQQuestionForWriterErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if res == nil {
		return FakeSQLResult{}, nil
	}
	return res, nil
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
