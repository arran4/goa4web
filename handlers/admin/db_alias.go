package admin

import db "github.com/arran4/goa4web/internal/db"

type (
	DBTX                                         = db.DBTX
	Queries                                      = db.Queries
	BannedIp                                     = db.BannedIp
	BoardPostCountRow                            = db.BoardPostCountRow
	CategoryCountRow                             = db.CategoryCountRow
	CountPermissionSectionsRow                   = db.CountPermissionSectionsRow
	GetAllForumCategoriesWithSubcategoryCountRow = db.GetAllForumCategoriesWithSubcategoryCountRow
	GetLinkerCategoryLinkCountsRow               = db.GetLinkerCategoryLinkCountsRow
	InsertBannedIpParams                         = db.InsertBannedIpParams
	InsertNotificationParams                     = db.InsertNotificationParams
	Language                                     = db.Language
	ListAnnouncementsWithNewsRow                 = db.ListAnnouncementsWithNewsRow
	ListAuditLogsParams                          = db.ListAuditLogsParams
	ListAuditLogsRow                             = db.ListAuditLogsRow
	MonthlyUsageRow                              = db.MonthlyUsageRow
	Notification                                 = db.Notification
	PendingEmail                                 = db.PendingEmail
	PermissionWithUser                           = db.PermissionWithUser
	RenamePermissionSectionParams                = db.RenamePermissionSectionParams
	User                                         = db.User
	UserMonthlyUsageRow                          = db.UserMonthlyUsageRow
	UserPostCountRow                             = db.UserPostCountRow
	Writingcategory                              = db.Writingcategory
)

func New(d db.DBTX) *Queries { return db.New(d) }
