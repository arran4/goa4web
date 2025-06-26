package blogs

import db "github.com/arran4/goa4web/internal/db"

type (
	DBTX                                                              = db.DBTX
	Queries                                                           = db.Queries
	AssignThreadIdToBlogEntryParams                                   = db.AssignThreadIdToBlogEntryParams
	Blog                                                              = db.Blog
	BloggerCountRow                                                   = db.BloggerCountRow
	CreateBlogEntryParams                                             = db.CreateBlogEntryParams
	CreateCommentParams                                               = db.CreateCommentParams
	CreateForumTopicParams                                            = db.CreateForumTopicParams
	GetBlogEntriesForUserDescendingLanguagesParams                    = db.GetBlogEntriesForUserDescendingLanguagesParams
	GetBlogEntriesForUserDescendingLanguagesRow                       = db.GetBlogEntriesForUserDescendingLanguagesRow
	GetBlogEntryForUserByIdRow                                        = db.GetBlogEntryForUserByIdRow
	GetCommentByIdForUserParams                                       = db.GetCommentByIdForUserParams
	GetCommentsByThreadIdForUserParams                                = db.GetCommentsByThreadIdForUserParams
	GetCommentsByThreadIdForUserRow                                   = db.GetCommentsByThreadIdForUserRow
	GetCountOfBlogPostsByUserRow                                      = db.GetCountOfBlogPostsByUserRow
	GetPermissionsByUserIdAndSectionBlogsRow                          = db.GetPermissionsByUserIdAndSectionBlogsRow
	GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams = db.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams
	InsertPendingEmailParams                                          = db.InsertPendingEmailParams
	Language                                                          = db.Language
	ListBloggersParams                                                = db.ListBloggersParams
	ListUsersSubscribedToBlogsParams                                  = db.ListUsersSubscribedToBlogsParams
	ListUsersSubscribedToThreadParams                                 = db.ListUsersSubscribedToThreadParams
	Notification                                                      = db.Notification
	PermissionUserAllowParams                                         = db.PermissionUserAllowParams
	SearchBloggersParams                                              = db.SearchBloggersParams
	UpdateBlogEntryParams                                             = db.UpdateBlogEntryParams
	UpdateCommentParams                                               = db.UpdateCommentParams
	User                                                              = db.User
)

func New(d db.DBTX) *Queries { return db.New(d) }
