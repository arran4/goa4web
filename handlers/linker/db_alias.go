package linker

import db "github.com/arran4/goa4web/internal/db"

type (
	DBTX                                                                              = db.DBTX
	Queries                                                                           = db.Queries
	AssignLinkerThisThreadIdParams                                                    = db.AssignLinkerThisThreadIdParams
	CreateCommentParams                                                               = db.CreateCommentParams
	CreateForumTopicParams                                                            = db.CreateForumTopicParams
	CreateLinkerCategoryParams                                                        = db.CreateLinkerCategoryParams
	CreateLinkerItemParams                                                            = db.CreateLinkerItemParams
	CreateLinkerQueuedItemParams                                                      = db.CreateLinkerQueuedItemParams
	GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams = db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams
	GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow    = db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
	GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow                        = db.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow
	GetCommentsByThreadIdForUserParams                                                = db.GetCommentsByThreadIdForUserParams
	GetCommentsByThreadIdForUserRow                                                   = db.GetCommentsByThreadIdForUserRow
	GetLinkerCategoryLinkCountsRow                                                    = db.GetLinkerCategoryLinkCountsRow
	GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow                  = db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
	GetLinkerItemsByUserDescendingParams                                              = db.GetLinkerItemsByUserDescendingParams
	GetLinkerItemsByUserDescendingRow                                                 = db.GetLinkerItemsByUserDescendingRow
	GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams                 = db.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams
	GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow                    = db.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow
	InsertPendingEmailParams                                                          = db.InsertPendingEmailParams
	Language                                                                          = db.Language
	Linkercategory                                                                    = db.Linkercategory
	ListUsersSubscribedToLinkerParams                                                 = db.ListUsersSubscribedToLinkerParams
	ListUsersSubscribedToThreadParams                                                 = db.ListUsersSubscribedToThreadParams
	Notification                                                                      = db.Notification
	Permission                                                                        = db.Permission
	PermissionUserAllowParams                                                         = db.PermissionUserAllowParams
	RenameLinkerCategoryParams                                                        = db.RenameLinkerCategoryParams
	UpdateLinkerCategorySortOrderParams                                               = db.UpdateLinkerCategorySortOrderParams
	UpdateLinkerQueuedItemParams                                                      = db.UpdateLinkerQueuedItemParams
	User                                                                              = db.User
)

func New(d db.DBTX) *Queries { return db.New(d) }
