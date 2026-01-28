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

	ListActiveBansReturns               []*BannedIp
	ListActiveBansErr                   error
	ListActiveBansCalls                 int
	ContentLabelStatus                  map[string]map[int32]map[string]struct{}
	ContentPrivateLabels                map[string]map[int32]map[int32]map[string]bool
	ContentPublicLabels                 map[string]map[int32]map[string]struct{}
	AddContentLabelStatusErr            error
	AddContentLabelStatusCalls          []AddContentLabelStatusParams
	AddContentPrivateLabelErr           error
	AddContentPrivateLabelCalls         []AddContentPrivateLabelParams
	AddContentPrivateLabelFn            func(context.Context, AddContentPrivateLabelParams) error
	AddContentPrivateLabelIgnoreLabels  map[string]bool
	AddContentPublicLabelErr            error
	AddContentPublicLabelCalls          []AddContentPublicLabelParams
	ListContentLabelStatusErr           error
	ListContentLabelStatusCalls         []ListContentLabelStatusParams
	ListContentLabelStatusReturns       []*ListContentLabelStatusRow
	ListContentPublicLabelsErr          error
	ListContentPublicLabelsCalls        []ListContentPublicLabelsParams
	ListContentPublicLabelsReturns      []*ListContentPublicLabelsRow
	ListContentPublicLabelsFn           func(ListContentPublicLabelsParams) ([]*ListContentPublicLabelsRow, error)
	RemoveContentLabelStatusErr         error
	RemoveContentLabelStatusCalls       []RemoveContentLabelStatusParams
	RemoveContentPrivateLabelErr        error
	RemoveContentPrivateLabelCalls      []RemoveContentPrivateLabelParams
	RemoveContentPublicLabelErr         error
	RemoveContentPublicLabelCalls       []RemoveContentPublicLabelParams
	SystemClearContentPrivateLabelErr   error
	SystemClearContentPrivateLabelCalls []SystemClearContentPrivateLabelParams

	ClearUnreadContentPrivateLabelExceptUserErr   error
	ClearUnreadContentPrivateLabelExceptUserCalls []ClearUnreadContentPrivateLabelExceptUserParams
	ClearUnreadContentPrivateLabelExceptUserFn    func(context.Context, ClearUnreadContentPrivateLabelExceptUserParams) error

	SystemGetUserByIDRow   *SystemGetUserByIDRow
	SystemGetUserByIDErr   error
	SystemGetUserByIDCalls []int32

	SystemGetUserByEmailRow   *SystemGetUserByEmailRow
	SystemGetUserByEmailErr   error
	SystemGetUserByEmailCalls []string
	SystemGetUserByEmailFn    func(context.Context, string) (*SystemGetUserByEmailRow, error)

	SystemGetUserByUsernameRow   *SystemGetUserByUsernameRow
	SystemGetUserByUsernameErr   error
	SystemGetUserByUsernameCalls []sql.NullString
	SystemGetUserByUsernameFn    func(context.Context, sql.NullString) (*SystemGetUserByUsernameRow, error)

	SystemGetLastNotificationForRecipientByMessageRow   *Notification
	SystemGetLastNotificationForRecipientByMessageErr   error
	SystemGetLastNotificationForRecipientByMessageCalls []SystemGetLastNotificationForRecipientByMessageParams

	SystemCreateNotificationErr   error
	SystemCreateNotificationCalls []SystemCreateNotificationParams

	SystemCreateThreadCalls   []int32
	SystemCreateThreadReturns int64
	SystemCreateThreadErr     error
	SystemCreateThreadFn      func(context.Context, int32) (int64, error)

	SystemInsertDeadLetterCalls int

	InsertPendingEmailErr   error
	InsertPendingEmailCalls []InsertPendingEmailParams

	AdminListAdministratorEmailsErr     error
	AdminListAdministratorEmailsReturns []string
	AdminListAdministratorEmailsCalls   int

	AdminListUserEmailsReturns []*UserEmail
	AdminListUserEmailsErr     error
	AdminListUserEmailsCalls   []int32

	AdminGetImagePostRow   *AdminGetImagePostRow
	AdminGetImagePostErr   error
	AdminGetImagePostCalls []int32
	AdminGetImagePostFn    func(context.Context, int32) (*AdminGetImagePostRow, error)

	AdminApproveImagePostCalls []int32
	AdminApproveImagePostErr   error
	AdminApproveImagePostFn    func(context.Context, int32) error

	AdminUserPostCountsByIDReturns *AdminUserPostCountsByIDRow
	AdminUserPostCountsByIDErr     error
	AdminUserPostCountsByIDCalls   []int32

	GetBookmarksForUserReturns *GetBookmarksForUserRow
	GetBookmarksForUserErr     error
	GetBookmarksForUserCalls   []int32

	ListGrantsByUserIDReturns []*Grant
	ListGrantsByUserIDErr     error
	ListGrantsByUserIDCalls   []sql.NullInt32

	AdminInsertBannedIpCalls []AdminInsertBannedIpParams
	AdminInsertBannedIpErr   error
	AdminInsertBannedIpFn    func(context.Context, AdminInsertBannedIpParams) error

	InsertPasswordCalls []InsertPasswordParams
	InsertPasswordErr   error
	InsertPasswordFn    func(context.Context, InsertPasswordParams) error

	SystemDeletePasswordResetsByUserCalls  []int32
	SystemDeletePasswordResetsByUserErr    error
	SystemDeletePasswordResetsByUserResult sql.Result
	SystemDeletePasswordResetsByUserFn     func(context.Context, int32) (sql.Result, error)

	AdminPromoteAnnouncementCalls []int32
	AdminPromoteAnnouncementErr   error
	AdminPromoteAnnouncementFn    func(context.Context, int32) error

	AdminDemoteAnnouncementCalls []int32
	AdminDemoteAnnouncementErr   error
	AdminDemoteAnnouncementFn    func(context.Context, int32) error

	AdminDeleteForumThreadCalls []int32
	AdminDeleteForumThreadErr   error
	AdminDeleteForumThreadFn    func(context.Context, int32) error

	AdminCancelBannedIpCalls      []string
	AdminCancelBannedIpErr        error
	AdminCancelBannedIpFn         func(context.Context, string) error
	GetPasswordResetByUserCalls   []GetPasswordResetByUserParams
	GetPasswordResetByUserReturns *PendingPassword
	GetPasswordResetByUserErr     error
	GetPasswordResetByUserFn      func(context.Context, GetPasswordResetByUserParams) (*PendingPassword, error)

	CreatePasswordResetForUserCalls []CreatePasswordResetForUserParams
	CreatePasswordResetForUserErr   error
	CreatePasswordResetForUserFn    func(context.Context, CreatePasswordResetForUserParams) error

	AdminInsertRequestQueueCalls   []AdminInsertRequestQueueParams
	AdminInsertRequestQueueReturns sql.Result
	AdminInsertRequestQueueErr     error
	AdminInsertRequestQueueFn      func(context.Context, AdminInsertRequestQueueParams) (sql.Result, error)

	SystemGetLoginRow   *SystemGetLoginRow
	SystemGetLoginErr   error
	SystemGetLoginCalls []sql.NullString
	SystemGetLoginFn    func(context.Context, sql.NullString) (*SystemGetLoginRow, error)

	SystemListVerifiedEmailsByUserIDReturn []*UserEmail
	SystemListVerifiedEmailsByUserIDErr    error
	SystemListVerifiedEmailsByUserIDCalls  []int32
	SystemListVerifiedEmailsByUserIDFn     func(context.Context, int32) ([]*UserEmail, error)

	SystemRebuildForumTopicMetaByIDCalls []int32
	SystemRebuildForumTopicMetaByIDErr   error
	SystemRebuildForumTopicMetaByIDFn    func(context.Context, int32) error

	GetLoginRoleForUserReturns int32
	GetLoginRoleForUserErr     error
	GetLoginRoleForUserCalls   []int32
	GetLoginRoleForUserFn      func(context.Context, int32) (int32, error)

	SystemListPendingEmailsCalls  []SystemListPendingEmailsParams
	SystemListPendingEmailsReturn []*SystemListPendingEmailsRow
	SystemListPendingEmailsErr    error
	SystemListPendingEmailsFn     func(context.Context, SystemListPendingEmailsParams) ([]*SystemListPendingEmailsRow, error)

	SystemMarkPendingEmailSentCalls []int32
	SystemMarkPendingEmailSentErr   error
	SystemMarkPendingEmailSentFn    func(context.Context, int32) error

	ListGrantsExtendedReturns []*ListGrantsExtendedRow
	ListGrantsExtendedErr     error
	ListGrantsExtendedCalls   []ListGrantsExtendedParams
	ListGrantsExtendedFn      func(context.Context, ListGrantsExtendedParams) ([]*ListGrantsExtendedRow, error)

	AdminDeletePendingEmailCalls []int32
	AdminDeletePendingEmailErr   error
	AdminDeletePendingEmailFn    func(context.Context, int32) error

	ListAdminUserCommentsReturns []*AdminUserComment
	ListAdminUserCommentsErr     error
	ListAdminUserCommentsCalls   []int32

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
	SystemGetTemplateOverrideFn      func(context.Context, string) (string, error)

	AdminSetTemplateOverrideCalls []AdminSetTemplateOverrideParams
	AdminSetTemplateOverrideErr   error

	ListSubscribersForPatternsParams []ListSubscribersForPatternsParams
	ListSubscribersForPatternsReturn map[string][]int32

	GetPreferenceForListerParams []int32
	GetPreferenceForListerReturn map[int32]*Preference

	InsertSubscriptionParams         []InsertSubscriptionParams
	DeleteSubscriptionParams         []DeleteSubscriptionForSubscriberParams
	DeleteSubscriptionErr            error
	CreateFAQQuestionForWriterCalls  []CreateFAQQuestionForWriterParams
	CreateFAQQuestionForWriterErr    error
	CreateFAQQuestionForWriterFn     func(context.Context, CreateFAQQuestionForWriterParams) error
	InsertFAQQuestionForWriterCalls  []InsertFAQQuestionForWriterParams
	InsertFAQQuestionForWriterResult sql.Result
	InsertFAQQuestionForWriterErr    error

	ListSubscribersForPatternParams []ListSubscribersForPatternParams
	ListSubscribersForPatternReturn map[string][]int32

	GetCommentByIdForUserCalls []GetCommentByIdForUserParams
	GetCommentByIdForUserRow   *GetCommentByIdForUserRow
	GetCommentByIdForUserErr   error
	GetCommentByIdForUserFn    func(context.Context, GetCommentByIdForUserParams) (*GetCommentByIdForUserRow, error)

	CreateCommentInSectionForCommenterCalls  []CreateCommentInSectionForCommenterParams
	CreateCommentInSectionForCommenterFn     func(context.Context, CreateCommentInSectionForCommenterParams) (int64, error)
	CreateCommentInSectionForCommenterResult int64
	CreateCommentInSectionForCommenterErr    error

	ListUploadedImagePathsByUserCalls   []ListUploadedImagePathsByUserParams
	ListUploadedImagePathsByUserFn      func(context.Context, ListUploadedImagePathsByUserParams) ([]sql.NullString, error)
	ListUploadedImagePathsByUserReturns []sql.NullString
	ListUploadedImagePathsByUserErr     error

	ListThreadImagePathsCalls   []ListThreadImagePathsParams
	ListThreadImagePathsFn      func(context.Context, ListThreadImagePathsParams) ([]sql.NullString, error)
	ListThreadImagePathsReturns []sql.NullString
	ListThreadImagePathsErr     error

	CreateThreadImageCalls []CreateThreadImageParams
	CreateThreadImageFn    func(context.Context, CreateThreadImageParams) error
	CreateThreadImageErr   error

	UpsertContentReadMarkerCalls []UpsertContentReadMarkerParams
	UpsertContentReadMarkerFn    func(context.Context, UpsertContentReadMarkerParams) error
	UpsertContentReadMarkerErr   error

	SystemGetUserByIDFn func(context.Context, int32) (*SystemGetUserByIDRow, error)

	SystemInsertDeadLetterFn  func(context.Context, string) error
	SystemInsertDeadLetterErr error

	GetThreadLastPosterAndPermsFn func(context.Context, GetThreadLastPosterAndPermsParams) (*GetThreadLastPosterAndPermsRow, error)

	GetThreadBySectionThreadIDForReplierFn func(context.Context, GetThreadBySectionThreadIDForReplierParams) (*Forumthread, error)

	RemoveContentPrivateLabelFn func(context.Context, RemoveContentPrivateLabelParams) error

	GetForumTopicByIdForUserFn      func(context.Context, GetForumTopicByIdForUserParams) (*GetForumTopicByIdForUserRow, error)
	GetForumTopicByIdForUserCalls   []GetForumTopicByIdForUserParams
	GetForumTopicByIdForUserReturns *GetForumTopicByIdForUserRow
	GetForumTopicByIdForUserErr     error
	GetForumTopicByIdFn             func(context.Context, int32) (*Forumtopic, error)
	GetForumTopicByIdCalls          []int32
	GetForumTopicByIdReturns        *Forumtopic
	GetForumTopicByIdErr            error

	ListPrivateTopicParticipantsByTopicIDForUserFn func(context.Context, ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*ListPrivateTopicParticipantsByTopicIDForUserRow, error)

	GetCommentsBySectionThreadIdForUserCalls   []GetCommentsBySectionThreadIdForUserParams
	GetCommentsBySectionThreadIdForUserReturns []*GetCommentsBySectionThreadIdForUserRow
	GetCommentsBySectionThreadIdForUserErr     error

	GetCommentsByThreadIdForUserCalls   []GetCommentsByThreadIdForUserParams
	GetCommentsByThreadIdForUserReturns []*GetCommentsByThreadIdForUserRow
	GetCommentsByThreadIdForUserErr     error

	DeleteThreadsByTopicIDCalls []int32
	DeleteThreadsByTopicIDErr   error

	SystemCheckGrantReturns int32
	SystemCheckGrantErr     error
	SystemCheckGrantCalls   []SystemCheckGrantParams
	SystemCheckGrantFn      func(SystemCheckGrantParams) (int32, error)

	GetWritingForListerByIDRow   *GetWritingForListerByIDRow
	GetWritingForListerByIDErr   error
	GetWritingForListerByIDCalls []GetWritingForListerByIDParams

	GetBlogEntryForListerByIDRow   *GetBlogEntryForListerByIDRow
	GetBlogEntryForListerByIDErr   error
	GetBlogEntryForListerByIDCalls []GetBlogEntryForListerByIDParams

	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow   *GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingErr   error
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingCalls []int32

	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow   *GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserErr   error
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls []GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams

	GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextCalls []GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams
	GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextFn    func(context.Context, GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error)

	GetAllForumCategoriesCalls   []GetAllForumCategoriesParams
	GetAllForumCategoriesReturns []*Forumcategory
	GetAllForumCategoriesErr     error
	GetAllForumCategoriesFn      func(context.Context, GetAllForumCategoriesParams) ([]*Forumcategory, error)
	GetForumCategoryByIdCalls    []GetForumCategoryByIdParams
	GetForumCategoryByIdReturns  *Forumcategory
	GetForumCategoryByIdErr      error
	GetForumCategoryByIdFn       func(context.Context, GetForumCategoryByIdParams) (*Forumcategory, error)

	GetAllForumTopicsByCategoryIdForUserWithLastPosterNameCalls   []GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams
	GetAllForumTopicsByCategoryIdForUserWithLastPosterNameReturns []*GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow
	GetAllForumTopicsByCategoryIdForUserWithLastPosterNameErr     error
	GetAllForumTopicsByCategoryIdForUserWithLastPosterNameFn      func(context.Context, GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams) ([]*GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow, error)

	SystemCheckRoleGrantReturns int32
	SystemCheckRoleGrantErr     error
	SystemCheckRoleGrantCalls   []SystemCheckRoleGrantParams
	SystemCheckRoleGrantFn      func(SystemCheckRoleGrantParams) (int32, error)

	GetPermissionsByUserIDReturns []*GetPermissionsByUserIDRow
	GetPermissionsByUserIDErr     error
	GetPermissionsByUserIDCalls   []int32
	GetPermissionsByUserIDFn      func(int32) ([]*GetPermissionsByUserIDRow, error)

	GetThreadBySectionThreadIDForReplierCalls  []GetThreadBySectionThreadIDForReplierParams
	GetThreadBySectionThreadIDForReplierReturn *Forumthread
	GetThreadBySectionThreadIDForReplierErr    error

	GetUnreadNotificationCountForListerCalls   []int32
	GetUnreadNotificationCountForListerReturns int64
	GetUnreadNotificationCountForListerErr     error

	GetNotificationCountForListerCalls   []int32
	GetNotificationCountForListerReturns int64
	GetNotificationCountForListerErr     error

	GetThreadLastPosterAndPermsCalls   []GetThreadLastPosterAndPermsParams
	GetThreadLastPosterAndPermsReturns *GetThreadLastPosterAndPermsRow
	GetThreadLastPosterAndPermsErr     error

	ListPrivateTopicParticipantsByTopicIDForUserCalls   []ListPrivateTopicParticipantsByTopicIDForUserParams
	ListPrivateTopicParticipantsByTopicIDForUserReturns []*ListPrivateTopicParticipantsByTopicIDForUserRow
	ListPrivateTopicParticipantsByTopicIDForUserErr     error

	ListPrivateTopicsByUserIDCalls   []sql.NullInt32
	ListPrivateTopicsByUserIDReturns []*ListPrivateTopicsByUserIDRow
	ListPrivateTopicsByUserIDErr     error
	ListPrivateTopicsByUserIDFn      func(context.Context, sql.NullInt32) ([]*ListPrivateTopicsByUserIDRow, error)

	AdminListForumTopicGrantsByTopicIDCalls   []sql.NullInt32
	AdminListForumTopicGrantsByTopicIDReturns []*AdminListForumTopicGrantsByTopicIDRow
	AdminListForumTopicGrantsByTopicIDErr     error

	AdminListPrivateTopicParticipantsByTopicIDCalls   []sql.NullInt32
	AdminListPrivateTopicParticipantsByTopicIDReturns []*AdminListPrivateTopicParticipantsByTopicIDRow
	AdminListPrivateTopicParticipantsByTopicIDErr     error

	AdminCreateForumCategoryCalls   []AdminCreateForumCategoryParams
	AdminCreateForumCategoryReturns int64
	AdminCreateForumCategoryErr     error
	AdminCreateForumCategoryFn      func(context.Context, AdminCreateForumCategoryParams) (int64, error)

	AdminUpdateForumCategoryCalls []AdminUpdateForumCategoryParams
	AdminUpdateForumCategoryErr   error
	AdminUpdateForumCategoryFn    func(context.Context, AdminUpdateForumCategoryParams) error

	AdminCountForumTopicsCalls   int
	AdminCountForumTopicsReturns int64
	AdminCountForumTopicsErr     error
	AdminCountForumTopicsFn      func(context.Context) (int64, error)

	AdminListForumTopicsCalls   []AdminListForumTopicsParams
	AdminListForumTopicsReturns []*Forumtopic
	AdminListForumTopicsErr     error
	AdminListForumTopicsFn      func(context.Context, AdminListForumTopicsParams) ([]*Forumtopic, error)

	AdminGetTopicGrantsCalls   []sql.NullInt32
	AdminGetTopicGrantsReturns []*AdminGetTopicGrantsRow
	AdminGetTopicGrantsErr     error
	AdminGetTopicGrantsFn      func(context.Context, sql.NullInt32) ([]*AdminGetTopicGrantsRow, error)

	AdminListRolesCalls   int
	AdminListRolesReturns []*Role
	AdminListRolesErr     error
	AdminListRolesFn      func(context.Context) ([]*Role, error)

	ListGrantsCalls   int
	ListGrantsReturns []*Grant
	ListGrantsErr     error
	ListGrantsFn      func(context.Context) ([]*Grant, error)

	ListBloggersForListerCalls   []ListBloggersForListerParams
	ListBloggersForListerReturns []*ListBloggersForListerRow
	ListBloggersForListerErr     error
	ListBloggersForListerFn      func(ListBloggersForListerParams) ([]*ListBloggersForListerRow, error)

	ListWritersForListerCalls   []ListWritersForListerParams
	ListWritersForListerReturns []*ListWritersForListerRow
	ListWritersForListerErr     error
	ListWritersForListerFn      func(ListWritersForListerParams) ([]*ListWritersForListerRow, error)

	ListWritersSearchForListerCalls   []ListWritersSearchForListerParams
	ListWritersSearchForListerReturns []*ListWritersSearchForListerRow
	ListWritersSearchForListerErr     error
	ListWritersSearchForListerFn      func(ListWritersSearchForListerParams) ([]*ListWritersSearchForListerRow, error)

	ListBlogEntriesByAuthorForListerCalls   []ListBlogEntriesByAuthorForListerParams
	ListBlogEntriesByAuthorForListerReturns []*ListBlogEntriesByAuthorForListerRow
	ListBlogEntriesByAuthorForListerErr     error
	ListBlogEntriesByAuthorForListerFn      func(context.Context, ListBlogEntriesByAuthorForListerParams) ([]*ListBlogEntriesByAuthorForListerRow, error)

	ListBlogEntriesByIDsForListerCalls   []ListBlogEntriesByIDsForListerParams
	ListBlogEntriesByIDsForListerReturns []*ListBlogEntriesByIDsForListerRow
	ListBlogEntriesByIDsForListerErr     error
	ListBlogEntriesByIDsForListerFn      func(context.Context, ListBlogEntriesByIDsForListerParams) ([]*ListBlogEntriesByIDsForListerRow, error)
  
	ListSiteNewsSearchFirstForListerCalls   []ListSiteNewsSearchFirstForListerParams
	ListSiteNewsSearchFirstForListerReturns []int32
	ListSiteNewsSearchFirstForListerErr     error
	ListSiteNewsSearchFirstForListerFn      func(ListSiteNewsSearchFirstForListerParams) ([]int32, error)

	ListSiteNewsSearchNextForListerCalls   []ListSiteNewsSearchNextForListerParams
	ListSiteNewsSearchNextForListerReturns []int32
	ListSiteNewsSearchNextForListerErr     error
	ListSiteNewsSearchNextForListerFn      func(ListSiteNewsSearchNextForListerParams) ([]int32, error)

	GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountCalls   []GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams
	GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountReturns []*GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow
	GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountErr     error
	GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountFn      func(GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams) ([]*GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow, error)

	ListImagePostsByBoardForListerCalls   []ListImagePostsByBoardForListerParams
	ListImagePostsByBoardForListerReturns []*ListImagePostsByBoardForListerRow
	ListImagePostsByBoardForListerErr     error
	ListImagePostsByBoardForListerFn      func(ListImagePostsByBoardForListerParams) ([]*ListImagePostsByBoardForListerRow, error)

	ListBoardsByParentIDForListerCalls   []ListBoardsByParentIDForListerParams
	ListBoardsByParentIDForListerReturns []*Imageboard
	ListBoardsByParentIDForListerErr     error
	ListBoardsByParentIDForListerFn      func(ListBoardsByParentIDForListerParams) ([]*Imageboard, error)

	UpdateTimezoneForListerCalls []UpdateTimezoneForListerParams
	UpdateTimezoneForListerErr   error
	UpdateTimezoneForListerFn    func(context.Context, UpdateTimezoneForListerParams) error
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
	if s.ListContentPublicLabelsFn != nil {
		return s.ListContentPublicLabelsFn(arg)
	}
	if s.ListContentPublicLabelsReturns != nil && s.ContentPublicLabels == nil {
		return s.ListContentPublicLabelsReturns, s.ListContentPublicLabelsErr
	}
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

func (s *QuerierStub) CreateCommentInSectionForCommenter(ctx context.Context, arg CreateCommentInSectionForCommenterParams) (int64, error) {
	s.mu.Lock()
	s.CreateCommentInSectionForCommenterCalls = append(s.CreateCommentInSectionForCommenterCalls, arg)
	fn := s.CreateCommentInSectionForCommenterFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.CreateCommentInSectionForCommenterResult, s.CreateCommentInSectionForCommenterErr
}

// ListUploadedImagePathsByUser records the call and returns stored paths.
func (s *QuerierStub) ListUploadedImagePathsByUser(ctx context.Context, arg ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
	s.mu.Lock()
	s.ListUploadedImagePathsByUserCalls = append(s.ListUploadedImagePathsByUserCalls, arg)
	fn := s.ListUploadedImagePathsByUserFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	if s.ListUploadedImagePathsByUserErr != nil {
		return nil, s.ListUploadedImagePathsByUserErr
	}
	return s.ListUploadedImagePathsByUserReturns, nil
}

// ListThreadImagePaths records the call and returns stored paths.
func (s *QuerierStub) ListThreadImagePaths(ctx context.Context, arg ListThreadImagePathsParams) ([]sql.NullString, error) {
	s.mu.Lock()
	s.ListThreadImagePathsCalls = append(s.ListThreadImagePathsCalls, arg)
	fn := s.ListThreadImagePathsFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	if s.ListThreadImagePathsErr != nil {
		return nil, s.ListThreadImagePathsErr
	}
	return s.ListThreadImagePathsReturns, nil
}

// CreateThreadImage records the call and returns a stubbed error.
func (s *QuerierStub) CreateThreadImage(ctx context.Context, arg CreateThreadImageParams) error {
	s.mu.Lock()
	s.CreateThreadImageCalls = append(s.CreateThreadImageCalls, arg)
	fn := s.CreateThreadImageFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.CreateThreadImageErr
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

func (s *QuerierStub) ClearUnreadContentPrivateLabelExceptUser(ctx context.Context, arg ClearUnreadContentPrivateLabelExceptUserParams) error {
	s.mu.Lock()
	s.ClearUnreadContentPrivateLabelExceptUserCalls = append(s.ClearUnreadContentPrivateLabelExceptUserCalls, arg)
	fn := s.ClearUnreadContentPrivateLabelExceptUserFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.ClearUnreadContentPrivateLabelExceptUserErr
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
	if s.ListContentLabelStatusReturns != nil && s.ContentLabelStatus == nil {
		return s.ListContentLabelStatusReturns, s.ListContentLabelStatusErr
	}
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
	s.AddContentPrivateLabelCalls = append(s.AddContentPrivateLabelCalls, arg)
	fn := s.AddContentPrivateLabelFn
	err := s.AddContentPrivateLabelErr
	ignoreErr := s.AddContentPrivateLabelIgnoreLabels != nil && s.AddContentPrivateLabelIgnoreLabels[arg.Label]
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	if err != nil && !ignoreErr {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	labels := s.ensurePrivateLabelSetLocked(arg.Item, arg.ItemID, arg.UserID)
	labels[arg.Label] = arg.Invert
	return nil
}

// ListContentPrivateLabels records the call and returns stored private labels.
func (s *QuerierStub) ListContentPrivateLabels(ctx context.Context, arg ListContentPrivateLabelsParams) ([]*ListContentPrivateLabelsRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListContentPrivateLabelsCalls = append(s.ListContentPrivateLabelsCalls, arg)

	if s.ListContentPrivateLabelsFn != nil {
		return s.ListContentPrivateLabelsFn(arg)
	}
	if s.ListContentPrivateLabelsReturns != nil && s.ContentPrivateLabels == nil {
		return s.ListContentPrivateLabelsReturns, s.ListContentPrivateLabelsErr
	}

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
	s.RemoveContentPrivateLabelCalls = append(s.RemoveContentPrivateLabelCalls, arg)
	fn := s.RemoveContentPrivateLabelFn
	err := s.RemoveContentPrivateLabelErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (s *QuerierStub) AdminCreateForumCategory(ctx context.Context, arg AdminCreateForumCategoryParams) (int64, error) {
	s.mu.Lock()
	s.AdminCreateForumCategoryCalls = append(s.AdminCreateForumCategoryCalls, arg)
	fn := s.AdminCreateForumCategoryFn
	ret := s.AdminCreateForumCategoryReturns
	err := s.AdminCreateForumCategoryErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) AdminUpdateForumCategory(ctx context.Context, arg AdminUpdateForumCategoryParams) error {
	s.mu.Lock()
	s.AdminUpdateForumCategoryCalls = append(s.AdminUpdateForumCategoryCalls, arg)
	fn := s.AdminUpdateForumCategoryFn
	err := s.AdminUpdateForumCategoryErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminCountForumTopics(ctx context.Context) (int64, error) {
	s.mu.Lock()
	s.AdminCountForumTopicsCalls++
	fn := s.AdminCountForumTopicsFn
	ret := s.AdminCountForumTopicsReturns
	err := s.AdminCountForumTopicsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) AdminListForumTopics(ctx context.Context, arg AdminListForumTopicsParams) ([]*Forumtopic, error) {
	s.mu.Lock()
	s.AdminListForumTopicsCalls = append(s.AdminListForumTopicsCalls, arg)
	fn := s.AdminListForumTopicsFn
	ret := s.AdminListForumTopicsReturns
	err := s.AdminListForumTopicsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) AdminGetTopicGrants(ctx context.Context, topicID sql.NullInt32) ([]*AdminGetTopicGrantsRow, error) {
	s.mu.Lock()
	s.AdminGetTopicGrantsCalls = append(s.AdminGetTopicGrantsCalls, topicID)
	fn := s.AdminGetTopicGrantsFn
	ret := s.AdminGetTopicGrantsReturns
	err := s.AdminGetTopicGrantsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, topicID)
	}
	return ret, err
}

func (s *QuerierStub) AdminListRoles(ctx context.Context) ([]*Role, error) {
	s.mu.Lock()
	s.AdminListRolesCalls++
	fn := s.AdminListRolesFn
	ret := s.AdminListRolesReturns
	err := s.AdminListRolesErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) ListGrants(ctx context.Context) ([]*Grant, error) {
	s.mu.Lock()
	s.ListGrantsCalls++
	fn := s.ListGrantsFn
	ret := s.ListGrantsReturns
	err := s.ListGrantsErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx)
	}
	return ret, err
}

func (s *QuerierStub) ListBloggersForLister(ctx context.Context, arg ListBloggersForListerParams) ([]*ListBloggersForListerRow, error) {
	s.mu.Lock()
	s.ListBloggersForListerCalls = append(s.ListBloggersForListerCalls, arg)
	fn := s.ListBloggersForListerFn
	ret := s.ListBloggersForListerReturns
	err := s.ListBloggersForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
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

func (s *QuerierStub) ListSiteNewsSearchFirstForLister(ctx context.Context, arg ListSiteNewsSearchFirstForListerParams) ([]int32, error) {
	s.mu.Lock()
	s.ListSiteNewsSearchFirstForListerCalls = append(s.ListSiteNewsSearchFirstForListerCalls, arg)
	fn := s.ListSiteNewsSearchFirstForListerFn
	ret := s.ListSiteNewsSearchFirstForListerReturns
	err := s.ListSiteNewsSearchFirstForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

func (s *QuerierStub) ListSiteNewsSearchNextForLister(ctx context.Context, arg ListSiteNewsSearchNextForListerParams) ([]int32, error) {
	s.mu.Lock()
	s.ListSiteNewsSearchNextForListerCalls = append(s.ListSiteNewsSearchNextForListerCalls, arg)
	fn := s.ListSiteNewsSearchNextForListerFn
	ret := s.ListSiteNewsSearchNextForListerReturns
	err := s.ListSiteNewsSearchNextForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

func (s *QuerierStub) GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCount(ctx context.Context, arg GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountParams) ([]*GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow, error) {
	s.mu.Lock()
	s.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountCalls = append(s.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountCalls, arg)
	fn := s.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountFn
	ret := s.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountReturns
	err := s.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

func (s *QuerierStub) ListImagePostsByBoardForLister(ctx context.Context, arg ListImagePostsByBoardForListerParams) ([]*ListImagePostsByBoardForListerRow, error) {
	s.mu.Lock()
	s.ListImagePostsByBoardForListerCalls = append(s.ListImagePostsByBoardForListerCalls, arg)
	fn := s.ListImagePostsByBoardForListerFn
	ret := s.ListImagePostsByBoardForListerReturns
	err := s.ListImagePostsByBoardForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

func (s *QuerierStub) ListBoardsByParentIDForLister(ctx context.Context, arg ListBoardsByParentIDForListerParams) ([]*Imageboard, error) {
	s.mu.Lock()
	s.ListBoardsByParentIDForListerCalls = append(s.ListBoardsByParentIDForListerCalls, arg)
	fn := s.ListBoardsByParentIDForListerFn
	ret := s.ListBoardsByParentIDForListerReturns
	err := s.ListBoardsByParentIDForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(arg)
	}
	return ret, err
}

func (s *QuerierStub) UpdateTimezoneForLister(ctx context.Context, arg UpdateTimezoneForListerParams) error {
	s.mu.Lock()
	s.UpdateTimezoneForListerCalls = append(s.UpdateTimezoneForListerCalls, arg)
	fn := s.UpdateTimezoneForListerFn
	err := s.UpdateTimezoneForListerErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
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
	fn := s.GetCommentByIdForUserFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	if s.GetCommentByIdForUserErr != nil {
		return nil, s.GetCommentByIdForUserErr
	}
	return s.GetCommentByIdForUserRow, nil
}

func (s *QuerierStub) GetCommentsBySectionThreadIdForUser(ctx context.Context, arg GetCommentsBySectionThreadIdForUserParams) ([]*GetCommentsBySectionThreadIdForUserRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetCommentsBySectionThreadIdForUserCalls = append(s.GetCommentsBySectionThreadIdForUserCalls, arg)
	return s.GetCommentsBySectionThreadIdForUserReturns, s.GetCommentsBySectionThreadIdForUserErr
}

func (s *QuerierStub) GetCommentsByThreadIdForUser(ctx context.Context, arg GetCommentsByThreadIdForUserParams) ([]*GetCommentsByThreadIdForUserRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetCommentsByThreadIdForUserCalls = append(s.GetCommentsByThreadIdForUserCalls, arg)
	return s.GetCommentsByThreadIdForUserReturns, s.GetCommentsByThreadIdForUserErr
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

func (s *QuerierStub) GetBlogEntryForListerByID(ctx context.Context, arg GetBlogEntryForListerByIDParams) (*GetBlogEntryForListerByIDRow, error) {
	s.mu.Lock()
	s.GetBlogEntryForListerByIDCalls = append(s.GetBlogEntryForListerByIDCalls, arg)
	row := s.GetBlogEntryForListerByIDRow
	err := s.GetBlogEntryForListerByIDErr
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("GetBlogEntryForListerByID not stubbed")
	}
	return row, nil
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
	return row, nil
}

func (s *QuerierStub) GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(ctx context.Context, arg GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams) ([]*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	s.mu.Lock()
	s.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextCalls = append(s.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextCalls, arg)
	fn := s.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return nil, nil
}

func (s *QuerierStub) GetAllForumCategories(ctx context.Context, arg GetAllForumCategoriesParams) ([]*Forumcategory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetAllForumCategoriesCalls = append(s.GetAllForumCategoriesCalls, arg)
	fn := s.GetAllForumCategoriesFn
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.GetAllForumCategoriesReturns, s.GetAllForumCategoriesErr
}

func (s *QuerierStub) GetForumCategoryById(ctx context.Context, arg GetForumCategoryByIdParams) (*Forumcategory, error) {
	s.mu.Lock()
	s.GetForumCategoryByIdCalls = append(s.GetForumCategoryByIdCalls, arg)
	fn := s.GetForumCategoryByIdFn
	row := s.GetForumCategoryByIdReturns
	err := s.GetForumCategoryByIdErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return row, err
}

func (s *QuerierStub) GetAllForumTopicsByCategoryIdForUserWithLastPosterName(ctx context.Context, arg GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams) ([]*GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow, error) {
	s.mu.Lock()
	s.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameCalls = append(s.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameCalls, arg)
	fn := s.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameFn
	rows := s.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameReturns
	err := s.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return rows, err
}

// SystemCheckRoleGrant records the call and returns the configured response.

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

func (s *QuerierStub) GetThreadBySectionThreadIDForReplier(ctx context.Context, arg GetThreadBySectionThreadIDForReplierParams) (*Forumthread, error) {
	s.mu.Lock()
	s.GetThreadBySectionThreadIDForReplierCalls = append(s.GetThreadBySectionThreadIDForReplierCalls, arg)
	fn := s.GetThreadBySectionThreadIDForReplierFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.GetThreadBySectionThreadIDForReplierReturn, s.GetThreadBySectionThreadIDForReplierErr
}

func (s *QuerierStub) GetThreadLastPosterAndPerms(ctx context.Context, arg GetThreadLastPosterAndPermsParams) (*GetThreadLastPosterAndPermsRow, error) {
	s.mu.Lock()
	s.GetThreadLastPosterAndPermsCalls = append(s.GetThreadLastPosterAndPermsCalls, arg)
	fn := s.GetThreadLastPosterAndPermsFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.GetThreadLastPosterAndPermsReturns, s.GetThreadLastPosterAndPermsErr
}

func (s *QuerierStub) GetForumTopicByIdForUser(ctx context.Context, arg GetForumTopicByIdForUserParams) (*GetForumTopicByIdForUserRow, error) {
	s.mu.Lock()
	s.GetForumTopicByIdForUserCalls = append(s.GetForumTopicByIdForUserCalls, arg)
	fn := s.GetForumTopicByIdForUserFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.GetForumTopicByIdForUserReturns, s.GetForumTopicByIdForUserErr
}

func (s *QuerierStub) GetForumTopicById(ctx context.Context, idforumtopic int32) (*Forumtopic, error) {
	s.mu.Lock()
	s.GetForumTopicByIdCalls = append(s.GetForumTopicByIdCalls, idforumtopic)
	fn := s.GetForumTopicByIdFn
	row := s.GetForumTopicByIdReturns
	err := s.GetForumTopicByIdErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, idforumtopic)
	}
	return row, err
}

func (s *QuerierStub) ListPrivateTopicParticipantsByTopicIDForUser(ctx context.Context, arg ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
	s.mu.Lock()
	s.ListPrivateTopicParticipantsByTopicIDForUserCalls = append(s.ListPrivateTopicParticipantsByTopicIDForUserCalls, arg)
	fn := s.ListPrivateTopicParticipantsByTopicIDForUserFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.ListPrivateTopicParticipantsByTopicIDForUserReturns, s.ListPrivateTopicParticipantsByTopicIDForUserErr
}

// ListPrivateTopicsByUserID records the call and returns stubbed topics.
func (s *QuerierStub) ListPrivateTopicsByUserID(ctx context.Context, userID sql.NullInt32) ([]*ListPrivateTopicsByUserIDRow, error) {
	s.mu.Lock()
	s.ListPrivateTopicsByUserIDCalls = append(s.ListPrivateTopicsByUserIDCalls, userID)
	fn := s.ListPrivateTopicsByUserIDFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, userID)
	}
	return s.ListPrivateTopicsByUserIDReturns, s.ListPrivateTopicsByUserIDErr
}

// SystemGetUserByID records the call and returns the configured response.
func (s *QuerierStub) SystemGetUserByID(ctx context.Context, idusers int32) (*SystemGetUserByIDRow, error) {
	s.mu.Lock()
	s.SystemGetUserByIDCalls = append(s.SystemGetUserByIDCalls, idusers)
	fn := s.SystemGetUserByIDFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, idusers)
	}
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
	fn := s.SystemGetUserByEmailFn
	row := s.SystemGetUserByEmailRow
	err := s.SystemGetUserByEmailErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, email)
	}
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("SystemGetUserByEmail not stubbed")
	}
	return row, nil
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

func (s *QuerierStub) SystemCreateThread(ctx context.Context, forumtopicIdforumtopic int32) (int64, error) {
	s.mu.Lock()
	s.SystemCreateThreadCalls = append(s.SystemCreateThreadCalls, forumtopicIdforumtopic)
	fn := s.SystemCreateThreadFn
	ret := s.SystemCreateThreadReturns
	err := s.SystemCreateThreadErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, forumtopicIdforumtopic)
	}
	return ret, err
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

func (s *QuerierStub) AdminListUserEmails(ctx context.Context, id int32) ([]*UserEmail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminListUserEmailsCalls = append(s.AdminListUserEmailsCalls, id)
	return s.AdminListUserEmailsReturns, s.AdminListUserEmailsErr
}

func (s *QuerierStub) AdminUserPostCountsByID(ctx context.Context, id int32) (*AdminUserPostCountsByIDRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdminUserPostCountsByIDCalls = append(s.AdminUserPostCountsByIDCalls, id)
	return s.AdminUserPostCountsByIDReturns, s.AdminUserPostCountsByIDErr
}

func (s *QuerierStub) GetBookmarksForUser(ctx context.Context, id int32) (*GetBookmarksForUserRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetBookmarksForUserCalls = append(s.GetBookmarksForUserCalls, id)
	return s.GetBookmarksForUserReturns, s.GetBookmarksForUserErr
}

func (s *QuerierStub) ListGrantsByUserID(ctx context.Context, userID sql.NullInt32) ([]*Grant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListGrantsByUserIDCalls = append(s.ListGrantsByUserIDCalls, userID)
	return s.ListGrantsByUserIDReturns, s.ListGrantsByUserIDErr
}

func (s *QuerierStub) ListGrantsExtended(ctx context.Context, arg ListGrantsExtendedParams) ([]*ListGrantsExtendedRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListGrantsExtendedCalls = append(s.ListGrantsExtendedCalls, arg)
	if s.ListGrantsExtendedFn != nil {
		return s.ListGrantsExtendedFn(ctx, arg)
	}
	return s.ListGrantsExtendedReturns, s.ListGrantsExtendedErr
}

func (s *QuerierStub) ListAdminUserComments(ctx context.Context, userID int32) ([]*AdminUserComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ListAdminUserCommentsCalls = append(s.ListAdminUserCommentsCalls, userID)
	return s.ListAdminUserCommentsReturns, s.ListAdminUserCommentsErr
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
	fn := s.SystemGetTemplateOverrideFn
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
	if fn != nil {
		return fn(ctx, name)
	}
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

func (s *QuerierStub) DeleteSubscriptionForSubscriber(ctx context.Context, arg DeleteSubscriptionForSubscriberParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DeleteSubscriptionParams = append(s.DeleteSubscriptionParams, arg)
	return s.DeleteSubscriptionErr
}

func (s *QuerierStub) CreateFAQQuestionForWriter(ctx context.Context, arg CreateFAQQuestionForWriterParams) error {
	s.mu.Lock()
	s.CreateFAQQuestionForWriterCalls = append(s.CreateFAQQuestionForWriterCalls, arg)
	fn := s.CreateFAQQuestionForWriterFn
	err := s.CreateFAQQuestionForWriterErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
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
func (s *QuerierStub) GetUnreadNotificationCountForLister(ctx context.Context, listerID int32) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetUnreadNotificationCountForListerCalls = append(s.GetUnreadNotificationCountForListerCalls, listerID)
	return s.GetUnreadNotificationCountForListerReturns, s.GetUnreadNotificationCountForListerErr
}

func (s *QuerierStub) GetNotificationCountForLister(ctx context.Context, listerID int32) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetNotificationCountForListerCalls = append(s.GetNotificationCountForListerCalls, listerID)
	return s.GetNotificationCountForListerReturns, s.GetNotificationCountForListerErr
}

func (s *QuerierStub) UpsertContentReadMarker(ctx context.Context, arg UpsertContentReadMarkerParams) error {
	s.mu.Lock()
	s.UpsertContentReadMarkerCalls = append(s.UpsertContentReadMarkerCalls, arg)
	fn := s.UpsertContentReadMarkerFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return s.UpsertContentReadMarkerErr
}

func (s *QuerierStub) SystemInsertDeadLetter(ctx context.Context, message string) error {
	s.mu.Lock()
	s.SystemInsertDeadLetterCalls++
	fn := s.SystemInsertDeadLetterFn
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, message)
	}
	return s.SystemInsertDeadLetterErr
}

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
	return ret, err
}

func (s *QuerierStub) GetPasswordResetByUser(ctx context.Context, arg GetPasswordResetByUserParams) (*PendingPassword, error) {
	s.mu.Lock()
	s.GetPasswordResetByUserCalls = append(s.GetPasswordResetByUserCalls, arg)
	fn := s.GetPasswordResetByUserFn
	ret := s.GetPasswordResetByUserReturns
	err := s.GetPasswordResetByUserErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return ret, err
}

func (s *QuerierStub) CreatePasswordResetForUser(ctx context.Context, arg CreatePasswordResetForUserParams) error {
	s.mu.Lock()
	s.CreatePasswordResetForUserCalls = append(s.CreatePasswordResetForUserCalls, arg)
	fn := s.CreatePasswordResetForUserFn
	err := s.CreatePasswordResetForUserErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	return err
}

func (s *QuerierStub) AdminInsertRequestQueue(ctx context.Context, arg AdminInsertRequestQueueParams) (sql.Result, error) {
	s.mu.Lock()
	s.AdminInsertRequestQueueCalls = append(s.AdminInsertRequestQueueCalls, arg)
	fn := s.AdminInsertRequestQueueFn
	ret := s.AdminInsertRequestQueueReturns
	err := s.AdminInsertRequestQueueErr
	s.mu.Unlock()
	if fn != nil {
		return fn(ctx, arg)
	}
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return FakeSQLResult{}, nil
	}
	return ret, nil
}
