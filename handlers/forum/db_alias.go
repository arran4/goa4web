package forum

import db "github.com/arran4/goa4web/internal/db"

type (
	DBTX                                                                             = db.DBTX
	Queries                                                                          = db.Queries
	CreateCommentParams                                                              = db.CreateCommentParams
	CreateForumCategoryParams                                                        = db.CreateForumCategoryParams
	CreateForumTopicParams                                                           = db.CreateForumTopicParams
	DeleteUsersForumTopicLevelPermissionParams                                       = db.DeleteUsersForumTopicLevelPermissionParams
	Forumcategory                                                                    = db.Forumcategory
	Forumtopic                                                                       = db.Forumtopic
	GetAllForumCategoriesWithSubcategoryCountRow                                     = db.GetAllForumCategoriesWithSubcategoryCountRow
	GetAllForumThreadsWithTopicRow                                                   = db.GetAllForumThreadsWithTopicRow
	GetAllForumTopicRestrictionsWithForumTopicTitleRow                               = db.GetAllForumTopicRestrictionsWithForumTopicTitleRow
	GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams                     = db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams
	GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopicRow                   = db.GetAllForumTopicsForUserWithPermissionsRestrictionsAndTopicRow
	GetAllForumTopicsWithPermissionsAndTopicRow                                      = db.GetAllForumTopicsWithPermissionsAndTopicRow
	GetCommentByIdForUserParams                                                      = db.GetCommentByIdForUserParams
	GetCommentsByThreadIdForUserParams                                               = db.GetCommentsByThreadIdForUserParams
	GetCommentsByThreadIdForUserRow                                                  = db.GetCommentsByThreadIdForUserRow
	GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams = db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams
	GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow    = db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
	GetForumTopicByIdForUserParams                                                   = db.GetForumTopicByIdForUserParams
	GetForumTopicByIdForUserRow                                                      = db.GetForumTopicByIdForUserRow
	GetForumTopicRestrictionsByForumTopicIdRow                                       = db.GetForumTopicRestrictionsByForumTopicIdRow
	GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams                = db.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams
	GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow                   = db.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow
	GetUsersTopicLevelByUserIdAndThreadIdParams                                      = db.GetUsersTopicLevelByUserIdAndThreadIdParams
	Language                                                                         = db.Language
	ListUsersSubscribedToThreadParams                                                = db.ListUsersSubscribedToThreadParams
	UpdateCommentParams                                                              = db.UpdateCommentParams
	UpdateForumCategoryParams                                                        = db.UpdateForumCategoryParams
	UpdateForumTopicParams                                                           = db.UpdateForumTopicParams
	UpsertForumTopicRestrictionsParams                                               = db.UpsertForumTopicRestrictionsParams
	UpsertUsersForumTopicLevelPermissionParams                                       = db.UpsertUsersForumTopicLevelPermissionParams
	User                                                                             = db.User
)

func New(d db.DBTX) *Queries { return db.New(d) }
